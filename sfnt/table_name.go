package sfnt

import (
	"bytes"
	"encoding/binary"
	"io"
	"io/ioutil"
	"strconv"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// TableName represents the OpenType 'name' table. This contains
// human-readable meta-data about the font, for example the Author
// and Copyright.
// https://developer.apple.com/fonts/TrueType-Reference-Manual/RM06/Chap6name.html
type TableName struct {
	bytes         []byte
	bytesAreStale bool
	entries       []*NameEntry
}

type nameHeader struct {
	Format       uint16
	Count        uint16
	StringOffset uint16
}

// PlatformID represents the platform id for entries in the name table.
type PlatformID uint16

var (
	PlatformUnicode   = PlatformID(0)
	PlatformMac       = PlatformID(1)
	PlatformMicrosoft = PlatformID(3)
)

// String returns an idenfying string for each platform or "Platform X" for unknown values.
func (p PlatformID) String() string {
	switch p {
	case PlatformUnicode:
		return "Unicode"
	case PlatformMac:
		return "Mac"
	case PlatformMicrosoft:
		return "Microsoft"
	default:
		return "Platform " + strconv.Itoa(int(p))
	}
}

// PlatformEncodingID represents the platform specific id for entries in the name table.
// the three most common values are provided as constants.
type PlatformEncodingID uint16

var (
	PlatformEncodingMacRoman         = PlatformEncodingID(0)
	PlatformEncodingUnicodeDefault   = PlatformEncodingID(0)
	PlatformEncodingMicrosoftUnicode = PlatformEncodingID(1)
)

// PlatformLanguageID represents the language used by an entry in the name table,
// the three most common values are provided as constants.
type PlatformLanguageID uint16

var (
	PlatformLanguageMacEnglish       = PlatformLanguageID(0)
	PlatformLanguageUnicodeDefault   = PlatformLanguageID(0)
	PlatformLanguageMicrosoftEnglish = PlatformLanguageID(0x0409)
)

// NameID is the ID for entries in the font table.
type NameID uint16

var (
	NameCopyrightNotice        = NameID(0)
	NameFontFamily             = NameID(1)
	NameFontSubfamily          = NameID(2)
	NameUniqueIdentifier       = NameID(3)
	NameFull                   = NameID(4)
	NameVersion                = NameID(5)
	NamePostscript             = NameID(6)
	NameTrademark              = NameID(7)
	NameManufacturer           = NameID(8)
	NameDesigner               = NameID(9)
	NameDescription            = NameID(10)
	NameVendorURL              = NameID(11)
	NameDesignerURL            = NameID(12)
	NameLicenseDescription     = NameID(13)
	_NameReserved              = NameID(15)
	NameLicenseURL             = NameID(14)
	NamePreferredFamily        = NameID(16)
	NamePreferredSubfamily     = NameID(17)
	NameCompatibleFull         = NameID(18)
	NameSampleText             = NameID(19)
	NamePostscriptCID          = NameID(20)
	NameWWSFamily              = NameID(21)
	NameWWSSubfamily           = NameID(22)
	NameLightBackgroundPalette = NameID(23)
	NameDarkBackgroundPalette  = NameID(24)
)

// String returns an identifying
func (nameId NameID) String() string {
	switch nameId {
	case NameCopyrightNotice:
		return "Copyright Notice"
	case NameFontFamily:
		return "Font Family"
	case NameFontSubfamily:
		return "Font Subfamily"
	case NameUniqueIdentifier:
		return "Unique Identifier"
	case NameFull:
		return "Full Name"
	case NameVersion:
		return "Version"
	case NamePostscript:
		return "PostScript Name"
	case NameTrademark:
		return "Trademark Notice"
	case NameManufacturer:
		return "Manufacturer"
	case NameDesigner:
		return "Designer"
	case NameDescription:
		return "Description"
	case NameVendorURL:
		return "Vendor URL"
	case NameDesignerURL:
		return "Designer URL"
	case NameLicenseDescription:
		return "License Description"
	case NameLicenseURL:
		return "License URL"
	case NamePreferredFamily:
		return "Preferred Family"
	case NamePreferredSubfamily:
		return "Preferred Subfamily"
	case NameCompatibleFull:
		return "Compatible Full"
	case NameSampleText:
		return "Sample Text"
	case NamePostscriptCID:
		return "PostScript CID"
	case NameWWSFamily:
		return "WWS Family"
	case NameWWSSubfamily:
		return "WWS Subfamily"
	case NameLightBackgroundPalette:
		return "Light Background Palette"
	case NameDarkBackgroundPalette:
		return "Dark Background Palette"
	default:
		return "Name " + strconv.Itoa(int(nameId))
	}

}

type nameRecord struct {
	PlatformID PlatformID
	EncodingID PlatformEncodingID
	LanguageID PlatformLanguageID
	NameID     NameID
	Length     uint16
	Offset     uint16
}

type NameEntry struct {
	PlatformID PlatformID
	EncodingID PlatformEncodingID
	LanguageID PlatformLanguageID
	NameID     NameID
	Value      []byte
}

// String is a best-effort attempt to get a UTF-8 encoded version of
// Value. Only MicrosoftUnicode (3,1 ,X), MacRomain (1,0,X) and Unicode platform
// strings are supported.
func (nameEntry *NameEntry) String() string {

	if nameEntry.PlatformID == PlatformUnicode || (nameEntry.PlatformID == PlatformMicrosoft &&
		nameEntry.EncodingID == PlatformEncodingMicrosoftUnicode) {

		decoder := unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM).NewDecoder()

		outstr, _, err := transform.String(decoder, string(nameEntry.Value))

		if err == nil {
			return outstr
		}
	}

	if nameEntry.PlatformID == PlatformMac &&
		nameEntry.EncodingID == PlatformEncodingMacRoman {

		decoder := charmap.Macintosh.NewDecoder()

		outstr, _, err := transform.String(decoder, string(nameEntry.Value))

		if err == nil {
			return outstr
		}
	}

	return string(nameEntry.Value)
}

