package commands

import (
	"fmt"

	"github.com/ConradIrwin/font/sfnt"
)

func Stats(font *sfnt.Font) error {
	for _, tag := range font.Tags() {
		table := font.Table(tag)
		fmt.Println(tag, len(table.Bytes()))
	}
	return nil
}
