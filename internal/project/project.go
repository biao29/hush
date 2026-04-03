// Package project handles project detection and registration.
package project

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/getctx/hush/internal/config"
	"github.com/getctx/hush/internal/gitops"
)

// DetectedProject contains information about a project detected from a directory.
type DetectedProject struct {
	GitRoot  string // absolute path to git repo root
	Remote   string // normalized remote URL
	Subdir   string // relative path from git root (empty if at root)
	FullPath string // absolute path to the project directory
}

// Detect inspects a directory and returns project metadata.
func Detect(dir string) (*DetectedProject, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("resolve path: %w", err)
	}

	if !gitops.IsGitRepo(absDir) {
		return nil, fmt.Errorf("%s is not a git repository", absDir)
	}

	root, err := gitops.RepoRoot(absDir)
	if err != nil {
		return nil, fmt.Errorf("find repo root: %w", err)
	}

	rawURL, err := gitops.RemoteURL(absDir)
	if err != nil {
		return nil, fmt.Errorf("get remote URL: %w", err)
	}

	remote := gitops.NormalizeURL(rawURL)

	// Compute subdir relative to git root.
	subdir := ""
	rootResolved, _ := filepath.EvalSymlinks(root)
	dirResolved, _ := filepath.EvalSymlinks(absDir)
	if rootResolved != dirResolved {
		rel, err := filepath.Rel(rootResolved, dirResolved)
		if err == nil && rel != "." {
			subdir = filepath.ToSlash(rel)
		}
	}

	return &DetectedProject{
		GitRoot:  root,
		Remote:   remote,
		Subdir:   subdir,
		FullPath: absDir,
	}, nil
}

// Register adds a detected project to the config and creates the repo directory.
func Register(cfg *config.HushConfig, det *DetectedProject, id string, hostname string) error {
	entry := config.ProjectEntry{
		ID:       id,
		Remote:   det.Remote,
		Path:     det.FullPath,
		Subdir:   det.Subdir,
		Hostname: hostname,
	}

	if err := cfg.AddProject(entry); err != nil {
		return err
	}

	// Create project directory in private repo.
	projDir := config.ProjectDir(id)
	if err := os.MkdirAll(projDir, 0o755); err != nil {
		return fmt.Errorf("create project dir: %w", err)
	}

	return nil
}
