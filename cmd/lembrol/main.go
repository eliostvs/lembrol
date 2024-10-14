package main

import (
	"os"

	"github.com/eliostvs/lembrol/internal/tui"
)

func main() {
	os.Exit(tui.CLI(os.Args[:], os.Stdout, os.Stderr))
}
