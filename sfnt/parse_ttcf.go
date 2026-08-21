package sfnt

import (
	"encoding/binary"
	"fmt"
	"io"
)

type ttcfHeaderV1 struct {
	ScalerType   Tag
	MajorVersion uint16
	MinorVersion uint16
	NumFonts     uint32
}

// parseTTCFHeader parses ttcf header of a TrueType Collection file
func parseTTCFHeader(file File) (ttcfHeaderV1, error) {
	header := ttcfHeaderV1{}
	err := binary.Read(file, binary.BigEndian, &header)

	return header, err
}

// parseTTCF reads a TrueType Collection and returns an array of fonts
func parseTTCF(file File) ([]*Font, error) {
	header, err := parseTTCFHeader(file)
	if err != nil {
		return nil, err
	}

	// header.NumFonts is an untrusted 32-bit value read straight from the file.
	// Grow the slice as fonts are actually read instead of pre-allocating from
	// the header, so a corrupt count can't force a huge allocation before any
	// offset has been validated.
	var fonts []*Font

	for i := uint32(0); i < header.NumFonts; i++ {
		var offset uint32
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
	magic, err := ReadTag(file)
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

// StrictParseCollection parses a TrueType Collection file and returns
// an array of fonts. Each table from each font will be fully parsed
// and an error is returned if any fail.
func StrictParseCollection(file File) ([]*Font, error) {
	fonts, err := ParseCollection(file)
	if err != nil {
		return nil, err
	}
	for i, font := range fonts {
		for _, tag := range font.Tags() {
			if _, err := font.Table(tag); err != nil {
				return nil, fmt.Errorf("font[%d]: failed to parse %q: %w", i, tag, err)
			}
		}
	}
	return fonts, nil
}

// ParseCollectionIndex parses a single font from a TrueType Collection (.ttc) file with
// font index starting from 0. An error is returned if a file is not a collection.
func ParseCollectionIndex(file File, index uint32) (*Font, error) {
	magic, err := ReadTag(file)
	if err != nil {
		return nil, err
	}
	if magic != TypeTrueTypeCollection {
		return nil, fmt.Errorf("expected \"ttcf\" for head bytes, got \"%s\" instead", magic)
	}
	file.Seek(0, io.SeekStart)

	header, err := parseTTCFHeader(file)
	if err != nil {
		return nil, err
	}

	if index > header.NumFonts-1 {
		return nil, fmt.Errorf("index can't be larger than %d (got %d)", header.NumFonts-1, index)
	}

	file.Seek(int64(index*4), io.SeekCurrent)

	var offset uint32
	err = binary.Read(file, binary.BigEndian, &offset)
	if err != nil {
		return nil, err
	}

	return parse(io.NewSectionReader(file, int64(offset), 1<<63-1), file)
}

// StrictParseCollectionIndex parses a single font from a TrueType
// Collection (.ttc) file with font index starting from 0.
// Each table will be fully parsed and an error is returned if any fail.
func StrictParseCollectionIndex(file File, index uint32) (*Font, error) {
	font, err := ParseCollectionIndex(file, index)
	if err != nil {
		return nil, err
	}

	for _, tag := range font.Tags() {
		if _, err := font.Table(tag); err != nil {
			return nil, fmt.Errorf("failed to parse %q: %w", tag, err)
		}
	}

	return font, nil
}
