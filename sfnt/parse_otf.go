// Package sfnt provides support for sfnt based font formats.
//
// This includes OpenType, TrueType, wOFF and EOT (though EOT is currently unimplemented).
//
// Usually you will want to Parse a font, make modifications, and then output the modified
// font. If you're really brave, you can build a New Font from scratch.
package sfnt

import (
	"encoding/binary"
	"fmt"
	"math"
)

type otfHeader struct {
	ScalerType    Tag
	NumTables     uint16
	SearchRange   uint16
	EntrySelector uint16
	RangeShift    uint16
}

const otfHeaderLength = 12
const directoryEntryLength = 16

func newOtfHeader(scalerType Tag, numTables uint16) *otfHeader {

	// http://www.opensource.apple.com/source/ICU/ICU-491.11.3/icuSources/layout/KernTable.cpp?txt
	entrySelector := uint16(math.Logb(float64(numTables)))
	searchRange := ((1 << entrySelector) * uint16(16))
	rangeShift := (numTables * uint16(16)) - searchRange

	return &otfHeader{
		ScalerType:    scalerType,
		NumTables:     numTables,
		EntrySelector: entrySelector,
		SearchRange:   searchRange,
		RangeShift:    rangeShift,
	}

}

func (header *otfHeader) checkSum() uint32 {
	return header.ScalerType.Number +
		(uint32(header.NumTables)<<16 | uint32(header.SearchRange)) +
		(uint32(header.EntrySelector)<<16 + uint32(header.RangeShift))
}

// An Entry in an OpenType table.
type directoryEntry struct {
	Tag      Tag
	CheckSum uint32
	Offset   uint32
	Length   uint32
}

func (entry *directoryEntry) checkSum() uint32 {
	return entry.Tag.Number + entry.CheckSum + entry.Offset + entry.Length
}

// ParseOTF reads an OpenTyp (.otf) or TrueType (.ttf) file and returns a Font.
// If parsing fails, then an error is returned and Font will be nil.
func parseOTF(file File) (*Font, error) {
	header := otfHeader{}
	err := binary.Read(file, binary.BigEndian, &header)
	if err != nil {
		return nil, err
	}

	font := &Font{
		file:       file,
		scalerType: header.ScalerType,
		tables:     make(map[Tag]tableSection, header.NumTables),
	}

	for i := uint16(0); i < header.NumTables; i++ {
		var entry directoryEntry
		err := binary.Read(file, binary.BigEndian, &entry)
		if err != nil {
			return nil, err
		}

		// TODO Check the checksum.

		if _, found := font.tables[entry.Tag]; found {
			return nil, fmt.Errorf("found multiple %q tables", entry.Tag)
		}

		font.tables[entry.Tag] = &tableSection{
			tag: entry.Tag,

			offset: entry.Offset,
			length: entry.Length,
		}
	}

	if _, ok := font.tables[TagHead]; !ok {
		return nil, ErrMissingHead
	}

	return font, nil
}
