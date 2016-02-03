package sfnt

import (
	"bytes"
	"encoding/binary"
	"io"
	"io/ioutil"
)

type tableOS2Fields struct {
	Version             uint16
	XAvgCharWidth       uint16
	USWeightClass       uint16
	USWidthClass        uint16
	FSType              uint16
	YSubscriptXSize     int16
	YSubscriptYSize     int16
	YSubscriptXOffset   int16
	YSubscriptYOffset   int16
	YSuperscriptXSize   int16
	YSuperscriptYSize   int16
	YSuperscriptXOffset int16
	YSuperscriptYOffset int16
	YStrikeoutSize      int16
	YStrikeoutPosition  int16
	SFamilyClass        int16
	Panose              [10]byte
	UlCharRange         [4]uint32
	AchVendID           Tag
	FsSelection         uint16
	FsFirstCharIndex    uint16
	FsLastCharIndex     uint16
	STypoAscender       int16
	STypoDescender      int16
	STypoLineGap        int16
	UsWinAscent         uint16
	UsWinDescent        uint16
	UlCodePageRange1    uint32
	UlCodePageRange2    uint32
	SxHeigh             int16
	SCapHeight          int16
	UsDefaultChar       uint16
	UsBreakChar         uint16
	UsMaxContext        uint16
	UsLowerPointSize    uint16
	UsUpperPointSize    uint16
}

type TableOS2 struct {
	tableOS2Fields
	bytes []byte
}

func parseTableOS2(buffer io.Reader) (*TableOS2, error) {
	rawBytes, err := ioutil.ReadAll(buffer)

	if err != nil {
		return nil, err
	}

	mySize := binary.Size(tableOS2Fields{})
	buf := make([]byte, mySize, mySize)

	copy(buf, rawBytes)

	readable := bytes.NewBuffer(buf)

	table := tableOS2Fields{}
	err = binary.Read(readable, binary.BigEndian, &table)
	if err != nil {
		return nil, err
	}

	return &TableOS2{tableOS2Fields: table, bytes: rawBytes}, nil
}

func (t *TableOS2) Bytes() []byte {
	return t.bytes
}
