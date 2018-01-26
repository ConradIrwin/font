package commands

import (
	"fmt"

	"github.com/ConradIrwin/font/sfnt"
)

func Metrics(font *sfnt.Font) error {
	if font.HasTable(sfnt.TagHhea) {
		hhea := font.HheaTable()
		fmt.Println("Ascent:", hhea.Ascent)
		fmt.Println("Descent:", hhea.Descent)
		fmt.Println("Line gap:", hhea.LineGap)
		fmt.Println("Caret offset:", hhea.CaretOffset)
		fmt.Println("Caret slope rise:", hhea.CaretSlopeRise)
		fmt.Println("Caret slope run:", hhea.CaretSlopeRun)
		fmt.Println("Advance with max:", hhea.AdvanceWidthMax)
		fmt.Println("Min left side bearing:", hhea.MinLeftSideBearing)
		fmt.Println("Min right side bearing:", hhea.MinRightSideBearing)
	}

	if font.HasTable(sfnt.TagOS2) {
		fmt.Printf("%#v\n", font.OS2Table())

		os2 := font.OS2Table()

		fmt.Println("Cap Height:", os2.SCapHeight)
		fmt.Println("Typographic Ascender:", os2.STypoAscender)
		fmt.Println("Typographic Descender:", os2.STypoDescender)
		fmt.Println("Win Ascent:", os2.UsWinAscent)
		fmt.Println("Win Descent:", os2.UsWinDescent)

		fmt.Println("TODO: SHOW MORE METRICS")
	}

	return nil
}
