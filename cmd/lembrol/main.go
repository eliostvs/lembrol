package main

import (
	"os"

	"github.com/eliostvs/lembrol/internal/terminal"
)

func main() {
	os.Exit(terminal.CLI(os.Args[:], os.Stdout, os.Stderr))
}
