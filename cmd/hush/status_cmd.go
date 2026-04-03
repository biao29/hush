package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/getctx/hush/internal/config"
	"github.com/getctx/hush/internal/output"
	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show sync status for all projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				printer.PrintError(fmt.Sprintf("load config: %v", err))
				return err
			}

			if len(cfg.Projects) == 0 {
				printer.Print(output.Response{
					Summary: "No projects registered — run 'hush link' in a project directory",
				})
				return nil
			}

			hostname, _ := os.Hostname()
			projects := cfg.ProjectsForHost(hostname)

			var rows [][]string
			var statuses []map[string]any

			for _, p := range projects {
				status := "ok"
				detail := ""

				if _, err := os.Stat(p.Path); err != nil {
					status = "unlinked"
					detail = "path not found"
				} else if _, err := os.Stat(filepath.Join(p.Path, "AGENTS.override.md")); err != nil {
					status = "missing"
					detail = "no AGENTS.override.md"
				} else {
					detail = p.Path
				}

				statusIcon := statusSymbol(status)
				rows = append(rows, []string{statusIcon, p.ID, detail})
				statuses = append(statuses, map[string]any{
					"id":     p.ID,
					"status": status,
					"path":   p.Path,
				})
			}

			if printer.Mode == output.ModeStyled {
				printer.Table(rows)
			}

			printer.Print(output.Response{
				Data:    statuses,
				Summary: fmt.Sprintf("%d projects", len(projects)),
			})

			return nil
		},
	}
}

func statusSymbol(status string) string {
	switch status {
	case "ok":
		return "\033[32m✓\033[0m"
	case "missing":
		return "\033[33m⚠\033[0m"
	case "unlinked":
		return "\033[31m✗\033[0m"
	default:
		return "?"
	}
}
