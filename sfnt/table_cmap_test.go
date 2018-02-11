package sfnt

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

/*

sfnt/testdata/NotoSansCJKjp-Regular.otf
[0] Unicode Encoding 3 (format: 4)
[1] Unicode Encoding 4 (format: 12)
[2] Unicode Encoding 5 (format: 14)
[3] Mac Encoding 1 (format: 6)
[4] Microsoft Unicode BMP (UCS-2) (format: 4)
[5] Microsoft Unicode UCS-4 (format: 12)

sfnt/testdata/NotoColorEmoji.ttf
[0] Unicode Encoding 5 (format: 14)
[1] Microsoft Unicode UCS-4 (format: 12)

sfnt/testdata/open-sans-v15-latin-regular.woff
[0] Microsoft Unicode BMP (UCS-2) (format: 4)

sfnt/testdata/Roboto-BoldItalic.ttf
[0] Unicode Encoding 3 (format: 4)
[1] Unicode Encoding 4 (format: 12)
[2] Microsoft Unicode BMP (UCS-2) (format: 4)
[3] Microsoft Unicode UCS-4 (format: 12)

sfnt/testdata/Raleway-v4020-Regular.otf
[0] Unicode Encoding 3 (format: 4)
[1] Mac Roman (format: 6)
[2] Microsoft Unicode BMP (UCS-2) (format: 4)

*/

// firstCmapEncoding is a helper function to return the first Encoding subtable for the particular format.
func firstCmapEncoding(fontname string, format EncodingFormatID) (CharacterToGlyph, error) {
	filename := filepath.Join("testdata", fontname)
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open %q: %s", filename, err)
	}

	font, err := Parse(file)
	if err != nil {
		return nil, fmt.Errorf("Parse(%q) err = %q, want nil", filename, err)
	}

	cmap, err := font.CmapTable()
	if err != nil {
		return nil, fmt.Errorf("Parse(%q).CmapTable() err = %q, want nil", filename, err)
	}

	for _, encoding := range cmap.Encodings {
		if encoding.Format == format {
			return encoding, nil
		}
	}

	return nil, fmt.Errorf("format %d not found", format)
}

type charTest struct {
	input rune
	want  int
}

func TestCmapFormat4(t *testing.T) {
	tests := []struct {
		font   string
		mapper CharacterToGlyph
		tests  []charTest
	}{
		{
			font: "Fake Example",
			mapper: &cmapFormat4{
				endCode:       []uint16{20, 90, 153, 0xFFFF},
				startCode:     []uint16{10, 30, 100, 0xFFFF},
				idDelta:       []int16{-9, -18, -27, 1},
				idRangeOffset: []uint16{0, 0, 0, 0},
			},
			tests: []charTest{
				{10, 1},
				{20, 11},
				{30, 12},
				{90, 72},

				{0, MissingGlyph},
				{9, MissingGlyph},
				{154, MissingGlyph},
				{0xFFFF, MissingGlyph},
			},
		}, {
			font: "Roboto-BoldItalic.ttf",
			tests: []charTest{
				{0, 1},
				{1, MissingGlyph},
				{2, 2},
				{9, 3},
				{' ', 5},
				{'!', 6},
				{'a', 70},
				{'£', 102},
				{'À', 2253},
				{'Ç', 2260},
				{'È', 2261},
				{0xFFFF, MissingGlyph},
			},
		}, {
			font: "Raleway-v4020-Regular.otf",
			tests: []charTest{
				{0, 848}, // TODO Double check this is right!
				{0xFFFF, MissingGlyph},
			},
		}, {
			font: "open-sans-v15-latin-regular.woff",
			tests: []charTest{
				{0, MissingGlyph},
				{0xFFFF, MissingGlyph},
			},
		},
	}

	for _, test := range tests {
		if test.mapper == nil {
			var err error
			if test.mapper, err = firstCmapEncoding(test.font, EncodingFormatID(4)); err != nil {
				t.Errorf("%s", err)
				continue
			}
		}

		for _, char := range test.tests {
			if got := test.mapper.Map(char.input); got != char.want {
				t.Errorf("[%q] Map(%q) = %d want %d", test.font, char.input, got, char.want)
			}
		}

		// Make sure testing the full range doesn't panic
		for _, c := range test.mapper.Characters() {
			test.mapper.Map(c)
			// TODO Check the returned glyph is within the font.
		}
	}
}
