package sfnt

import (
	"compress/zlib"
	"io"
	"io/ioutil"
)

var parsers = map[Tag]tableParser{
	TagHead: parseTableHead,
	TagName: parseTableName,
	TagHhea: parseTableHhea,
	TagOS2:  parseTableOS2,
	TagGpos: parseTableLayout,
	TagGsub: parseTableLayout,
	TagCmap: parseTableCmap,
}

// Table is an interface for each section of the font file.
type Table interface {
	Bytes() []byte
}

type unparsedTable struct {
	bytes []byte
}

type tableParser func(buffer io.Reader) (Table, error)

func newUnparsedTable(buffer io.Reader) (Table, error) {
	bytes, err := ioutil.ReadAll(buffer)
	if err != nil {
		return nil, err
	}
	return &unparsedTable{bytes}, nil
}

func (font *Font) parseTable(s tableSection) (Table, error) {
	var reader io.Reader
	reader = io.NewSectionReader(font.file, int64(s.offset), int64(s.length))
	if s.zLength > 0 && s.zLength < s.length {
		var err error
		if reader, err = zlib.NewReader(reader); err != nil {
			return nil, err
		}
	}

	parser, found := parsers[s.tag]
	if !found {
		parser = newUnparsedTable
	}

	return parser(reader)
}
