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

		fmt.Println(tag, len(table.Bytes()))
	}
	return nil
}
