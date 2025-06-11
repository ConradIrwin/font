package sfnt

import (
	"encoding/binary"
	"io"
)

type ttcfHeaderV1 struct {
	ScalerType   Tag
	MajorVersion uint16
	MinorVersion uint16
	NumFonts     uint32
}

// ParseTTCF reads a TrueType Collection and returns an array of fonts
func parseTTCF(file File) ([]*Font, error) {
	header := ttcfHeaderV1{}
	err := binary.Read(file, binary.BigEndian, &header)
	if err != nil {
		return nil, err
	}

	fonts := []*Font{}

	for i := uint32(0); i < header.NumFonts; i++ {
		offset := uint32(0)
		err := binary.Read(file, binary.BigEndian, &offset)
		if err != nil {
			return nil, err
		}

		font, err := parse(io.NewSectionReader(file, int64(offset), 1<<63-1), file)
		if err != nil {
			return nil, err
		}
		fonts = append(fonts, font)
	}

	return fonts, nil
}

// ParseCollection parses a TrueType Collection (.ttc) file and returns an array of fonts.
// It also accepts a font file that Parse accepts and returns an array of fonts with a length of 1.
func ParseCollection(file File) ([]*Font, error) {
	magic := Tag{}
	err := binary.Read(file, binary.BigEndian, &magic)
	if err != nil {
		return nil, err
	}
	file.Seek(0, io.SeekStart)

	if magic != TypeTrueTypeCollection {
		font, err := Parse(file)
		if err != nil {
			return nil, err
		}
		return []*Font{font}, nil
	}

	return parseTTCF(file)
}
