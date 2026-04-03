package main

import (
	"fmt"
	"os"

	"github.com/getctx/hush/internal/config"
	"github.com/getctx/hush/internal/gitops"
	"github.com/getctx/hush/internal/output"
	hsync "github.com/getctx/hush/internal/sync"
	"github.com/spf13/cobra"
)

func newSyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync [project]",
		Short: "Sync private rules from repo to projects",
		Long:  "Pulls the private repo and distributes AGENTS.override.md to registered project directories.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoDir := config.RepoDir()
			if _, err := os.Stat(repoDir); err != nil {
				printer.PrintError("not initialized — run 'hush init <repo-url>' first")
				return fmt.Errorf("not initialized")
			}

			// Pull latest.
			printer.Debug("pulling private repo")
			if err := gitops.Pull(repoDir); err != nil {
				printer.Warn("pull failed (using local state): %v", err)
			}

			cfg, err := config.Load()
			if err != nil {
				printer.PrintError(fmt.Sprintf("load config: %v", err))
				return err
			}

			hostname, _ := os.Hostname()
			projects := cfg.ProjectsForHost(hostname)

			// Filter to specific project if arg given.
			if len(args) > 0 {
				p := cfg.FindProject(args[0])
				if p == nil {
					printer.PrintError(fmt.Sprintf("project %q not found", args[0]))
					return fmt.Errorf("project not found")
				}
				projects = []config.ProjectEntry{*p}
			}

			var synced int
			var results []map[string]any

			for _, p := range projects {
				if _, err := os.Stat(p.Path); err != nil {
					printer.Warn("skipping %s: path not found (%s)", p.ID, p.Path)
					continue
				}

				result, err := hsync.SyncToProject(&p)
				if err != nil {
					printer.Warn("sync %s: %v", p.ID, err)
					continue
				}

				// Run scaffold checks — .gitignore must go at git root, not subdir.
				gitRoot, err := gitops.RepoRoot(p.Path)
				if err != nil {
					gitRoot = p.Path // fallback
				}
				sr, _ := hsync.ScaffoldProject(p.Path, gitRoot, flagDryRun)
				for _, w := range sr.Warnings {
					printer.Warn("%s: %s", p.ID, w)
				}

				if result.Changed || len(sr.Changes) > 0 {
					synced++
					printer.Info("%s synced (%d files)", p.ID, len(result.Files))
				} else {
					printer.Debug("%s: up to date", p.ID)
				}

				results = append(results, map[string]any{
					"id":      p.ID,
					"changed": result.Changed,
					"files":   result.Files,
				})
			}

			printer.Print(output.Response{
				Data:    results,
				Summary: fmt.Sprintf("%d of %d projects synced", synced, len(projects)),
				Breadcrumbs: []output.Breadcrumb{
					{Action: "status", Cmd: "hush status", Description: "View sync status"},
					{Action: "push", Cmd: "hush push", Description: "Push local changes to repo"},
				},
			})

			return nil
		},
	}
}
