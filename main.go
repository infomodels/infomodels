package main

import (
	"fmt"
	"os"

	"github.com/infomodels/infomodels/cmd"
)

func main() {

	// Handle the version printing at this level, where the progVersion is
	// available. The progVersion is defined in this package so that the
	// git sha and build num can be defined in the Makefile.
	for _, arg := range os.Args {
		if arg == "--version" {
			fmt.Printf("infomodels, version %s\n", progVersion)
			os.Exit(0)
		}
	}

	// Run the CLI from the cmd package.
	cmd.Execute()
	os.Exit(0)
}
