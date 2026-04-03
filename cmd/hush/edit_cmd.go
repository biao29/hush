package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/getctx/hush/internal/config"
	"github.com/getctx/hush/internal/project"
	"github.com/spf13/cobra"
)

func newEditCmd() *cobra.Command {
	var public bool

	cmd := &cobra.Command{
		Use:   "edit [project]",
		Short: "Open private rules in your editor",
		Long:  "Opens AGENTS.override.md (or AGENTS.md with --public) in $EDITOR. Auto-resolves project from CWD or name.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireInit(); err != nil {
				printer.PrintError(err.Error())
				return err
			}

			cfg, err := config.Load()
			if err != nil {
				printer.PrintError(fmt.Sprintf("load config: %v", err))
				return err
			}

			name := ""
			if len(args) > 0 {
				name = args[0]
			}

			entry, err := project.Resolve(cfg, name)
			if err != nil {
				printer.PrintError(err.Error())
				return err
			}

			filename := "AGENTS.override.md"
			if public {
				filename = "AGENTS.md"
			}
			filePath := filepath.Join(entry.Path, filename)

			// Ensure file exists for override.
			if !public {
				if _, err := os.Stat(filePath); os.IsNotExist(err) {
					content := "# Private Rules\n#\n# This file is managed by hush and gitignored.\n"
					os.WriteFile(filePath, []byte(content), 0o644)
				}
			}

			editor := getEditor()
			printer.Debug("opening %s with %s", filePath, editor)

			return runEditor(editor, filePath)
		},
	}

	cmd.Flags().BoolVar(&public, "public", false, "Edit AGENTS.md (public rules) instead of override")
	return cmd
}

func getEditor() string {
	if e := os.Getenv("EDITOR"); e != "" {
		return e
	}
	if e := os.Getenv("VISUAL"); e != "" {
		return e
	}
	// Try common editors.
	for _, e := range []string{"vim", "vi", "nano"} {
		if _, err := exec.LookPath(e); err == nil {
			return e
		}
	}
	return "vi"
}

// runEditor launches the editor, handling compound EDITOR values like "code --wait".
// It delegates to the user's shell so that quoting and arguments work correctly.
func runEditor(editor, filePath string) error {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "sh"
	}
	// Use shell -c so "code --wait", "nvim -f", etc. are parsed correctly.
	c := exec.Command(shell, "-c", editor+" "+shellQuote(filePath))
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

// shellQuote wraps a path in single quotes for safe shell interpolation.
func shellQuote(s string) string {
	// Replace single quotes with '\'' (end quote, escaped quote, start quote).
	quoted := "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
	return quoted
}
