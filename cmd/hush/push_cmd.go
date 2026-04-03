package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/getctx/hush/internal/config"
	"github.com/getctx/hush/internal/gitops"
	"github.com/getctx/hush/internal/output"
	hsync "github.com/getctx/hush/internal/sync"
	"github.com/spf13/cobra"
)

func newPushCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "push [project]",
		Short: "Push project overrides to the private repo",
		Long:  "Copies AGENTS.override.md from project directories to the private repo, commits, and pushes.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoDir := config.RepoDir()
			if _, err := os.Stat(repoDir); err != nil {
				printer.PrintError("not initialized — run 'hush init <repo-url>' first")
				return fmt.Errorf("not initialized")
			}

			cfg, err := config.Load()
			if err != nil {
				printer.PrintError(fmt.Sprintf("load config: %v", err))
				return err
			}

			hostname, _ := os.Hostname()
			projects := cfg.ProjectsForHost(hostname)

			if len(args) > 0 {
				p := cfg.FindProject(args[0])
				if p == nil {
					printer.PrintError(fmt.Sprintf("project %q not found", args[0]))
					return fmt.Errorf("project not found")
				}
				projects = []config.ProjectEntry{*p}
			}

			var changedIDs []string

			for _, p := range projects {
				if _, err := os.Stat(p.Path); err != nil {
					printer.Warn("skipping %s: path not found", p.ID)
					continue
				}

				result, err := hsync.SyncFromProject(&p)
				if err != nil {
					printer.Warn("push %s: %v", p.ID, err)
					continue
				}

				if result.Changed {
					changedIDs = append(changedIDs, p.ID)
					printer.Info("%s: %d files collected", p.ID, len(result.Files))
				} else {
					printer.Debug("%s: up to date", p.ID)
				}
			}

			if len(changedIDs) == 0 {
				printer.Print(output.Response{
					Summary: "Already up to date",
				})
				return nil
			}

			if flagDryRun {
				printer.Print(output.Response{
					Summary: fmt.Sprintf("Would commit and push changes for: %s", strings.Join(changedIDs, ", ")),
				})
				return nil
			}

			// Commit and push.
			commitMsg := fmt.Sprintf("hush: update %s", strings.Join(changedIDs, ", "))
			if err := gitops.CommitAll(repoDir, commitMsg); err != nil {
				printer.PrintError(fmt.Sprintf("commit: %v", err))
				return err
			}

			if err := gitops.Push(repoDir); err != nil {
				printer.Warn("committed locally but push failed: %v", err)
				printer.Print(output.Response{
					Summary: fmt.Sprintf("Committed %d projects (push failed — retry when online)", len(changedIDs)),
				})
				return nil
			}

			printer.Print(output.Response{
				Data: map[string]any{
					"projects": changedIDs,
				},
				Summary: fmt.Sprintf("Pushed %d projects: %s", len(changedIDs), strings.Join(changedIDs, ", ")),
				Breadcrumbs: []output.Breadcrumb{
					{Action: "sync", Cmd: "hush sync", Description: "Sync to all projects"},
					{Action: "status", Cmd: "hush status", Description: "View sync status"},
				},
			})

			return nil
		},
	}
}
