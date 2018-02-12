package sfnt

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var testFiles = []string{
	"Roboto-BoldItalic.ttf",
	"Raleway-v4020-Regular.otf",
	"open-sans-v15-latin-regular.woff",
	"NotoColorEmoji.ttf",
	"NotoSansCJKjp-Regular.otf",
	// "NotoSerifCJK-Regular.ttc", // TTC files not supported, yet.
}

// TestSmokeTest very simply checks we can parse, and write the sample fonts
// without error.
// TODO We should check what is returned is valid.
func TestSmokeTest(t *testing.T) {
	for _, filename := range testFiles {
		filename = filepath.Join("testdata", filename)
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
//   go test -cpuprofile cpu.prof -bench . -run=^$ -benchtime=30s
//   go tool pprof cpu.prof
//
// BenchmarkParseOtf-8          	 5000000	      2784 ns/op	    1440 B/op	      33 allocs/op
// BenchmarkStrictParseOtf-8    	  100000	    185088 ns/op	  372422 B/op	    1615 allocs/op
// BenchmarkParseWoff-8         	 5000000	      3573 ns/op	    2005 B/op	      41 allocs/op
// BenchmarkStrictParseWoff-8   	   20000	    615948 ns/op	  543514 B/op	     484 allocs/op
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
