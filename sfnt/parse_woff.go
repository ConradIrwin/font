package sfnt

import (
	"encoding/binary"
	"fmt"
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
	var header woffHeader
	if err := binary.Read(file, binary.BigEndian, &header); err != nil {
		return nil, err
	}

	font := &Font{
		file:       file,
		scalerType: header.Flavor,
		tables:     make(map[Tag]tableSection, header.NumTables),
	}

	for i := uint16(0); i < header.NumTables; i++ {
		var entry woffEntry
		if err := binary.Read(file, binary.BigEndian, &entry); err != nil {
			return nil, err
		}

		// TODO Check the checksum.

		if _, found := font.tables[entry.Tag]; found {
			return nil, fmt.Errorf("found multiple %q tables", entry.Tag)
		}

		font.tables[entry.Tag] = tableSection{
			tag: entry.Tag,

			offset:  entry.Offset,
			length:  entry.OrigLength,
			zLength: entry.CompLength,
		}
	}

	if _, ok := font.tables[TagHead]; !ok {
		return nil, ErrMissingHead
	}

	return font, nil
}
