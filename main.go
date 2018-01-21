package main

import (
	"fmt"
	"os"

	"github.com/ConradIrwin/font/commands"
)

func main() {

	command := "help"

	if len(os.Args) > 1 {
		command = os.Args[1]
		os.Args = os.Args[1:]
	}

	switch command {
	case "scrub":
		commands.Scrub()
	case "info":
		commands.Info()
	case "stats":
		commands.Stats()
	case "metrics":
		commands.Metrics()
	case "features":
		commands.Features()
	default:
		fmt.Println(`
Usage: font [features|info|metrics|scrub|stats] font.[otf,ttf,woff]

features: prints the gpos/gsub tables (contins font features)
info: prints the name table (contains metadata)
metrics: prints the hhea table (contains font metrics)
scrub: remove the name table (saves significant space)
stats: prints each table and the amount of space used`)
	}

}
