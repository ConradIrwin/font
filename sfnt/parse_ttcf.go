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

// ParseTTCF reads a TrueTypeFontCollection and returns an array of fonts
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
		font, err := Parse(io.NewSectionReader(file, int64(offset), 102410241024))

		if err != nil {
			return nil, err
		}

		fonts = append(fonts, font)
	}

	return fonts, nil

}

func ParseCollection(file File) ([]*Font, error) {

	magic := Tag{}

	err := binary.Read(file, binary.BigEndian, &magic)

	if err != nil {
		return nil, err
	}
	file.Seek(0, 0)

	if magic != TypeTrueTypeCollection {
		font, err := Parse(file)
		if err != nil {
			return nil, err
		}
		return []*Font{font}, nil
	}

	return parseTTCF(file)
}
