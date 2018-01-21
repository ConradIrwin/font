package commands

import (
	"fmt"
	"os"

	"github.com/ConradIrwin/font/sfnt"
)

// Features prints the gpos/gsub tables (contins font features).
func Features() {
	if len(os.Args) < 2 {
		panic(fmt.Errorf("Specify a font file"))
	}

	file, err := os.Open(os.Args[1])
	if err != nil {
		panic(fmt.Errorf("Failed to open font: %s", err))
	}
	defer file.Close()

	font, err := sfnt.Parse(file)
	if err != nil {
		panic(fmt.Errorf("Failed to parse font: %s", err))
	}

	layoutTable(font, sfnt.TagGsub, "Glyph Substitution Table (GSUB)")
	layoutTable(font, sfnt.TagGpos, "Glyph Positioning Table (GPOS)")
}

func layoutTable(font *sfnt.Font, tag sfnt.Tag, name string) {
	if font.HasTable(tag) {
		fmt.Printf("%s:\n", name)

		t := font.Table(tag).(*sfnt.TableLayout)
		for _, script := range t.Scripts {
			fmt.Printf("\tScript %q (%s):\n", script.Tag, script.String())

			fmt.Printf("\t\tDefault Language:\n")
			for _, feature := range script.DefaultLanguage.Features {
				fmt.Printf("\t\t\tFeature %q (%s)\n", feature.Tag, feature.String())
			}

			for _, lang := range script.Languages {
				fmt.Printf("\t\tLanguage %q (%s):\n", lang.Tag, lang.String())
				for _, feature := range lang.Features {
					fmt.Printf("\t\t\tFeature %q (%s)\n", feature.Tag, feature.String())
				}
			}
		}
	} else {
		fmt.Printf("No %s\n", name)
	}
}
