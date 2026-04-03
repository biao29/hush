// Package config handles hush configuration loading and path resolution.
package config

import (
	"os"
	"path/filepath"
)

// HomeDir returns the hush home directory (~/.hush or $HUSH_HOME).
func HomeDir() string {
	if v := os.Getenv("HUSH_HOME"); v != "" {
		return v
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".hush")
	}
	return filepath.Join(home, ".hush")
}

// RepoDir returns the path to the cloned private repo.
func RepoDir() string {
	return filepath.Join(HomeDir(), "repo")
}

// ConfigPath returns the path to hush.yaml inside the private repo.
func ConfigPath() string {
	return filepath.Join(RepoDir(), "hush.yaml")
}

// ProjectsDir returns the projects directory inside the private repo.
func ProjectsDir() string {
	return filepath.Join(RepoDir(), "projects")
}

// ProjectDir returns the directory for a specific project in the private repo.
func ProjectDir(id string) string {
	return filepath.Join(ProjectsDir(), id)
}

// StateFile returns the path to the local state file.
func StateFile() string {
	return filepath.Join(HomeDir(), "state.json")
}
