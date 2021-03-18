package sfnt

import (
	"bytes"
	"os"
	"testing"
)

func TestParseCrashers(t *testing.T) {

	font, err := Parse(bytes.NewReader([]byte{}))
	if font != nil || err == nil {
		t.Fail()
	}

}

func TestGSubLookupRecords(t *testing.T) {
	var file, err = os.Open("testdata/Roboto-BoldItalic.ttf")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	sfnt, err := Parse(file)
	if err != nil {
		t.Fatal(err)
	}
	table, err := sfnt.GsubTable()
	if err != nil {
		t.Fatal(err)
	}
	numLookups := len(table.Lookups)
	for i, l := range table.Lookups {
		t.Logf("[%d] lookup record of type %s", i, l.GSubString())
	}
	if numLookups != 35 {
		t.Errorf("font 'Roboto bold italic' has 35 lookup records, found %d", numLookups)
	}
	if table.Lookups[34].Type != 1 {
		t.Errorf("35th lookup record of 'Roboto bold italic' has type 1 (Single), found %s",
			table.Lookups[34].GSubString())
	}
}
