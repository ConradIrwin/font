package sfnt

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
)

// MissingGlyph is the missing glyph value.
const MissingGlyph = 0

// TableCmap represents the character to glyph table used by 'cmap'.
//
// See https://www.microsoft.com/typography/otspec/cmap.htm
type TableCmap struct {
	bytes []byte

	header    cmapHeader
	Encodings []*EncodingSubtable
}

// EncodingFormatID represents the encoding format for entries in the 'cmap' table.
type EncodingFormatID uint16

// EncodingSubtable is one of the encoding Subtables in the 'cmap' table.
// Implementing the CharacterToGlyph interface.
type EncodingSubtable struct {
	Platform PlatformID
	Encoding PlatformEncodingID
	Format   EncodingFormatID

	subtable CharacterToGlyph
}

func (e *EncodingSubtable) String() string {
	return fmt.Sprintf("%s %s (format: %d)", e.Platform, e.Encoding.String(e.Platform), e.Format)
}

// Map returns the glyph for this character, or MissingGlyph if one is not found.
func (e *EncodingSubtable) Map(r rune) int {
	return e.subtable.Map(r)
}

// Characters returns all characters encoded by this subtable.
func (e *EncodingSubtable) Characters() []rune {
	return e.subtable.Characters()
}

// The CharacterToGlyph interface maps Characters to Glyphs.
type CharacterToGlyph interface {
	// Map returns the glyph for this character, or MissingGlyph if one is not found.
	Map(r rune) int

	// Characters returns all characters encoded by this subtable.
	Characters() []rune
}

// Bytes returns the bytes for this table. The TableCmap is read only, so
// the bytes will always be the same as what is read in.
func (t *TableCmap) Bytes() []byte {
	// TODO Write out the table ourselves (instead of stored bytes)
	return t.bytes
}

// cmapHeader is the beginning of on-disk format of the 'cmap' table.
// See https://www.microsoft.com/typography/otspec/cmap.htm
type cmapHeader struct {
	Version   uint16 // Table version number (0).
	NumTables uint16 // Number of encoding tables that follow.
}

type cmapEncodingRecord struct {
	PlatformID uint16 // Platform ID.
	EncodingID uint16 // Platform-specific encoding ID.
	Offset     uint32 // Byte offset from beginning of table to the subtable for this encoding.
}

type cmapUnsupportedFormat struct{}

func (*cmapUnsupportedFormat) Map(r rune) int {
	return MissingGlyph
}

func (*cmapUnsupportedFormat) Characters() []rune {
	return nil
}

// Format 4: Segment mapping to delta values
type cmapFormat4 struct {
	header      cmapFormat4Header
	reservedPad uint16 // Set to 0.

	startCode     []uint16 // Start character code for each segment.
	endCode       []uint16 // End character code for each segment, last=0xFFFF.
	idDelta       []int16  // Delta for all character codes in segment.
	idRangeOffset []uint16 // Offsets into glyphs or 0
	glyphs        []uint16 // Glyph index array (arbitrary length)
}

type cmapFormat4Header struct {
	Length        uint16 // This is the length in bytes of the subtable.
	Language      uint16 // Language
	SegCountX2    uint16 // 2 x segCount.
	SearchRange   uint16 // 2 x (2**floor(log2(segCount)))
	EntrySelector uint16 // log2(searchRange/2)
	RangeShift    uint16 // 2 x segCount - searchRange
}

func (f *cmapFormat4) Map(r rune) int {
	c := uint16(r)

	for i, end := range f.endCode {
		// Search for the first endCode that is greater than or equal to
		// the character code to be mapped
		if c > end {
			continue
		}

		if f.startCode[i] > c {
			return MissingGlyph
		}

		if f.idRangeOffset[i] == 0 {
			return (int(c) + int(f.idDelta[i])) % 65536
		}

		glyphIndex := int(c) - int(f.startCode[i]) - len(f.idRangeOffset) + i + int(f.idRangeOffset[i]/2)
		if glyphIndex < 0 || glyphIndex >= len(f.glyphs) || f.glyphs[glyphIndex] == MissingGlyph {
			return MissingGlyph
		}

		return (int(f.glyphs[glyphIndex]) + int(f.idDelta[i])) % 65536
	}

	return MissingGlyph
}

