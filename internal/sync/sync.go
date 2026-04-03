package sync

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/getctx/hush/internal/config"
)

// SyncResult tracks what a sync operation changed.
type SyncResult struct {
	ProjectID string
	Changed   bool
	Files     []string
	Warnings  []string
}

// SyncToProject copies agents.override.md from the private repo to the project directory.
// Uses content comparison for idempotency.
func SyncToProject(entry *config.ProjectEntry) (*SyncResult, error) {
	result := &SyncResult{ProjectID: entry.ID}

	srcPath := filepath.Join(config.ProjectDir(entry.ID), "agents.override.md")
	dstPath := filepath.Join(entry.Path, overrideFile)

	changed, err := copyIfDifferent(srcPath, dstPath)
	if err != nil {
		// Source doesn't exist = nothing to sync (not an error).
		if os.IsNotExist(err) {
			return result, nil
		}
		return nil, fmt.Errorf("sync override to project: %w", err)
	}

	if changed {
		result.Changed = true
		result.Files = append(result.Files, overrideFile)
	}

	return result, nil
}

// SyncFromProject copies AGENTS.override.md and snapshots AGENTS.md from project to private repo.
func SyncFromProject(entry *config.ProjectEntry) (*SyncResult, error) {
	result := &SyncResult{ProjectID: entry.ID}
	projRepoDir := config.ProjectDir(entry.ID)

	// Ensure project directory exists in private repo.
	if err := os.MkdirAll(projRepoDir, 0o755); err != nil {
		return nil, fmt.Errorf("create project dir: %w", err)
	}

	// Copy AGENTS.override.md → private repo.
	overrideSrc := filepath.Join(entry.Path, overrideFile)
	overrideDst := filepath.Join(projRepoDir, "agents.override.md")
	changed, err := copyIfDifferent(overrideSrc, overrideDst)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("sync override from project: %w", err)
	}
	if changed {
		result.Changed = true
		result.Files = append(result.Files, "agents.override.md")
	}

	// Snapshot AGENTS.md → private repo.
	agentsSrc := filepath.Join(entry.Path, agentsMDFile)
	agentsDst := filepath.Join(projRepoDir, "agents.md")
	changed, err = copyIfDifferent(agentsSrc, agentsDst)
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("snapshot agents.md: %w", err)
	}
	if changed {
		result.Changed = true
		result.Files = append(result.Files, "agents.md")
	}

	return result, nil
}

// copyIfDifferent copies src to dst only if contents differ.
// Returns true if the file was written, false if already identical.
// Returns the underlying error if src doesn't exist.
func copyIfDifferent(src, dst string) (bool, error) {
	srcData, err := os.ReadFile(src)
	if err != nil {
		return false, err
	}

	dstData, err := os.ReadFile(dst)
	if err == nil && bytes.Equal(srcData, dstData) {
		return false, nil
	}

	if err := os.WriteFile(dst, srcData, 0o644); err != nil {
		return false, fmt.Errorf("write %s: %w", dst, err)
	}
	return true, nil
}
