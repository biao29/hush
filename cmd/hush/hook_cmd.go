package main

import (
	"fmt"
	"os"

	"github.com/getctx/hush/internal/gitops"
	"github.com/getctx/hush/internal/output"
	hsync "github.com/getctx/hush/internal/sync"
	"github.com/spf13/cobra"
)

func newHookCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hook",
		Short: "Manage git hooks for hush",
	}

	cmd.AddCommand(newHookInstallCmd())
	return cmd
}

func newHookInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Ensure .gitignore includes AGENTS.override.md",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}

			if !gitops.IsGitRepo(cwd) {
				printer.PrintError("not in a git repository")
				return fmt.Errorf("not a git repo")
			}

			gitRoot, err := gitops.RepoRoot(cwd)
			if err != nil {
				return err
			}

			result, err := hsync.EnsureGitignore(gitRoot, flagDryRun)
			if err != nil {
				printer.PrintError(err.Error())
				return err
			}

			if len(result.Changes) == 0 {
				printer.Print(output.Response{
					Summary: ".gitignore already includes AGENTS.override.md",
				})
			} else {
				printer.Print(output.Response{
					Data:    map[string]any{"changes": result.Changes},
					Summary: "Updated .gitignore",
				})
			}

			return nil
		},
	}
}
