package commands

import (
	"github.com/ConradIrwin/font/sfnt"
	"os"
)

func Scrub(font *sfnt.Font) error {
	if font.HasTable(sfnt.TagName) {
		font.AddTable(sfnt.TagName, sfnt.NewTableName())
	}

	_, err := font.WriteOTF(os.Stdout)
	return err
}
