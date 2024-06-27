package sfnt

import (
	"bytes"
	"encoding/binary"
)

type v5Fields struct {
	UsLowerPointSize uint16
	UsUpperPointSize uint16
}

type v4Fields struct {
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
}

type TableOS2 struct {
	baseTable
	v4Fields
	v5Fields
	bytes []byte
}

func parseTableOS2(tag Tag, buf []byte) (Table, error) {
	r := bytes.NewBuffer(buf)

	var v4fields v4Fields
	var v5fields v5Fields
	if err := binary.Read(r, binary.BigEndian, &v4fields); err != nil {
		if err != nil {
			return nil, err
		}
	}
	if v4fields.Version == 5 {
		if err := binary.Read(r, binary.BigEndian, &v5fields); err != nil {
			if err != nil {
				return nil, err
			}
		}
	}

	return &TableOS2{
		baseTable: baseTable(tag),
		v4Fields:  v4fields,
		v5Fields:  v5fields,
		bytes:     buf,
	}, nil
}

func (t *TableOS2) Bytes() []byte {
	return t.bytes
}
