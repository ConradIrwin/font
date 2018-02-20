package main

import (
	"fmt"

	"github.com/ConradIrwin/font/sfnt"
)

// Stats prints each table and the amount of space used.
func Stats(font *sfnt.Font) error {
	for _, tag := range font.Tags() {
		table, err := font.Table(tag)
		if err != nil {
			return err
		}

		fmt.Printf("%6d %q %s\n", len(table.Bytes()), tag, table.Name())
	}
	return nil
}
