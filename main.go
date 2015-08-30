package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ConradIrwin/font/sfnt"
)

func main() {

	data, _ := ioutil.ReadAll(os.Stdin)

	font, err := sfnt.Parse(bytes.NewReader(data))

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if font != nil {
		if font.Type() == sfnt.TypeOpenType {
			fmt.Println("OpenType Font")

		} else if font.Type() == sfnt.TypeTrueType || font.Type() == sfnt.TypeAppleTrueType {
			fmt.Println("TrueType Font")
		} else if font.Type() == sfnt.TypePostScript1 {
			fmt.Println("PostScript Type 1 Font")
		} else {
			fmt.Println("Unknown Font Format")
		}

		if font.HasTable(sfnt.TagName) {
			for _, line := range font.NameTable().List() {
				fmt.Println(line)
			}
		} else {
			fmt.Println("(no metadata)")
		}
	}

}
