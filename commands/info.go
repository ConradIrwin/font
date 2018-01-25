package commands

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/ConradIrwin/font/sfnt"
)

func Info() {
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

	if name, ok := font.NameTable(); ok {
		for _, entry := range name.List() {
			ids := " (" + strconv.Itoa(int(entry.PlatformID)) + "," + strconv.Itoa(int(entry.EncodingID)) + "," + strconv.Itoa(int(entry.LanguageID)) + "," + strconv.Itoa(int(entry.NameID)) + ") "
			fmt.Println(entry.Platform() + ids + entry.Label() + ": " + entry.String())
		}
	} else {
		fmt.Fprintf(os.Stderr, "Missing %q table\n", sfnt.TagName.String())
	}
}
