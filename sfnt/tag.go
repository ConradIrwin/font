package sfnt

import "encoding/hex"

var (
	// TagHead represents the 'head' table, which contains the font header
	TagHead = NamedTag("head")
	// TagMaxp represents the 'maxp' table, which contains the maximum profile
	TagMaxp = NamedTag("maxp")
	// TagHmtx represents the 'hmtx' table, which contains the horizontal metrics
	TagHmtx = NamedTag("hmtx")
	// TagHhea represents the 'hhea' table, which contains the horizonal header
	TagHhea = NamedTag("hhea")
	// TagOS2 represents the 'OS/2' table, which contains windows-specific metadata
	TagOS2 = NamedTag("OS/2")
	// TagName represents the 'name' table, which contains font name information
	TagName = NamedTag("name")
	// TagGpos represents the 'GPOS' table, which contains Glyph Positioning features
	TagGpos = NamedTag("GPOS")
	// TagGsub represents the 'GSUB' table, which contains Glyph Substitution features
	TagGsub = NamedTag("GSUB")

	// TypeTrueType is the first four bytes of an OpenType file containing a TrueType font
	TypeTrueType = Tag{0x00010000}
	// TypeAppleTrueType is the first four bytes of an OpenType file containing a TrueType font
	// (specifically one designed for Apple products, it's recommended to use TypeTrueType instead)
	TypeAppleTrueType = NamedTag("true")
	// TypePostScript1 is the first four bytes of an OpenType file containing a PostScript Type 1 font
	TypePostScript1 = NamedTag("typ1")
	// TypeOpenType is the first four bytes of an OpenType file containing a PostScript Type 2 font
	// as specified by OpenType
	TypeOpenType = NamedTag("OTTO")
	// TypeTrueTypeCollection is a font file that contains multiple fonts
	TypeTrueTypeCollection = NamedTag("ttcf")

	// SignatureWoff if the magic number at the start of a wOFF file.
	SignatureWoff = NamedTag("wOFF")
)

// Tag represents an open-type table name.
// These are technically uint32's, but are usually
// displayed in ASCII as they are all acronyms.
// see https://developer.apple.com/fonts/TrueType-Reference-Manual/RM06/Chap6.html#Overview
type Tag struct {
	Number uint32
}

// NamedTag gives you the Tag corresponding to the acronym.
// This function will panic if the string passed in is not 4 bytes long.
func NamedTag(str string) Tag {
	bytes := []byte(str)

	if len(bytes) != 4 {
		panic("invalid tag")
	}

	return Tag{uint32(bytes[0])<<24 |
		uint32(bytes[1])<<16 |
		uint32(bytes[2])<<8 |
		uint32(bytes[3])}

}

// String returns the ASCII representation of the tag.
func (tag Tag) String() string {
	return string(tag.bytes())
}

func (tag Tag) bytes() []byte {
	return []byte{
		byte(tag.Number >> 24 & 0xFF),
		byte(tag.Number >> 16 & 0xFF),
		byte(tag.Number >> 8 & 0xFF),
		byte(tag.Number & 0xFF),
	}
}

func (tag Tag) hex() string {
	return "0x" + hex.EncodeToString(tag.bytes())
}
