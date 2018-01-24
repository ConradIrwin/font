package sfnt

import (
	"encoding/binary"
	"errors"
	"sort"
)

type fixed struct {
	Major int16
	Minor uint16
}

type longdatetime struct {
	SecondsSince1904 uint64
}

func (u *unparsedTable) Bytes() []byte {
	return u.bytes
}

// ErrMissingHead is returned by ParseOTF when the font has no head section.
var ErrMissingHead = errors.New("missing head table in font")

// ErrInvalidChecksum is returned by ParseOTF if the font's checksum is wrong
var ErrInvalidChecksum = errors.New("invalid checksum")

// ErrUnsupportedFormat is returned from Parse if parsing failed
var ErrUnsupportedFormat = errors.New("unsupported font format")

// Font represents a SFNT font, which is the underlying representation found
// in .otf and .ttf files (and .woff and .eot files)
// SFNT is a container format, which contains a number of tables identified by
// Tags. Depending on the type of glyphs embedded in the file which tables will
// exist. In particular, there's a big different between TrueType glyphs (usually .ttf)
//and CFF/PostScript Type 2 glyphs (usually .otf)
type Font struct {
	scalerType Tag
	tables     map[Tag]Table
}

type tagList []Tag

func (tl tagList) Len() int           { return len(tl) }
func (tl tagList) Swap(i, j int)      { tl[i], tl[j] = tl[j], tl[i] }
func (tl tagList) Less(i, j int) bool { return tl[i].Number < tl[j].Number }

// Tags is the list of tags that are defined in this font, sorted by numeric value.
func (font *Font) Tags() []Tag {

	tags := make(tagList, 0, len(font.tables))

	for t := range font.tables {
		tags = append(tags, t)
	}

	sort.Sort(tags)

	return tags
}

// HasTable returns true if this font has an entry for the given table.
func (font *Font) HasTable(tag Tag) bool {
	_, ok := font.tables[tag]
	return ok
}

// AddTable adds a table to the font. If a table with the
// given tag is already present, it will be overwritten.
func (font *Font) AddTable(tag Tag, table Table) {
	font.tables[tag] = table
}

// RemoveTable removes a table from the font. If the table
// doesn't exist, this method will do nothing.
func (font *Font) RemoveTable(tag Tag) {
	delete(font.tables, tag)
}

// Type represents the kind of glyphs in this font.
// It is one of TypeTrueType, TypeTrueTypeApple, TypePostScript1, TypeOpenType
func (font *Font) Type() Tag {
	return font.scalerType
}

// String provides a debugging representation of a font.
func (font *Font) String() string {
	str := "Parsed font with scalerType=" + font.scalerType.hex()

	if font.scalerType != TypeTrueType {
		str += " (" + font.scalerType.String() + ")"
	}

	for _, t := range font.Tags() {
		str += "\n" + t.String()
	}

	return str
}

// HeadTable returns the table corresponding to the 'head' tag.
func (font *Font) HeadTable() (*TableHead, bool) {
	t, found := font.tables[TagHead].(*TableHead)
	return t, found
}

// NameTable returns the table corresponding to the 'name' tag.
func (font *Font) NameTable() (*TableName, bool) {
	t, found := font.tables[TagName].(*TableName)
	return t, found
}

func (font *Font) HheaTable() (*TableHhea, bool) {
	t, found := font.tables[TagHhea].(*TableHhea)
	return t, found
}

func (font *Font) OS2Table() (*TableOS2, bool) {
	t, found := font.tables[TagOS2].(*TableOS2)
	return t, found
}

// GposTable returns the Glyph Positioning table identified with the 'GPOS' tag.
func (font *Font) GposTable() (*TableLayout, bool) {
	t, found := font.tables[TagGpos].(*TableLayout)
	return t, found
}

// GsubTable returns the Glyph Substitution table identified with the 'GSUB' tag.
func (font *Font) GsubTable() (*TableLayout, bool) {
	t, found := font.tables[TagGsub].(*TableLayout)
	return t, found
}

func (font *Font) Table(tag Tag) (Table, bool) {
	t, found := font.tables[tag]
	return t, found
}

func (font *Font) checkSum() uint32 {
	var total uint32

	for _, table := range font.tables {
		total += checkSum(table.Bytes())
	}

	return total
}

// New returns an empty Font. It has only an empty 'head' table.
func New(scalerType Tag) *Font {
	font := &Font{
		scalerType,
		make(map[Tag]Table),
	}
	font.AddTable(TagHead, &TableHead{})
	return font
}

// File is a combination of io.Reader, io.Seeker and io.ReaderAt.
// This interface is satisfied by most things that you'd want
// to parse, for example os.File, or io.SectionReader.
type File interface {
	Read([]byte) (int, error)
	ReadAt([]byte, int64) (int, error)
	Seek(int64, int) (int64, error)
}

// Parse parses an OpenType, TrueType or wOFF File and returns a Font.
// If parsing fails, an error is returned and *Font will be nil.
func Parse(file File) (*Font, error) {

	magic := Tag{}

	err := binary.Read(file, binary.BigEndian, &magic)

	if err != nil {
		return nil, err
	}

	file.Seek(0, 0)

	if magic == SignatureWoff {
		return parseWoff(file)
	} else if magic == TypeTrueType || magic == TypeOpenType || magic == TypePostScript1 || magic == TypeAppleTrueType {
		return parseOTF(file)
	} else {
		return nil, ErrUnsupportedFormat
	}
}
