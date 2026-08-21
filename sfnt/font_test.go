package sfnt

import (
	"bytes"
	"io"
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
		{filename: "Go-Regular.woff2"},
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

		if _, err := font.WriteOTF(io.Discard); err != nil {
			t.Errorf("WriteOTF(%q) err = %q, want nil", filename, err)
			continue
		}

		file.Close()
	}
}

// TestIsCollection checks that IsCollection correctly identifies TTC files.
func TestIsCollection(t *testing.T) {
	tests := []struct {
		filename string
		want     bool
	}{
		{filename: "TestTTC.ttc", want: true},
		{filename: "Roboto-BoldItalic.ttf", want: false},
	}

	for _, test := range tests {
		filename := filepath.Join("testdata", test.filename)
		file, err := os.Open(filename)
		if err != nil {
			t.Errorf("Failed to open %q: %s\n", filename, err)
		}

		got, err := IsCollection(file)
		if err != nil {
			t.Errorf("IsCollection(%q) err = %q, want nil", filename, err)
			continue
		}
		if got != test.want {
			t.Errorf("IsCollection(%q) = %v, want %v", filename, got, test.want)
		}

		file.Close()
	}
}

// TestCollectionSmokeTest checks we can parse and write fonts
// from a TTC file.
func TestCollectionSmokeTest(t *testing.T) {
	filename := filepath.Join("testdata", "TestTTC.ttc")
	file, err := os.Open(filename)
	if err != nil {
		t.Errorf("Failed to open %q: %s\n", filename, err)
	}

	fonts, err := ParseCollection(file)
	if err != nil {
		t.Errorf("ParseCollection(%q) err = %q, want nil", filename, err)
	}

	if len(fonts) != 2 {
		t.Errorf("ParseCollection(%q) returned %d fonts, want 2", filename, len(fonts))
	}

	for i, font := range fonts {
		if _, err := font.WriteOTF(io.Discard); err != nil {
			t.Errorf("WriteOTF(%q[%d]) err = %q, want nil", filename, i, err)
			continue
		}
	}

	file.Close()
}

// TestParseCollectionIndex checks that ParseCollectionIndex can parse
// individual fonts from a TTC file by index.
func TestParseCollectionIndex(t *testing.T) {
	filename := filepath.Join("testdata", "TestTTC.ttc")
	file, err := os.Open(filename)
	if err != nil {
		t.Errorf("Failed to open %q: %s\n", filename, err)
	}

	for i := uint32(0); i < 2; i++ {
		file.Seek(0, io.SeekStart)
		font, err := ParseCollectionIndex(file, i)
		if err != nil {
			t.Errorf("ParseCollectionIndex(%q, %d) err = %q, want nil", filename, i, err)
			continue
		}

		if _, err := font.WriteOTF(io.Discard); err != nil {
			t.Errorf("WriteOTF(%q[%d]) err = %q, want nil", filename, i, err)
			continue
		}
	}

	file.Close()
}

// TestParseCollectionIndexOutOfBounds checks that ParseCollectionIndex returns
// an error when given an out-of-bounds index.
func TestParseCollectionIndexOutOfBounds(t *testing.T) {
	filename := filepath.Join("testdata", "TestTTC.ttc")
	file, err := os.Open(filename)
	if err != nil {
		t.Errorf("Failed to open %q: %s\n", filename, err)
	}

	_, err = ParseCollectionIndex(file, 2)
	if err == nil {
		t.Errorf("ParseCollectionIndex(%q, 2) err = nil, want error for out-of-bounds index", filename)
	}

	file.Close()
}

// TestParseCollectionWithNonTTC checks that ParseCollection gracefully handles
// non-TTC files by returning a single-element slice.
func TestParseCollectionWithNonTTC(t *testing.T) {
	filename := filepath.Join("testdata", "Roboto-BoldItalic.ttf")
	file, err := os.Open(filename)
	if err != nil {
		t.Errorf("Failed to open %q: %s\n", filename, err)
	}

	fonts, err := ParseCollection(file)
	if err != nil {
		t.Errorf("ParseCollection(%q) err = %q, want nil", filename, err)
	}

	if len(fonts) != 1 {
		t.Errorf("ParseCollection(%q) returned %d fonts, want 1", filename, len(fonts))
	}

	file.Close()
}