func (f *cmapFormat4) Characters() []rune {
	var runes []rune
	for i := range f.startCode {
		end := rune(f.endCode[i])
		if end == 0xFFFF {
			end = 0 // Avoid printing out 0xFFFF
		}
		for c := rune(f.startCode[i]); c <= end; c++ {
			runes = append(runes, c)
		}
	}
	return runes
}

func parseCmapFormat4(r io.Reader) (*cmapFormat4, error) {
	var subtable cmapFormat4
	if err := binary.Read(r, binary.BigEndian, &subtable.header); err != nil {
		return nil, fmt.Errorf("header: %s", err)
	}

	if subtable.header.SegCountX2%2 == 1 || subtable.header.SegCountX2 == 0 {
		return nil, fmt.Errorf("invalid SegCountX2: %d", subtable.header.SegCountX2)
	}

	segCount := subtable.header.SegCountX2 / 2
	subtable.endCode = make([]uint16, segCount)
	subtable.startCode = make([]uint16, segCount)
	subtable.idDelta = make([]int16, segCount)
	subtable.idRangeOffset = make([]uint16, segCount)

	if err := binary.Read(r, binary.BigEndian, subtable.endCode); err != nil {
		return nil, fmt.Errorf("endCode: %s", err)
	}

	if err := binary.Read(r, binary.BigEndian, &subtable.reservedPad); err != nil {
		return nil, fmt.Errorf("reservedPad: %s", err)
	}

	if err := binary.Read(r, binary.BigEndian, subtable.startCode); err != nil {
		return nil, fmt.Errorf("startCode: %s", err)
	}

	if err := binary.Read(r, binary.BigEndian, subtable.idDelta); err != nil {
		return nil, fmt.Errorf("idDelta: %s", err)
	}

	if err := binary.Read(r, binary.BigEndian, subtable.idRangeOffset); err != nil {
		return nil, fmt.Errorf("idRangeOffset: %s", err)
	}

	// The header is 14 bytes + 2 bytes for reservedPad + 4 x int16 arrays of segCount length
	remaining := int(subtable.header.Length) - 16 - int(segCount)*8
	if remaining < 0 || remaining%2 != 0 {
		return nil, fmt.Errorf("invalid length %d", subtable.header.Length)
	}

	subtable.glyphs = make([]uint16, remaining/2)
	if err := binary.Read(r, binary.BigEndian, subtable.glyphs); err != nil {
		return nil, fmt.Errorf("glyphs: %s", err)
	}

	for i, offset := range subtable.idRangeOffset {
		if offset%2 == 1 {
			return nil, fmt.Errorf("invalid idRangeOffset[%d]: %d (is odd)", i, offset)
		}
	}

	for i := 1; i < len(subtable.endCode); i++ {
		if subtable.endCode[i-1] >= subtable.endCode[i] {
			return nil, fmt.Errorf("endCode is expected to be sorted: %s", subtable.endCode)
		}
	}

	/*
		fmt.Printf("        start: %5d\n", subtable.startCode)
		fmt.Printf("          end: %5d\n", subtable.endCode)
		fmt.Printf("      idDelta: %5d\n", subtable.idDelta)
		fmt.Printf("idRangeOffset: %5d\n", subtable.idRangeOffset)
		fmt.Printf(" glyphs: %5d\n", subtable.glyphs)
	*/
	return &subtable, nil
}

//Format 6: Trimmed table mapping
type cmapFormat6 struct {
	header cmapFormat6Header

	glyphs []uint16 // Array of glyph index values for character codes in the range.
}

type cmapFormat6Header struct {
	Length   uint16 // This is the length in bytes of the subtable.
	Language uint16 // Language

	StartCode uint16 // First character code of subrange.
	NumChars  uint16 // Number of character codes in subrange.
}

func (f *cmapFormat6) Map(r rune) int {
	c := uint16(r)

	// TODO The rune should be encoded to the correct Platform/
	//c, _ := charmap.Macintosh.EncodeRune(r)

	if c < f.header.StartCode || (c-f.header.StartCode) >= f.header.NumChars {
		return MissingGlyph
	}

	return int(f.glyphs[c-f.header.StartCode])
}

