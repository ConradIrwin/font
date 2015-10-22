package sfnt

import (
	"bytes"
	"encoding/binary"
	"io"
)

type TableHhea struct {
	Version             fixed
	Ascent              int16
	Descent             int16
	LineGap             int16
	AdvanceWidthMax     uint16
	MinLeftSideBearing  int16
	MinRightSideBearing int16
	XMaxExtent          int16
	CaretSlopeRise      int16
	CaretSlopeRun       int16
	CaretOffset         int16
	Reserved1           int16
	Reserved2           int16
	Reserved3           int16
	Reserved4           int16
	MetricDataformat    int16
	NumOfLongHorMetrics int16
}

func parseTableHhea(buffer io.Reader) (*TableHhea, error) {
	table := TableHhea{}
	err := binary.Read(buffer, binary.BigEndian, &table)
	if err != nil {
		return nil, err
	}
	return &table, nil
}

// Bytes returns the byte representation of this header.
func (table *TableHhea) Bytes() []byte {
	buffer := &bytes.Buffer{}
	err := binary.Write(buffer, binary.BigEndian, table)
	// should never happen
	if err != nil {
		panic(err)
	}
	return buffer.Bytes()
}
