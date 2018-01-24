package commands

import (
	"fmt"

	"github.com/ConradIrwin/font/sfnt"
)

// Features prints the gpos/gsub tables (contains font features).
func Features(font *sfnt.Font) error {
	if err := layoutTable(font, sfnt.TagGsub, "Glyph Substitution Table (GSUB)"); err != nil {
		return err
	}
	if err := layoutTable(font, sfnt.TagGpos, "Glyph Positioning Table (GPOS)"); err != nil {
		return err
	}
	return nil
}

func layoutTable(font *sfnt.Font, tag sfnt.Tag, name string) error {
	if font.HasTable(tag) {
		fmt.Printf("%s:\n", name)

		t := font.Table(tag).(*sfnt.TableLayout)
		for _, script := range t.Scripts {
			fmt.Printf("\tScript %q%s:\n", script.Tag, bracketString(script))

			fmt.Printf("\t\tDefault Language:\n")
			for _, feature := range script.DefaultLanguage.Features {
				fmt.Printf("\t\t\tFeature %q%s\n", feature.Tag, bracketString(feature))
			}

			for _, lang := range script.Languages {
				fmt.Printf("\t\tLanguage %q%s:\n", lang.Tag, bracketString(lang))
				for _, feature := range lang.Features {
					fmt.Printf("\t\t\tFeature %q%s\n", feature.Tag, bracketString(feature))
				}
			}
		}
	} else {
		fmt.Printf("No %s\n", name)
	}

	return nil
}

func bracketString(o fmt.Stringer) string {
	if s := o.String(); s != "" {
		return fmt.Sprintf(" (%s)", s)
	}
	return ""
}
