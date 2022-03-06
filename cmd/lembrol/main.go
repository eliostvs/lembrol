package main

import (
	"os"

	"github.com/eliostvs/lembrol/internal/terminal"
)

var version = "dev"

func main() {
	os.Exit(terminal.CLI(os.Args[:], version, os.Stderr))
}
