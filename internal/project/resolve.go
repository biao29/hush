package project

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/getctx/hush/internal/config"
)

// Resolve finds a project entry from a name argument or CWD.
// If nameOrEmpty is provided, looks up by project ID.
// If empty, detects from the current working directory using remote + subdir matching.
func Resolve(cfg *config.HushConfig, nameOrEmpty string) (*config.ProjectEntry, error) {
	if nameOrEmpty != "" {
		p := cfg.FindProject(nameOrEmpty)
		if p == nil {
			return nil, fmt.Errorf("project %q not found — run 'hush link' in the project directory", nameOrEmpty)
		}
		return p, nil
	}

	// Detect from CWD with full remote + subdir matching.
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get working directory: %w", err)
	}

	det, err := Detect(cwd)
	if err != nil {
		return nil, fmt.Errorf("detect project: %w", err)
	}

	// First try exact match: remote + subdir.
	if p := cfg.FindByRemoteAndSubdir(det.Remote, det.Subdir); p != nil {
		return p, nil
	}

	// Walk up from CWD to git root trying each subdir prefix.
	// This handles being inside a deeper subdirectory of a registered project.
	rootResolved, _ := filepath.EvalSymlinks(det.GitRoot)
	cwdResolved, _ := filepath.EvalSymlinks(cwd)
	current := cwdResolved

	for current != rootResolved && len(current) > len(rootResolved) {
		current = filepath.Dir(current)
		rel, err := filepath.Rel(rootResolved, current)
		if err != nil {
			break
		}
		sub := filepath.ToSlash(rel)
		if sub == "." {
			sub = ""
		}
		if p := cfg.FindByRemoteAndSubdir(det.Remote, sub); p != nil {
			return p, nil
		}
	}

	// Last resort: try remote with empty subdir (root-level registration).
	if p := cfg.FindByRemoteAndSubdir(det.Remote, ""); p != nil {
		return p, nil
	}

	return nil, fmt.Errorf("no registered project for remote %q (subdir: %q) — run 'hush link' first", det.Remote, det.Subdir)
}
