package terminal

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

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
		fmt.Fprintf(fl.Output(), "%s %s", args[0], version)
		fmt.Fprintf(fl.Output(), "\n\nLearning things through spaced repetition.")
		fmt.Fprintf(fl.Output(), "\n\nUSAGE:\n  %s [options]", args[0])
		fmt.Fprintln(fl.Output(), "\n\nOPTIONS:")
		fl.PrintDefaults()
	}
	decksLocation := fl.String("decks", ".", "deck files location")

	if err := fl.Parse(args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return successResult
		}
		return paramErrResult
	}

	if os.Getenv("LEMBROL_DEBUG") != "" {
		if f, err := tea.LogToFile("debug.log", "help"); err != nil {
			fmt.Println("Couldn't open a file for logging:", err)
			os.Exit(1)
		} else {
			defer f.Close()
		}
	}

	program := tea.NewProgram(NewModel(*decksLocation), tea.WithAltScreen())
	if err := program.Start(); err != nil {
		fmt.Fprintf(output, "failed: %v", err)
		return programErrResult
	}

	return successResult
}
