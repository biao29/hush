// Package sync provides the core sync engine and project scaffolding.
package sync

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	agentsMDFile    = "AGENTS.md"
	claudeMDFile    = "CLAUDE.md"
	overrideFile    = "AGENTS.override.md"
	gitignoreFile   = ".gitignore"
	includeDirective = "@AGENTS.override.md"
	claudeInclude    = "@AGENTS.md"
)

// ScaffoldResult tracks what the scaffold changed.
type ScaffoldResult struct {
	Changes  []string
	Warnings []string
}

// EnsureAgentsMD ensures AGENTS.md ends with @AGENTS.override.md.
func EnsureAgentsMD(projectDir string, dryRun bool) (*ScaffoldResult, error) {
	result := &ScaffoldResult{}
	path := filepath.Join(projectDir, agentsMDFile)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			if dryRun {
				result.Changes = append(result.Changes, "would create "+agentsMDFile)
				return result, nil
			}
			content := "# Project Rules\n\n" + includeDirective + "\n"
			if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
				return nil, fmt.Errorf("create %s: %w", agentsMDFile, err)
			}
			result.Changes = append(result.Changes, "created "+agentsMDFile)
			return result, nil
		}
		return nil, fmt.Errorf("read %s: %w", agentsMDFile, err)
	}

	content := string(data)
	if strings.Contains(content, includeDirective) {
		return result, nil
	}

	if dryRun {
		result.Changes = append(result.Changes, "would append "+includeDirective+" to "+agentsMDFile)
		return result, nil
	}

	// Append the include directive.
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	content += "\n" + includeDirective + "\n"

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return nil, fmt.Errorf("update %s: %w", agentsMDFile, err)
	}
	result.Changes = append(result.Changes, "appended "+includeDirective+" to "+agentsMDFile)
	result.Warnings = append(result.Warnings, agentsMDFile+" was modified — review the change")
	return result, nil
}

// EnsureClaudeMD ensures CLAUDE.md exists and contains @AGENTS.md.
func EnsureClaudeMD(projectDir string, dryRun bool) (*ScaffoldResult, error) {
	result := &ScaffoldResult{}
	path := filepath.Join(projectDir, claudeMDFile)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			if dryRun {
				result.Changes = append(result.Changes, "would create "+claudeMDFile)
				return result, nil
			}
			if err := os.WriteFile(path, []byte(claudeInclude+"\n"), 0o644); err != nil {
				return nil, fmt.Errorf("create %s: %w", claudeMDFile, err)
			}
			result.Changes = append(result.Changes, "created "+claudeMDFile+" with "+claudeInclude)
			return result, nil
		}
		return nil, fmt.Errorf("read %s: %w", claudeMDFile, err)
	}

	content := string(data)
	if strings.Contains(content, claudeInclude) {
		return result, nil
	}

	if dryRun {
		result.Changes = append(result.Changes, "would prepend "+claudeInclude+" to "+claudeMDFile)
		return result, nil
	}

	// Prepend @AGENTS.md to existing content.
	content = claudeInclude + "\n" + content
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return nil, fmt.Errorf("update %s: %w", claudeMDFile, err)
	}
	result.Changes = append(result.Changes, "prepended "+claudeInclude+" to "+claudeMDFile)
	result.Warnings = append(result.Warnings, claudeMDFile+" was modified — review the change")
	return result, nil
}

// EnsureGitignore ensures .gitignore includes AGENTS.override.md.
func EnsureGitignore(gitRoot string, dryRun bool) (*ScaffoldResult, error) {
	result := &ScaffoldResult{}
	path := filepath.Join(gitRoot, gitignoreFile)

	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("read %s: %w", gitignoreFile, err)
	}

	content := string(data)
	if strings.Contains(content, overrideFile) {
		return result, nil
	}

	if dryRun {
		result.Changes = append(result.Changes, "would add "+overrideFile+" to "+gitignoreFile)
		return result, nil
	}

	if !strings.HasSuffix(content, "\n") && content != "" {
		content += "\n"
	}
	content += "\n# hush — private agent rules\n" + overrideFile + "\n"

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return nil, fmt.Errorf("update %s: %w", gitignoreFile, err)
	}
	result.Changes = append(result.Changes, "added "+overrideFile+" to "+gitignoreFile)
	return result, nil
}

// EnsureOverride creates an empty AGENTS.override.md if it doesn't exist.
func EnsureOverride(projectDir string, dryRun bool) (*ScaffoldResult, error) {
	result := &ScaffoldResult{}
	path := filepath.Join(projectDir, overrideFile)

	if _, err := os.Stat(path); err == nil {
		return result, nil
	}

	if dryRun {
		result.Changes = append(result.Changes, "would create "+overrideFile)
		return result, nil
	}

	content := "# Private Rules\n#\n# This file is managed by hush and gitignored.\n# Edit with: hush edit\n"
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return nil, fmt.Errorf("create %s: %w", overrideFile, err)
	}
	result.Changes = append(result.Changes, "created "+overrideFile)
	return result, nil
}

// ScaffoldProject runs all scaffold checks on a project directory.
func ScaffoldProject(projectDir, gitRoot string, dryRun bool) (*ScaffoldResult, error) {
	combined := &ScaffoldResult{}

	checks := []func(string, bool) (*ScaffoldResult, error){
		func(dir string, dry bool) (*ScaffoldResult, error) { return EnsureAgentsMD(dir, dry) },
		func(dir string, dry bool) (*ScaffoldResult, error) { return EnsureClaudeMD(dir, dry) },
		func(dir string, dry bool) (*ScaffoldResult, error) { return EnsureOverride(dir, dry) },
	}

	for _, check := range checks {
		r, err := check(projectDir, dryRun)
		if err != nil {
			return nil, err
		}
		combined.Changes = append(combined.Changes, r.Changes...)
		combined.Warnings = append(combined.Warnings, r.Warnings...)
	}

	// Gitignore is at git root, not project dir (for monorepos).
	r, err := EnsureGitignore(gitRoot, dryRun)
	if err != nil {
		return nil, err
	}
	combined.Changes = append(combined.Changes, r.Changes...)
	combined.Warnings = append(combined.Warnings, r.Warnings...)

	return combined, nil
}
