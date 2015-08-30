package sfnt

import (
	"bytes"
	"encoding/binary"
	"io"
	"io/ioutil"
	"strconv"
)

type TableName struct {
	bytes []byte
}

type nameHeader struct {
	Format       uint16
	Count        uint16
	StringOffset uint16
}

type nameRecord struct {
	PlatformId         uint16
	PlatformSpecificId uint16
	LanguageId         uint16
	NameId             uint16
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

func (table *TableName) Bytes() []byte {
	return table.bytes
}

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

		results = append(results, strconv.Itoa(int(record.PlatformId))+","+strconv.Itoa(int(record.PlatformSpecificId))+","+
			strconv.Itoa(int(record.LanguageId))+","+strconv.Itoa(int(record.NameId))+" "+strconv.Itoa(int(record.Offset))+" "+string(table.bytes[start:end]))
	}

	return results
}
