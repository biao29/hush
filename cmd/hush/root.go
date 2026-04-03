package main

import (
	"fmt"

	"github.com/getctx/hush/internal/config"
	"github.com/getctx/hush/internal/output"
	"github.com/spf13/cobra"
)

var (
	flagJSON    bool
	flagQuiet   bool
	flagAgent   bool
	flagVerbose bool
	flagDryRun  bool
	printer     *output.Printer
)

func newRootCmd(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hush",
		Short: "Dotfiles for AI agent rules",
		Long:  "hush syncs private AGENTS.override.md files across projects and devices via a personal Git repo.",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			printer = output.DefaultPrinter()
			printer.Mode = output.ParseMode(flagJSON, flagQuiet, flagAgent)
			printer.Verbose = flagVerbose
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	pf := cmd.PersistentFlags()
	pf.BoolVar(&flagJSON, "json", false, "JSON envelope output")
	pf.BoolVar(&flagQuiet, "quiet", false, "Minimal output")
	pf.BoolVar(&flagAgent, "agent", false, "Agent-optimized output (quiet + breadcrumbs)")
	pf.BoolVar(&flagVerbose, "verbose", false, "Show debug details")
	pf.BoolVar(&flagDryRun, "dry-run", false, "Show what would change without writing")

	cmd.AddCommand(
		newVersionCmd(version),
		newInitCmd(),
		newLinkCmd(),
		newSyncCmd(),
		newPushCmd(),
		newEditCmd(),
		newHookCmd(),
		newStatusCmd(),
		newDoctorCmd(),
		newSkillCmd(),
	)

	return cmd
}

// requireInit checks that hush is initialized and returns a helpful error if not.
func requireInit() error {
	if !config.IsInitialized() {
		return fmt.Errorf("hush is not initialized — run 'hush init <repo-url>' first")
	}
	return nil
}