// TestParseCollectionHugeNumFonts checks that a TrueType Collection header
// declaring an absurd font count does not cause a huge up-front allocation.
// Previously the parser did make([]*Font, header.NumFonts) before reading any
// offset, so a 12-byte file could request tens of gigabytes and crash the
// process with "runtime: out of memory".
func TestParseCollectionHugeNumFonts(t *testing.T) {
	var buf bytes.Buffer
	buf.WriteString("ttcf")                   // magic
	buf.Write([]byte{0x00, 0x01, 0x00, 0x00}) // version 1.0
	buf.Write([]byte{0xff, 0xff, 0xff, 0xf0}) // numFonts ~= 4.29 billion
	// No offset table follows, so the very first offset read must fail.

	fonts, err := ParseCollection(bytes.NewReader(buf.Bytes()))
	if err == nil {
		t.Fatalf("ParseCollection(huge numFonts) err = nil, want error; got %d fonts", len(fonts))
	}
}

// benchmarkParse tests the performance of a simple Parse.
// Example run:
//   go test -cpuprofile cpu.prof -benchmem -memprofile mem.prof -bench . -run=^$ -benchtime=30s github.com/ConradIrwin/font/sfnt
//   go tool pprof cpu.prof
//
// BenchmarkParseOTF-8           	20000000	      3209 ns/op	    1229 B/op	      32 allocs/op
// BenchmarkStrictParseOTF-8     	  200000	    184822 ns/op	  372415 B/op	    1616 allocs/op
// BenchmarkParseWOFF-8          	10000000	      3999 ns/op	    1993 B/op	      40 allocs/op
// BenchmarkStrictParseWOFF-8    	   50000	    776500 ns/op	  575990 B/op	     497 allocs/op
// BenchmarkParseWOFF2-8         	   20000	   2011769 ns/op	  742531 B/op	     468 allocs/op
// BenchmarkStrictParseWOFF2-8   	   20000	   2033596 ns/op	  875608 B/op	     818 allocs/op
func benchmarkParse(b *testing.B, filename string) {
	buf, err := os.ReadFile(filepath.Join("testdata", filename))
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
	buf, err := os.ReadFile(filepath.Join("testdata", filename))
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

func BenchmarkParseOTF(b *testing.B) {
	benchmarkParse(b, "Roboto-BoldItalic.ttf")
}

func BenchmarkStrictParseOTF(b *testing.B) {
	benchmarkStrictParse(b, "Roboto-BoldItalic.ttf")
}

func BenchmarkParseWOFF(b *testing.B) {
	benchmarkParse(b, "open-sans-v15-latin-regular.woff")
}

func BenchmarkStrictParseWOFF(b *testing.B) {
	benchmarkStrictParse(b, "open-sans-v15-latin-regular.woff")
}

func BenchmarkParseWOFF2(b *testing.B) {
	benchmarkParse(b, "Go-Regular.woff2")
}

func BenchmarkStrictParseWOFF2(b *testing.B) {
	benchmarkStrictParse(b, "Go-Regular.woff2")
}

// benchmarkParseCollection tests the performance of ParseCollection.
func benchmarkParseCollection(b *testing.B, filename string) {
	buf, err := os.ReadFile(filepath.Join("testdata", filename))
	if err != nil {
		b.Errorf("Failed to open %q: %s\n", filename, err)
	}

	for n := 0; n < b.N; n++ {
		r := bytes.NewReader(buf)
		if _, err := ParseCollection(r); err != nil {
			b.Errorf("ParseCollection(%q) err = %q, want nil", filename, err)
			return
		}
	}
}

func BenchmarkParseCollectionTTC(b *testing.B) {
	benchmarkParseCollection(b, "TestTTC.ttc")
}
