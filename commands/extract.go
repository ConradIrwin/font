package commands

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/ConradIrwin/font/sfnt"
)

func Extract() {

	if len(os.Args) <= 1 {
		fmt.Println("usage: font extract <font-collection>")
		os.Exit(1)
	}

	file, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Println(os.Args[1] + ": " + err.Error())
		os.Exit(1)
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}

	fonts, err := sfnt.ParseCollection(bytes.NewReader(data))
	if err != nil {
		panic(err)
	}

	for i, f := range fonts {
		writer, err := os.Create(fmt.Sprintf("%s-%v.ttf", strings.Replace(os.Args[1], ".ttc", "", -1), i))

		if err != nil {
			panic(err)
		}

		_, err = f.WriteOTF(writer)

		if err != nil {
			panic(err)
		}

	}
}
