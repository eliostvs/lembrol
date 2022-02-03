package terminal_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/eliostvs/lembrol/internal/terminal"
)

func TestCLI(t *testing.T) {
	name := "lembrol"
	version := "0.0.0-dev"

	t.Run("shows help when help flags is given", func(t *testing.T) {
		var output bytes.Buffer

		got := terminal.CLI([]string{name, "-h"}, version, &output)

		assert.Equal(t, 0, got)
		want := `lembrol 0.0.0-dev

Learning things through spaced repetition.

USAGE:
  lembrol [options]

OPTIONS:
  -decks string
    	deck files location (default ".")
`
		assert.Equal(t, want, output.String())
	})

	t.Run("fails when an invalid flag is given", func(t *testing.T) {
		var output bytes.Buffer

		got := terminal.CLI([]string{name, "-foo"}, version, &output)

		assert.Equal(t, 1, got)
		assert.Contains(t, output.String(), "flag provided but not defined: -foo")
	})
}
