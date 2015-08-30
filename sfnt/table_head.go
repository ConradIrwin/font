package sfnt

import (
	"bytes"
	"encoding/binary"
	"io"
)

// TableHead contains critical information about the rest of the font.
// https://developer.apple.com/fonts/TrueType-Reference-Manual/RM06/Chap6head.html
type TableHead struct {
	VersionNumber      fixed
	FontRevision       fixed
	CheckSumAdjustment uint32
	MagicNumber        uint32
	Flags              uint16
	UnitsPerEm         uint16
	Created            longdatetime
	Updated            longdatetime
	XMin               int16
	YMin               int16
	XMax               int16
	YMax               int16
	MacStyle           uint16
	LowestRecPPEM      uint16
	FontDirection      int16
	IndexToLocFormat   int16
	GlyphDataFormat    int16
}

func parseTableHead(buffer io.Reader) (*TableHead, error) {
	table := TableHead{}
	err := binary.Read(buffer, binary.BigEndian, &table)
	if err != nil {
		return nil, err
	}
	return &table, nil
}

// Bytes returns the byte representation of this header.
func (table *TableHead) Bytes() []byte {
	buffer := &bytes.Buffer{}
	err := binary.Write(buffer, binary.BigEndian, table)
	// should never happen
	if err != nil {
		panic(err)
	}
	return buffer.Bytes()
}

// ExpectedChecksum is the checksum that the file should have had.
func (table *TableHead) ExpectedChecksum() uint32 {
	return 0xB1B0AFBA - table.CheckSumAdjustment
}

// SetExpectedChecksum updates the table so it can be output with the correct checksum.
func (table *TableHead) SetExpectedChecksum(checksum uint32) {
	table.CheckSumAdjustment = 0xB1B0AFBA - checksum
}

// ClearExpectedChecksum updates the table so that the checksum can be calculated.
func (table *TableHead) ClearExpectedChecksum() {
	table.CheckSumAdjustment = 0
}
