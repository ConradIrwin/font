package commands

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ConradIrwin/font/sfnt"
)

func Stats() {

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

	for _, tag := range font.Tags() {
		table := font.Table(tag)
		fmt.Println(tag, len(table.Bytes()))
	}

}
