package main

import (
	"fmt"
	"os"

	"github.com/getctx/hush/internal/config"
	"github.com/getctx/hush/internal/gitops"
	"github.com/getctx/hush/internal/output"
	"github.com/getctx/hush/internal/project"
	hsync "github.com/getctx/hush/internal/sync"
	"github.com/spf13/cobra"
)

func newLinkCmd() *cobra.Command {
	var id string

	cmd := &cobra.Command{
		Use:   "link",
		Short: "Register current directory as a managed project",
		Long:  "Detects git remote, generates a project ID, and sets up AGENTS.md, CLAUDE.md, and .gitignore.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireInit(); err != nil {
				printer.PrintError(err.Error())
				return err
			}

			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("get working directory: %w", err)
			}

			det, err := project.Detect(cwd)
			if err != nil {
				printer.PrintError(err.Error())
				return err
			}

			// Generate ID if not provided.
			if id == "" {
				id = gitops.GenerateProjectIDWithSubdir(det.Remote, det.Subdir)
			}

			printer.Debug("detected remote=%s subdir=%s → id=%s", det.Remote, det.Subdir, id)

			cfg, err := config.Load()
			if err != nil {
				printer.PrintError(fmt.Sprintf("load config: %v", err))
				return err
			}

			hostname, _ := os.Hostname()

			if flagDryRun {
				printer.Info("would register project %q (remote: %s)", id, det.Remote)
				return nil
			}

			if err := project.Register(cfg, det, id, hostname); err != nil {
				printer.PrintError(err.Error())
				return err
			}

			if err := cfg.Save(); err != nil {
				return fmt.Errorf("save config: %w", err)
			}

			// Scaffold project files.
			r, err := hsync.ScaffoldProject(det.FullPath, det.GitRoot, flagDryRun)
			if err != nil {
				printer.Warn("scaffold error: %v", err)
			}

			for _, w := range r.Warnings {
				printer.Warn("%s", w)
			}

			// If override already exists, push to private repo.
			entry := cfg.FindProject(id)
			if entry != nil {
				hsync.SyncFromProject(entry)
			}

			printer.Print(output.Response{
				Data: map[string]any{
					"id":      id,
					"remote":  det.Remote,
					"subdir":  det.Subdir,
					"path":    det.FullPath,
					"changes": r.Changes,
				},
				Summary: fmt.Sprintf("Linked %s → %s", id, det.Remote),
				Breadcrumbs: []output.Breadcrumb{
					{Action: "edit", Cmd: "hush edit " + id, Description: "Edit private rules"},
					{Action: "sync", Cmd: "hush sync", Description: "Sync from private repo"},
					{Action: "push", Cmd: "hush push", Description: "Push changes to private repo"},
				},
			})

			return nil
		},
	}

	cmd.Flags().StringVar(&id, "id", "", "Custom project ID (default: auto-generated from remote)")
	return cmd
}
