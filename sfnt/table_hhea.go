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

func parseTableHhea(r io.Reader) (Table, error) {
	var table TableHhea
	if err := binary.Read(r, binary.BigEndian, &table); err != nil {
		return nil, err
	}
	return &table, nil
}

// Bytes returns the byte representation of this header.
func (table *TableHhea) Bytes() []byte {
	var buffer bytes.Buffer
	if err := binary.Write(&buffer, binary.BigEndian, table); err != nil {
		panic(err) // should never happen
	}
	return buffer.Bytes()
}
