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

type cmapFormat4 struct {
	header struct {
		Length        uint16 // This is the length in bytes of the subtable.
		Language      uint16 // Language
		SegCountX2    uint16 // 2 x segCount.
		SearchRange   uint16 // 2 x (2**floor(log2(segCount)))
		EntrySelector uint16 // log2(searchRange/2)
		RangeShift    uint16 // 2 x segCount - searchRange
	}

	reservedPad uint16 // Set to 0.

	startCode     []uint16 // Start character code for each segment.
	endCode       []uint16 // End character code for each segment, last=0xFFFF.
	idDelta       []int16  // Delta for all character codes in segment.
	idRangeOffset []uint16 // Offsets into glyphIDArray or 0
	glyphIDArray  []uint16 // Glyph index array (arbitrary length)
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
		if glyphIndex < 0 || glyphIndex >= len(f.glyphIDArray) || f.glyphIDArray[glyphIndex] == MissingGlyph {
			return MissingGlyph
		}

		return (int(f.glyphIDArray[glyphIndex]) + int(f.idDelta[i])) % 65536
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
	if err := binary.Read(r, binary.BigEndian, subtable.endCode); err != nil {
		return nil, fmt.Errorf("endCode: %s", err)
	}

	if err := binary.Read(r, binary.BigEndian, &subtable.reservedPad); err != nil {
		return nil, fmt.Errorf("reservedPad: %s", err)
	}

	subtable.startCode = make([]uint16, segCount)
	if err := binary.Read(r, binary.BigEndian, subtable.startCode); err != nil {
		return nil, fmt.Errorf("startCode: %s", err)
	}

	subtable.idDelta = make([]int16, segCount)
	if err := binary.Read(r, binary.BigEndian, subtable.idDelta); err != nil {
		return nil, fmt.Errorf("idDelta: %s", err)
	}

	subtable.idRangeOffset = make([]uint16, segCount)
	if err := binary.Read(r, binary.BigEndian, subtable.idRangeOffset); err != nil {
		return nil, fmt.Errorf("idRangeOffset: %s", err)
	}

	// The header is 14 bytes + 2 bytes for reservedPad + 4 x int16 arrays of segCount length
	remaining := int(subtable.header.Length) - 16 - int(segCount)*8
	if remaining < 0 || remaining%2 != 0 {
		return nil, fmt.Errorf("invalid length %d", subtable.header.Length)
	}

	subtable.glyphIDArray = make([]uint16, remaining/2)
	if err := binary.Read(r, binary.BigEndian, subtable.glyphIDArray); err != nil {
		return nil, fmt.Errorf("glyphIDArray: %s", err)
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
		fmt.Printf(" glyphIDArray: %5d\n", subtable.glyphIDArray)
	*/
	return &subtable, nil
}

// validCmapFormat returns true iff this cmap format is valid for this Platform and Encoding ID.
func validCmapFormat(platform PlatformID, encoding PlatformEncodingID, format int) error {
	if platform == PlatformUnicode {
		if encoding == 3 && !(format == 0 || format == 4 || format == 6) {
			return fmt.Errorf("%s %s requires format [0, 4 or 6] not %d", platform, encoding.String(platform), format)
		}
		if encoding == 4 && !(format == 0 || format == 4 || format == 6 || format == 10 || format == 12) {
			return fmt.Errorf("%s %s requires format [0, 4, 6, 10 or 12] not %d", platform, encoding.String(platform), format)
		}
		if encoding == 5 && !(format == 14) {
			return fmt.Errorf("%s %s requires format [14] not %d", platform, encoding.String(platform), format)
		}
		if encoding == 6 && !(format == 0 || format == 4 || format == 6 || format == 10 || format == 12 || format == 13) {
			return fmt.Errorf("%s %s requires format [0, 4, 6, 10, 12, 13] not %d", platform, encoding.String(platform), format)
		}
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

		r := bytes.NewReader(buf[record.Offset:])

		var format uint16
		if err := binary.Read(r, binary.BigEndian, &format); err != nil {
			return nil, fmt.Errorf("reading cmapEncodingTable[%d] at offset 0x%x: %s", i, record.Offset, err)
		}

		if err := validCmapFormat(record.PlatformID, record.EncodingID, format); err != nil {
			return nil, fmt.Errorf("invalid format for cmapEncodingTable[%d] at offset 0x%x: %s", i, record.Offset, err)
		}

		var subtable CharacterToGlyph

		if format == 4 {
			var err error
			if subtable, err = parseCmapFormat4(r); err != nil {
				return nil, fmt.Errorf("reading cmapEncodingTable[%d] at offset 0x%x: %s", i, record.Offset, err)
			}
		} else {
			subtable = &cmapUnsupportedFormat{}
		}

		t.Encodings = append(t.Encodings, &EncodingSubtable{
			Platform: PlatformID(record.PlatformID),
			Encoding: PlatformEncodingID(record.EncodingID),
			Format:   EncodingFormatID(format),

			subtable: subtable,
		})
	}

	return t, nil
}
