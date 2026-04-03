package main

import (
	"fmt"
	"os"

	"github.com/getctx/hush/internal/config"
	"github.com/getctx/hush/internal/gitops"
	"github.com/getctx/hush/internal/output"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "init <repo-url>",
		Short: "Initialize hush with a private Git repo",
		Long:  "Clones your private rules repo to ~/.hush/repo/. If the repo is empty, creates the initial structure.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoURL := args[0]
			repoDir := config.RepoDir()

			if _, err := os.Stat(repoDir); err == nil {
				if !force {
					printer.PrintError(fmt.Sprintf("already initialized at %s — use --force to re-clone", repoDir))
					return fmt.Errorf("already initialized")
				}
				if flagDryRun {
					printer.Info("would remove %s", repoDir)
				} else {
					os.RemoveAll(repoDir)
				}
			}

			if flagDryRun {
				printer.Info("would clone %s → %s", repoURL, repoDir)
				return nil
			}

			// Ensure parent directory exists.
			os.MkdirAll(config.HomeDir(), 0o755)

			printer.Debug("cloning %s → %s", repoURL, repoDir)
			if err := gitops.Clone(repoURL, repoDir); err != nil {
				printer.PrintError(fmt.Sprintf("clone failed: %v", err))
				return err
			}

			// If repo is empty (no hush.yaml), create initial structure.
			if _, err := os.Stat(config.ConfigPath()); os.IsNotExist(err) {
				printer.Debug("empty repo — creating initial structure")
				os.MkdirAll(config.ProjectsDir(), 0o755)
				cfg := &config.HushConfig{}
				if err := cfg.Save(); err != nil {
					return fmt.Errorf("create hush.yaml: %w", err)
				}
				// Configure git user for the repo.
				gitops.Run(repoDir, "config", "user.email", "hush@local")
				gitops.Run(repoDir, "config", "user.name", "hush")

				if err := gitops.CommitAll(repoDir, "hush: initial structure"); err != nil {
					printer.Warn("could not commit initial structure: %v", err)
				} else {
					if err := gitops.Push(repoDir); err != nil {
						printer.Warn("committed locally but could not push: %v", err)
					}
				}
			}

			// Count projects.
			cfg, _ := config.Load()
			count := len(cfg.Projects)

			printer.Print(output.Response{
				Data:    map[string]any{"repo": repoDir, "projects": count},
				Summary: fmt.Sprintf("Initialized hush at %s (%d projects)", repoDir, count),
				Breadcrumbs: []output.Breadcrumb{
					{Action: "link", Cmd: "hush link", Description: "Register a project"},
					{Action: "status", Cmd: "hush status", Description: "View all projects"},
				},
			})

			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Re-clone even if already initialized")
	return cmd
}
