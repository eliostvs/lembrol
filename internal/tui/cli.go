package tui

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/urfave/cli/v3"

	"github.com/eliostvs/lembrol/internal/version"
)

func CLI(args []string, stdout io.Writer, stderr io.Writer) int {
	const (
		logEnabledFlag = "log"
		logFileFlag    = "log-file"
		decksPath      = "decks"
	)

	cmd := &cli.Command{
		Name:      strings.ToLower(appName),
		Usage:     "Learning things through spaced repetition.",
		Writer:    stdout,
		ErrWriter: stderr,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  logEnabledFlag,
				Value: false,
				Usage: "enable logs",
			},
			&cli.StringFlag{
				Name:  logFileFlag,
				Value: "debug.log",
				Usage: "define the log file",
			},
			&cli.StringFlag{
				Name:  decksPath,
				Value: getDataHome(),
				Usage: "path to directory contains decks",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Bool(logEnabledFlag) {
				file, err := tea.LogToFile(cmd.String(logFileFlag), appName)
				if err != nil {
					return fmt.Errorf("failed to configure logging: %w", err)
				}
				defer file.Close()
			}

			program := tea.NewProgram(NewModel(cmd.String(decksPath)), tea.WithAltScreen())
			_, err := program.Run()
			return err
		},
		Commands: []*cli.Command{
			{
				Name:    "version",
				Aliases: []string{"v"},
				Usage:   "Show version",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					_, _ = fmt.Fprintf(stdout, "%s %s %s\n\n", appName, version.Version, version.Time)
					return nil
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), args); err != nil {
		_, _ = fmt.Fprintf(stderr, "failed: %v\n", err)
		return -1
	}

	return 0
}

func getDataHome() string {
	homeDir, _ := os.UserHomeDir()
	xdgDataHome := os.Getenv("XDG_DATA_HOME")

	if xdgDataHome == "" {
		xdgDataHome = filepath.Join(homeDir, ".local", "share")
	}

	return filepath.Join(xdgDataHome, strings.ToLower(appName))
}
