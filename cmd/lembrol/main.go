package main

import (
	"os"

	"github.com/eliostvs/lembrol/internal/terminal"
)

var Version = "0.0.0-dev"

func main() {
	os.Exit(terminal.CLI(os.Args[:], Version, os.Stderr))
}