func (f *cmapFormat6) Characters() []rune {
	var runes []rune
	r := rune(f.header.StartCode)
	for _, g := range f.glyphs {
		if g != MissingGlyph {
			runes = append(runes, r)
		}
		r++
	}
	return runes
}

func parseCmapFormat6(r io.Reader) (*cmapFormat6, error) {
	var subtable cmapFormat6
	if err := binary.Read(r, binary.BigEndian, &subtable.header); err != nil {
		return nil, fmt.Errorf("header: %s", err)
	}

	if length := (10 + subtable.header.NumChars*2); subtable.header.Length != length {
		return nil, fmt.Errorf("length does not match expected length %d != %d", subtable.header.Length, length)
	}

	subtable.glyphs = make([]uint16, subtable.header.NumChars)
	if err := binary.Read(r, binary.BigEndian, subtable.glyphs); err != nil {
		return nil, fmt.Errorf("glyphs: %s", err)
	}

	return &subtable, nil
}

// Format 10: Trimmed array
type cmapFormat10 struct {
	header cmapFormat10Header

	glyphs []uint16 // Array of glyph indices for the character codes covered.
}

type cmapFormat10Header struct {
	Reserved uint16 // Reserved; set to 0
	Length   uint32 // Byte length of this subtable (including the header)

	Language  uint32
	StartCode uint32 // First character code covered
	NumChars  uint32 // Number of character codes covered
}

func (f *cmapFormat10) Map(r rune) int {
	c := uint32(r)

	if c < f.header.StartCode || (c-f.header.StartCode) >= f.header.NumChars {
		return MissingGlyph
	}

	return int(f.glyphs[c-f.header.StartCode])
}

func (f *cmapFormat10) Characters() []rune {
	var runes []rune
	r := rune(f.header.StartCode)
	for _, g := range f.glyphs {
		if g != MissingGlyph {
			runes = append(runes, r)
		}
		r++
	}
	return runes
}

func parseCmapFormat10(r io.Reader) (*cmapFormat10, error) {
	var subtable cmapFormat10
	if err := binary.Read(r, binary.BigEndian, &subtable.header); err != nil {
		return nil, fmt.Errorf("header: %s", err)
	}

	if length := (20 + subtable.header.NumChars*4); subtable.header.Length != length {
		return nil, fmt.Errorf("length does not match expected length %d != %d", subtable.header.Length, length)
	}

	subtable.glyphs = make([]uint16, subtable.header.NumChars)
	if err := binary.Read(r, binary.BigEndian, subtable.glyphs); err != nil {
		return nil, fmt.Errorf("glyphs: %s", err)
	}

	return &subtable, nil
}

// Format 12: Segmented coverage
type cmapFormat12 struct {
	header cmapFormat12Header

	glyphs []sequentialMapGroup // Array of SequentialMapGroup records.
}

type cmapFormat12Header struct {
	Reserved uint16 // Reserved; set to 0
	Length   uint32 // Byte length of this subtable (including the header)

	Language  uint32
	NumGroups uint32 // Number of groupings which follow
}

type sequentialMapGroup struct {
	StartCode    uint32 // First character code in this group
	EndCode      uint32 // Last character code in this group
	StartGlyphID uint32 // Glyph index corresponding to the starting character code
}

func (f *cmapFormat12) Map(r rune) int {
	c := uint32(r)
	for _, g := range f.glyphs {
		if c < g.StartCode {
			break
		}
		if g.StartCode <= c && c <= g.EndCode {
			return int(g.StartGlyphID + (c - g.StartCode))
		}
	}
	return MissingGlyph
}

func (f *cmapFormat12) Characters() []rune {
	var runes []rune
	for _, g := range f.glyphs {
		end := rune(g.EndCode)
		for r := rune(g.StartCode); r <= end; r++ {
			runes = append(runes, r)
		}
	}
	return runes
}

