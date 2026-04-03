package main

import (
	"fmt"

	"github.com/getctx/hush/internal/doctor"
	"github.com/getctx/hush/internal/output"
	"github.com/spf13/cobra"
)

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose hush configuration issues",
		RunE: func(cmd *cobra.Command, args []string) error {
			checks := doctor.RunChecks()

			var rows [][]string
			var results []map[string]any
			var issues int

			for _, c := range checks {
				icon := checkIcon(c.Status)
				rows = append(rows, []string{icon, c.Name, c.Detail})
				results = append(results, map[string]any{
					"name":   c.Name,
					"status": c.Status,
					"detail": c.Detail,
				})
				if c.Status != "ok" {
					issues++
				}
			}

			if printer.Mode == output.ModeStyled {
				printer.Table(rows)
				fmt.Fprintln(cmd.OutOrStdout())
			}

			summary := fmt.Sprintf("%d checks, %d issues", len(checks), issues)
			printer.Print(output.Response{
				Data:    results,
				Summary: summary,
			})

			return nil
		},
	}
}

func checkIcon(status string) string {
	switch status {
	case "ok":
		return "\033[32m✓\033[0m"
	case "warn":
		return "\033[33m⚠\033[0m"
	case "fail":
		return "\033[31m✗\033[0m"
	default:
		return "?"
	}
}
