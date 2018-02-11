package main

import (
	"fmt"

	"github.com/ConradIrwin/font/sfnt"
)

// Cmap prints out Character To Glyph information
func Cmap(font *sfnt.Font) error {
	table, err := font.CmapTable()
	if err != nil {
		return err
	}

	for i, encoding := range table.Encodings {
		fmt.Printf("[%d] %s\n", i, encoding)
		//fmt.Printf("[%d] %v\n", i, string(encoding.Characters()))
	}

	return nil
}
