package commands

import (
	"bytes"
	"io/ioutil"
	"os"

	"github.com/ConradIrwin/font/sfnt"
)

func Scrub() {

	file, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer file.Close()

	data, err := ioutil.ReadFile(file.Name())
	if err != nil {
		panic(err)
	}

	font, err := sfnt.Parse(bytes.NewReader(data))

	if err != nil {
		panic(err)
	}

	if font.HasTable(sfnt.TagName) {
		font.AddTable(sfnt.TagName, sfnt.NewTableName())
	}

	font.WriteOTF(os.Stdout)
}