func parseCmapFormat12(r io.Reader) (*cmapFormat12, error) {
	var subtable cmapFormat12
	if err := binary.Read(r, binary.BigEndian, &subtable.header); err != nil {
		return nil, fmt.Errorf("header: %s", err)
	}

	if length := (16 + subtable.header.NumGroups*12); subtable.header.Length != length {
		return nil, fmt.Errorf("length does not match expected length %d != %d", subtable.header.Length, length)
	}

	subtable.glyphs = make([]sequentialMapGroup, subtable.header.NumGroups)
	if err := binary.Read(r, binary.BigEndian, subtable.glyphs); err != nil {
		return nil, fmt.Errorf("glyphs: %s", err)
	}

	for i := 1; i < len(subtable.glyphs); i++ {
		if subtable.glyphs[i-1].StartCode >= subtable.glyphs[i].StartCode {
			return nil, fmt.Errorf("glyphs is expected to be sorted by StartCode")
		}
		if subtable.glyphs[i-1].EndCode >= subtable.glyphs[i].StartCode {
			return nil, fmt.Errorf("glyphs groups overlap %v vs %v", subtable.glyphs[i-1], subtable.glyphs[i])
		}
	}

	return &subtable, nil
}

// validCmapFormat returns true iff this cmap format is valid based on the Platform and Encoding ID.
func validCmapFormat(t *EncodingSubtable) error {
	var allowed []EncodingFormatID

	if t.Platform == PlatformUnicode {
		switch t.Encoding {
		case 3:
			allowed = []EncodingFormatID{0, 4, 6}
		case 4:
			allowed = []EncodingFormatID{0, 4, 6, 10, 12}
		case 5:
			allowed = []EncodingFormatID{14}
		case 6:
			allowed = []EncodingFormatID{0, 4, 6, 10, 12, 13}
		}
	}

	if len(allowed) > 0 {
		for _, format := range allowed {
			if t.Format == format {
				return nil
			}
		}

		return fmt.Errorf("%s %s requires format %v not %d", t.Platform, t.Encoding.String(t.Platform), allowed, t.Format)
	}

	return nil
}

// parseTableCmap parses a Character To Glyph Index Mapping Table used by 'cmap'.
func parseTableCmap(r io.Reader) (Table, error) {
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	t := &TableCmap{
		bytes: buf,
	}

	r = bytes.NewReader(t.bytes)
	if err := binary.Read(r, binary.BigEndian, &t.header); err != nil {
		return nil, fmt.Errorf("reading cmap header: %s", err)
	}

	if t.header.Version != 0 {
		return nil, fmt.Errorf("unsupported cmap version: %d", t.header.Version)
	}

	for i := 0; i < int(t.header.NumTables); i++ {
		var record cmapEncodingRecord
		if err := binary.Read(r, binary.BigEndian, &record); err != nil {
			return nil, fmt.Errorf("reading cmapEncodingRecord[%d]: %s", i, err)
		}

		encoding := &EncodingSubtable{
			Platform: PlatformID(record.PlatformID),
			Encoding: PlatformEncodingID(record.EncodingID),
		}

		r := bytes.NewReader(buf[record.Offset:])

		var format uint16
		if err := binary.Read(r, binary.BigEndian, &format); err != nil {
			return nil, fmt.Errorf("reading cmapEncodingTable[%d] at offset 0x%x: %s", i, record.Offset, err)
		}
		encoding.Format = EncodingFormatID(format)

		if err := validCmapFormat(encoding); err != nil {
			return nil, fmt.Errorf("invalid format for cmapEncodingTable[%d] at offset 0x%x: %s", i, record.Offset, err)
		}

		var err error

		switch format {
		case 4:
			encoding.subtable, err = parseCmapFormat4(r)
		case 6:
			encoding.subtable, err = parseCmapFormat6(r)
		case 10:
			encoding.subtable, err = parseCmapFormat10(r)
		case 12:
			encoding.subtable, err = parseCmapFormat12(r)

		default:
			encoding.subtable = &cmapUnsupportedFormat{}
		}

		if err != nil {
			return nil, fmt.Errorf("reading cmapEncodingTable[%d] format: %d at offset 0x%x: %s", i, format, record.Offset, err)
		}

		t.Encodings = append(t.Encodings, encoding)
	}

	return t, nil
}
