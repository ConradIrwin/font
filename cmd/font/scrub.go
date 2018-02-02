package main

import (
	"os"

	"github.com/ConradIrwin/font/sfnt"
)

// Scrub remove the name table (saves significant space).
func Scrub(font *sfnt.Font) error {
	if font.HasTable(sfnt.TagName) {
		font.AddTable(sfnt.TagName, sfnt.NewTableName())
	}

	_, err := font.WriteOTF(os.Stdout)
	return err
}