func (nameEntry *NameEntry) Label() string {
	return nameEntry.NameID.String()
}

func (nameEntry *NameEntry) Platform() string {
	return nameEntry.PlatformID.String()
}

func parseTableName(buffer io.Reader) (*TableName, error) {
	bytes, err := ioutil.ReadAll(buffer)
	if err != nil {
		return nil, err
	}
	return &TableName{bytes: bytes}, nil
}

// NewTableName returns an empty NAME table.
func NewTableName() *TableName {
	return &TableName{bytes: []byte{}, bytesAreStale: true, entries: []*NameEntry{}}
}

// AddMicrosoftEnglishEntry adds an entry to the name table for the 'Microsoft' platform,
// with Unicode Encoding (UCS-2) and the language set to English. It returns an error
// if the string cannot be represented in UCS-2.
func (table *TableName) AddMicrosoftEnglishEntry(nameId NameID, value string) error {
	encoder := unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM).NewEncoder()
	outstr, _, err := transform.String(encoder, value)
	if err != nil {
		return err
	}

	table.Add(&NameEntry{
		PlatformID: PlatformMicrosoft,
		EncodingID: PlatformEncodingMicrosoftUnicode,
		LanguageID: PlatformLanguageMicrosoftEnglish,
		NameID:     nameId,
		Value:      []byte(outstr),
	})

	return nil
}

// AddMacEnglishEntry adds an entry to the name table for the 'Mac' platform,
// with Default Encoding (Mac Roman) and the Language set to English. It returns
// an error if the value cannot be represented in Mac Roman.
func (table *TableName) AddMacEnglishEntry(nameId NameID, value string) error {
	encoder := unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM).NewEncoder()
	outstr, _, err := transform.String(encoder, value)
	if err != nil {
		return err
	}

	table.Add(&NameEntry{
		PlatformID: PlatformMac,
		EncodingID: PlatformEncodingMacRoman,
		LanguageID: PlatformLanguageMacEnglish,
		NameID:     nameId,
		Value:      []byte(outstr),
	})

	return nil
}

// AddUnicodeEntry adds an entry to the name table for the 'Unicode' platform,
// with Default Encoding (UTF-16). It returns an error if the value cannot be
// represented in UTF-16.
func (table *TableName) AddUnicodeEntry(nameId NameID, value string) error {
	encoder := unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM).NewEncoder()
	outstr, _, err := transform.String(encoder, value)
	if err != nil {
		return err
	}

	table.Add(&NameEntry{
		PlatformID: PlatformUnicode,
		EncodingID: PlatformEncodingUnicodeDefault,
		LanguageID: PlatformLanguageUnicodeDefault,
		NameID:     nameId,
		Value:      []byte(outstr),
	})

	return nil
}

// Add an entry to the table. This is a relatively low-level method, most of what you need can be
// accomplished using AddUnicodeEntry,AddMacEnglishEntry, and AddMicrosoftEnglishEntry.
func (table *TableName) Add(entry *NameEntry) {
	table.bytesAreStale = true
	table.List()
	table.entries = append(table.entries, entry)
}

// Bytes returns the representation of this table to be stored in a font.
func (table *TableName) Bytes() []byte {
	if !table.bytesAreStale {
		return table.bytes
	}

	buf := &bytes.Buffer{}

	header := nameHeader{0, uint16(len(table.entries)), uint16(binary.Size(nameHeader{}) + len(table.entries)*binary.Size(nameRecord{}))}

	binary.Write(buf, binary.BigEndian, header)

	offset := 0
	for _, entry := range table.entries {
		length := len(entry.Value)
		binary.Write(buf, binary.BigEndian, &nameRecord{
			PlatformID: entry.PlatformID,
			EncodingID: entry.EncodingID,
			LanguageID: entry.LanguageID,
			NameID:     entry.NameID,
			Length:     uint16(length),
			Offset:     uint16(offset),
		})
		offset += length
	}

	for _, entry := range table.entries {
		buf.Write(entry.Value)
	}

	table.bytes = buf.Bytes()
	table.bytesAreStale = false
	return table.bytes
}

// List returns a list of all the strings defined in this table.
func (table *TableName) List() []*NameEntry {
	if table.entries != nil {
		return table.entries
	}
	reader := bytes.NewBuffer(table.bytes)

	header := nameHeader{}
	err := binary.Read(reader, binary.BigEndian, &header)
	if err != nil {
		panic(err)
	}

	results := make([]*NameEntry, 0, header.Count)

	for i := uint16(0); i < header.Count; i++ {
		record := nameRecord{}
		err := binary.Read(reader, binary.BigEndian, &record)
		if err != nil {
			panic(err)
		}

		start := header.StringOffset + record.Offset
		end := start + record.Length

		results = append(results, &NameEntry{record.PlatformID, record.EncodingID, record.LanguageID, record.NameID, table.bytes[start:end]})
	}
	table.entries = results

	return results
}
