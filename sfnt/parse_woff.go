package sfnt

import (
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
)

type woffHeader struct {
	Signature      Tag
	Flavor         Tag
	Length         uint32
	NumTables      uint16
	Reserved       uint16
	TotalSfntSize  uint32
	Version        fixed
	MetaOffset     uint32
	MetaLength     uint32
	MetaOrigLength uint32
	PrivOffset     uint32
	PrivLength     uint32
}

type woffEntry struct {
	Tag          Tag
	Offset       uint32
	CompLength   uint32
	OrigLength   uint32
	OrigChecksum uint32
}

func parseWoff(file File) (*Font, error) {

	header := woffHeader{}
	err := binary.Read(file, binary.BigEndian, &header)
	if err != nil {
		return nil, err
	}

	font := &Font{header.Flavor, make(map[Tag]Table, header.NumTables)}

	entries := make([]woffEntry, header.NumTables)

	for i := uint16(0); i < header.NumTables; i++ {

		entry := woffEntry{}
		err := binary.Read(file, binary.BigEndian, &entry)
		if err != nil {
			return nil, err
		}
		fmt.Println(entry)
		entries[i] = entry
	}

	for _, entry := range entries {

		var buffer io.Reader
		buffer = io.NewSectionReader(file, int64(entry.Offset), int64(entry.OrigLength))
		if entry.CompLength < entry.OrigLength {
			buffer, err = zlib.NewReader(buffer)
			if err != nil {
				return nil, err
			}
		}
		table, err := parseTable(entry.Tag, buffer)
		if err != nil {
			return nil, err
		}
		font.tables[entry.Tag] = table

	}

	_, ok := font.tables[TagHead]

	if !ok {
		return nil, ErrMissingHead
	}

	return font, nil

}
