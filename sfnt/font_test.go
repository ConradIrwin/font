package sfnt

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

// TestSmokeTest very simply checks we can parse, and write the sample fonts
// without error.
// TODO We should check what is returned is valid.
func TestSmokeTest(t *testing.T) {
	tests := []struct {
		filename string
	}{
		{filename: "Roboto-BoldItalic.ttf"},
		{filename: "Raleway-v4020-Regular.otf"},
		{filename: "open-sans-v15-latin-regular.woff"},
	}

	for _, test := range tests {
		filename := filepath.Join("testdata", test.filename)
		file, err := os.Open(filename)
		if err != nil {
			t.Errorf("Failed to open %q: %s\n", filename, err)
		}

		font, err := StrictParse(file)
		if err != nil {
			t.Errorf("StrictParse(%q) err = %q, want nil", filename, err)
			continue
		}

		if _, err := font.WriteOTF(ioutil.Discard); err != nil {
			t.Errorf("WriteOTF(%q) err = %q, want nil", filename, err)
			continue
		}

		file.Close()
	}
}

// benchmarkParse tests the performance of a simple Parse.
// Example run:
//   go test -cpuprofile cpu.prof -benchmem -memprofile mem.prof -bench . -run=^$ -benchtime=30s github.com/ConradIrwin/font/sfnt
//   go tool pprof cpu.prof
//
// BenchmarkParseOtf-8             20000000              2838 ns/op            1424 B/op         32 allocs/op
// BenchmarkStrictParseOtf-8         200000            193842 ns/op          372613 B/op       1617 allocs/op
// BenchmarkParseWoff-8            10000000              3742 ns/op            2005 B/op         41 allocs/op
// BenchmarkStrictParseWoff-8        100000            660052 ns/op          575993 B/op        498 allocs/op
func benchmarkParse(b *testing.B, filename string) {
	buf, err := ioutil.ReadFile(filepath.Join("testdata", filename))
	if err != nil {
		b.Errorf("Failed to open %q: %s\n", filename, err)
	}

	for n := 0; n < b.N; n++ {
		r := bytes.NewReader(buf)
		if _, err := Parse(r); err != nil {
			b.Errorf("Parse(%q) err = %q, want nil", filename, err)
			return
		}
	}
}

// benchmarkStrictParse tests the performance of a simple StrictParse.
func benchmarkStrictParse(b *testing.B, filename string) {
	buf, err := ioutil.ReadFile(filepath.Join("testdata", filename))
	if err != nil {
		b.Errorf("Failed to open %q: %s\n", filename, err)
	}

	for n := 0; n < b.N; n++ {
		r := bytes.NewReader(buf)
		if _, err := StrictParse(r); err != nil {
			b.Errorf("StrictParse(%q) err = %q, want nil", filename, err)
			return
		}
	}
}

func BenchmarkParseOtf(b *testing.B) {
	benchmarkParse(b, "Roboto-BoldItalic.ttf")
}

func BenchmarkStrictParseOtf(b *testing.B) {
	benchmarkStrictParse(b, "Roboto-BoldItalic.ttf")
}

func BenchmarkParseWoff(b *testing.B) {
	benchmarkParse(b, "open-sans-v15-latin-regular.woff")
}

func BenchmarkStrictParseWoff(b *testing.B) {
	benchmarkStrictParse(b, "open-sans-v15-latin-regular.woff")
}
