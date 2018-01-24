package main

import (
	"fmt"
	"os"

	"github.com/ConradIrwin/font/commands"
	"github.com/ConradIrwin/font/sfnt"
)

type Command func(*sfnt.Font) error

var cmds = map[string]Command{
	"scrub":    commands.Scrub,
	"info":     commands.Info,
	"stats":    commands.Stats,
	"metrics":  commands.Metrics,
	"features": commands.Features,
}

func usage() {
	fmt.Println(`
Usage: font [features|info|metrics|scrub|stats] font.[otf,ttf,woff]

features: prints the gpos/gsub tables (contains font features)
info: prints the name table (contains metadata)
metrics: prints the hhea table (contains font metrics)
scrub: remove the name table (saves significant space)
stats: prints each table and the amount of space used`)
}

func main() {
	command := "help"
	if len(os.Args) > 1 {
		command = os.Args[1]
		os.Args = os.Args[1:]
	}

	if _, found := cmds[command]; !found {
		usage()
		return
	}

	if len(os.Args) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: font %s <font file>\n", command)
		os.Exit(1)
	}

	filename := os.Args[1]
	file, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open font: %s\n", err)
		os.Exit(1)
	}
	defer file.Close()

	font, err := sfnt.Parse(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse font: %s\n", err)
		os.Exit(1)
	}

	if err := cmds[command](font); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
