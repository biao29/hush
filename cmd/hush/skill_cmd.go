package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/getctx/hush/internal/output"
	"github.com/getctx/hush/internal/skills"
	"github.com/spf13/cobra"
)

const skillFilename = "SKILL.md"

// skillLocation represents a predefined skill installation target.
type skillLocation struct {
	Name string
	Path string
}

func defaultSkillLocations() []skillLocation {
	codexHome := strings.TrimSpace(os.Getenv("CODEX_HOME"))
	codexPath := "~/.codex/skills/hush/SKILL.md"
	if codexHome != "" {
		codexPath = filepath.Join(codexHome, "skills", "hush", skillFilename)
	}

	return []skillLocation{
		{Name: "Agents (Global)", Path: "~/.agents/skills/hush/SKILL.md"},
		{Name: "Agents (Project)", Path: ".agents/skills/hush/SKILL.md"},
		{Name: "Claude Code (Global)", Path: "~/.claude/skills/hush/SKILL.md"},
		{Name: "Claude Code (Project)", Path: ".claude/skills/hush/SKILL.md"},
		{Name: "OpenCode (Global)", Path: "~/.config/opencode/skill/hush/SKILL.md"},
		{Name: "OpenCode (Project)", Path: ".opencode/skill/hush/SKILL.md"},
		{Name: "Codex (Global)", Path: codexPath},
	}
}

func newSkillCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skill",
		Short: "Manage the embedded agent skill",
		Long:  "Print or install the embedded SKILL.md that teaches AI agents how to use hush.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Default: print skill content to stdout.
			_, err := fmt.Fprint(cmd.OutOrStdout(), string(skills.Content))
			return err
		},
	}

	cmd.AddCommand(newSkillInstallCmd())
	cmd.AddCommand(newSkillPathCmd())
	return cmd
}

func newSkillInstallCmd() *cobra.Command {
	var target string

	cmd := &cobra.Command{
		Use:   "install [--target <name|path>]",
		Short: "Install the hush agent skill to a standard location",
		Long: `Copies the embedded SKILL.md to the specified agent skill directory.

Targets:
  agents        ~/.agents/skills/hush/ (default)
  claude        ~/.claude/skills/hush/
  opencode      ~/.config/opencode/skill/hush/
  codex         ~/.codex/skills/hush/
  all           All of the above
  <path>        Custom directory or file path`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if target == "" {
				target = "agents"
			}

			locations := resolveInstallTargets(target)
			if len(locations) == 0 {
				// Treat as custom path.
				locations = []skillLocation{{Name: "Custom", Path: normalizeSkillPath(target)}}
			}

			var installed []string
			for _, loc := range locations {
				expanded := expandTilde(loc.Path)

				if flagDryRun {
					printer.Info("would install to %s (%s)", loc.Name, expanded)
					continue
				}

				dir := filepath.Dir(expanded)
				if err := os.MkdirAll(dir, 0o755); err != nil {
					printer.Warn("skip %s: %v", loc.Name, err)
					continue
				}
				if err := os.WriteFile(expanded, skills.Content, 0o644); err != nil {
					printer.Warn("skip %s: %v", loc.Name, err)
					continue
				}

				installed = append(installed, expanded)
				printer.Info("installed to %s", expanded)
			}

			printer.Print(output.Response{
				Data:    map[string]any{"installed": installed},
				Summary: fmt.Sprintf("Skill installed to %d locations", len(installed)),
				Breadcrumbs: []output.Breadcrumb{
					{Action: "verify", Cmd: "hush skill path", Description: "Show installed skill locations"},
				},
			})

			return nil
		},
	}

	cmd.Flags().StringVar(&target, "target", "", "Installation target (agents, claude, opencode, codex, all, or custom path)")
	return cmd
}

func newSkillPathCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Show standard skill installation locations",
		RunE: func(cmd *cobra.Command, args []string) error {
			locations := defaultSkillLocations()
			var rows [][]string
			var data []map[string]any

			for _, loc := range locations {
				expanded := expandTilde(loc.Path)
				status := "\033[90m-\033[0m"
				exists := false
				if _, err := os.Stat(expanded); err == nil {
					status = "\033[32m✓\033[0m"
					exists = true
				}
				rows = append(rows, []string{status, loc.Name, loc.Path})
				data = append(data, map[string]any{
					"name":      loc.Name,
					"path":      loc.Path,
					"installed": exists,
				})
			}

			if printer.Mode == output.ModeStyled {
				printer.Table(rows)
			}

			printer.Print(output.Response{
				Data:    data,
				Summary: fmt.Sprintf("%d known locations", len(locations)),
				Breadcrumbs: []output.Breadcrumb{
					{Action: "install", Cmd: "hush skill install --target all", Description: "Install to all locations"},
				},
			})
			return nil
		},
	}
}

// resolveInstallTargets maps a target name to skill locations.
func resolveInstallTargets(target string) []skillLocation {
	all := defaultSkillLocations()

	switch strings.ToLower(target) {
	case "all":
		return all
	case "agents":
		return filterLocations(all, "Agents (Global)")
	case "claude":
		return filterLocations(all, "Claude Code (Global)")
	case "opencode":
		return filterLocations(all, "OpenCode (Global)")
	case "codex":
		return filterLocations(all, "Codex (Global)")
	case "agents-project":
		return filterLocations(all, "Agents (Project)")
	case "claude-project":
		return filterLocations(all, "Claude Code (Project)")
	case "opencode-project":
		return filterLocations(all, "OpenCode (Project)")
	default:
		return nil // custom path
	}
}

func filterLocations(locs []skillLocation, name string) []skillLocation {
	for _, l := range locs {
		if l.Name == name {
			return []skillLocation{l}
		}
	}
	return nil
}

// normalizeSkillPath appends hush/SKILL.md to directory paths.
func normalizeSkillPath(path string) string {
	path = strings.TrimSpace(path)
	if strings.HasSuffix(strings.ToLower(path), ".md") {
		return path
	}
	if strings.HasSuffix(path, "hush") || strings.HasSuffix(path, "hush/") ||
		strings.HasSuffix(path, "hush\\") {
		return filepath.Join(path, skillFilename)
	}
	return filepath.Join(path, "hush", skillFilename)
}

// expandTilde expands ~ to user home directory.
func expandTilde(path string) string {
	if strings.HasPrefix(path, "~/") || strings.HasPrefix(path, "~\\") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(home, path[2:])
	}
	return path
}
