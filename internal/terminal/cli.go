package terminal

import (
	"errors"
	"flag"
	"fmt"
	"io"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	paramErrResult   = 1
	programErrResult = 2
	successResult    = 0
)

func CLI(args []string, version string, output io.Writer) int {
	fl := flag.NewFlagSet(args[0], flag.ContinueOnError)
	fl.SetOutput(output)
	fl.Usage = func() {
		fmt.Fprintf(fl.Output(), "\nLearning things through spaced repetition.\n\n")
		fmt.Fprintf(fl.Output(), "Usage:\n  %s [options]\n\n", args[0])
		fmt.Fprintf(fl.Output(), "Example:\n  %s -decks ./decks/location\n\n", args[0])
		fmt.Fprintln(fl.Output(), "Options:")
		fl.PrintDefaults()
	}
	decksLocation := fl.String("decks", ".", "deck files location")

	if err := fl.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return successResult
		}
		return paramErrResult
	}

	program := tea.NewProgram(NewModel(*decksLocation))
	if err := program.Start(); err != nil {
		fmt.Fprintf(output, "failed: %v", err)
		return programErrResult
	}

	return successResult
}
