package sfnt

import (
	"io"
	"io/ioutil"
)

// Table is an interface for each section of the font file.
type Table interface {
	Bytes() []byte
}

type unparsedTable struct {
	bytes []byte
}

func newUnparsedTable(buffer io.Reader) (*unparsedTable, error) {
	bytes, err := ioutil.ReadAll(buffer)
	if err != nil {
		return nil, err
	}
	return &unparsedTable{bytes}, nil
}

func parseTable(tag Tag, buffer io.Reader) (Table, error) {
	if tag == TagHead {
		return parseTableHead(buffer)
	} else if tag == TagName {
		return parseTableName(buffer)
	} else if tag == TagHhea {
		return parseTableHhea(buffer)
	}
	return newUnparsedTable(buffer)
}
