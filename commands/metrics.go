package commands

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ConradIrwin/font/sfnt"
)

func Metrics() {

	file, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer file.Close()

	data, err := ioutil.ReadFile(file.Name())
	if err != nil {
		panic(err)
	}

	font, err := sfnt.Parse(bytes.NewReader(data))
	if err != nil {
		panic(err)
	}

	if hhea, ok := font.HheaTable(); ok {
		fmt.Println("Ascent:", hhea.Ascent)
		fmt.Println("Descent:", hhea.Descent)
		fmt.Println("Line gap:", hhea.LineGap)
		fmt.Println("Caret offset:", hhea.CaretOffset)
		fmt.Println("Caret slope rise:", hhea.CaretSlopeRise)
		fmt.Println("Caret slope run:", hhea.CaretSlopeRun)
		fmt.Println("Advance with max:", hhea.AdvanceWidthMax)
		fmt.Println("Min left side bearing:", hhea.MinLeftSideBearing)
		fmt.Println("Min right side bearing:", hhea.MinRightSideBearing)
	} else {
		fmt.Fprintf(os.Stderr, "No %q table\n", sfnt.TagHhea.String())
	}

	if os2, ok := font.OS2Table(); ok {
		fmt.Printf("%#v\n", os2)

		fmt.Println("Cap Height:", os2.SCapHeight)
		fmt.Println("Typographic Ascender:", os2.STypoAscender)
		fmt.Println("Typographic Descender:", os2.STypoDescender)
		fmt.Println("Win Ascent:", os2.UsWinAscent)
		fmt.Println("Win Descent:", os2.UsWinDescent)

		fmt.Println("TODO: SHOW MORE METRICS")
	} else {
		fmt.Fprintf(os.Stderr, "No %q table\n", sfnt.TagOS2.String())
	}
}
