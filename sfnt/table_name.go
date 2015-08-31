package sfnt

import (
	"bytes"
	"encoding/binary"
	"io"
	"io/ioutil"
	"strconv"
)

// TableName represents the OpenType 'name' table. This contains
// human-readable meta-data about the font, for example the Author
// and Copyright.
// https://developer.apple.com/fonts/TrueType-Reference-Manual/RM06/Chap6name.html
type TableName struct {
	bytes []byte
}

type nameHeader struct {
	Format       uint16
	Count        uint16
	StringOffset uint16
}

type nameRecord struct {
	PlatformID         uint16
	PlatformSpecificID uint16
	LanguageID         uint16
	NameID             uint16
	Length             uint16
	Offset             uint16
}

func parseTableName(buffer io.Reader) (*TableName, error) {
	bytes, err := ioutil.ReadAll(buffer)
	if err != nil {
		return nil, err
	}
	return &TableName{bytes}, nil
}

// Bytes returns the representation of this table to be stored in a font.
func (table *TableName) Bytes() []byte {
	return table.bytes
}

// List returns a list of all the strings defined in this table.
func (table *TableName) List() []string {
	reader := bytes.NewBuffer(table.bytes)

	header := nameHeader{}
	err := binary.Read(reader, binary.BigEndian, &header)
	if err != nil {
		panic(err)
	}

	results := make([]string, 0, header.Count)

	for i := uint16(0); i < header.Count; i++ {
		record := nameRecord{}
		err := binary.Read(reader, binary.BigEndian, &record)
		if err != nil {
			panic(err)
		}

		start := header.StringOffset + record.Offset
		end := start + record.Length

		results = append(results, strconv.Itoa(int(record.PlatformID))+","+strconv.Itoa(int(record.PlatformSpecificID))+","+
			strconv.Itoa(int(record.LanguageID))+","+strconv.Itoa(int(record.NameID))+" "+strconv.Itoa(int(record.Offset))+" "+string(table.bytes[start:end]))
	}

	return results
}
