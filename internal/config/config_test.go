package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "hush.yaml")

	cfg := &HushConfig{
		Projects: []ProjectEntry{
			{
				ID:       "my-project",
				Remote:   "github.com/user/repo",
				Path:     "/home/user/repo",
				Subdir:   "services/api",
				Hostname: "dev-box",
			},
		},
	}

	if err := cfg.SaveTo(path); err != nil {
		t.Fatalf("SaveTo: %v", err)
	}

	loaded, err := LoadFrom(path)
	if err != nil {
		t.Fatalf("LoadFrom: %v", err)
	}

	if len(loaded.Projects) != 1 {
		t.Fatalf("projects count = %d, want 1", len(loaded.Projects))
	}

	p := loaded.Projects[0]
	if p.ID != "my-project" {
		t.Errorf("ID = %q, want %q", p.ID, "my-project")
	}
	if p.Remote != "github.com/user/repo" {
		t.Errorf("Remote = %q", p.Remote)
	}
	if p.Subdir != "services/api" {
		t.Errorf("Subdir = %q", p.Subdir)
	}
	if p.Hostname != "dev-box" {
		t.Errorf("Hostname = %q", p.Hostname)
	}
}

func TestLoadMissing(t *testing.T) {
	cfg, err := LoadFrom("/nonexistent/hush.yaml")
	if err != nil {
		t.Fatalf("LoadFrom missing: %v", err)
	}
	if len(cfg.Projects) != 0 {
		t.Error("expected empty config for missing file")
	}
}

func TestFindProject(t *testing.T) {
	cfg := &HushConfig{
		Projects: []ProjectEntry{
			{ID: "a", Remote: "github.com/a/a"},
			{ID: "b", Remote: "github.com/b/b"},
		},
	}

	if p := cfg.FindProject("a"); p == nil || p.Remote != "github.com/a/a" {
		t.Error("FindProject(a) failed")
	}
	if p := cfg.FindProject("missing"); p != nil {
		t.Error("FindProject(missing) should return nil")
	}
}

func TestFindByRemoteAndSubdir(t *testing.T) {
	cfg := &HushConfig{
		Projects: []ProjectEntry{
			{ID: "mono-api", Remote: "github.com/org/mono", Subdir: "api"},
			{ID: "mono-web", Remote: "github.com/org/mono", Subdir: "web"},
		},
	}

	if p := cfg.FindByRemoteAndSubdir("github.com/org/mono", "api"); p == nil || p.ID != "mono-api" {
		t.Error("FindByRemoteAndSubdir(api) failed")
	}
	if p := cfg.FindByRemoteAndSubdir("github.com/org/mono", "web"); p == nil || p.ID != "mono-web" {
		t.Error("FindByRemoteAndSubdir(web) failed")
	}
	if p := cfg.FindByRemoteAndSubdir("github.com/org/mono", "other"); p != nil {
		t.Error("FindByRemoteAndSubdir(other) should return nil")
	}
}

func TestAddProjectDuplicateID(t *testing.T) {
	cfg := &HushConfig{
		Projects: []ProjectEntry{{ID: "dup", Remote: "github.com/a/a"}},
	}
	err := cfg.AddProject(ProjectEntry{ID: "dup", Remote: "github.com/b/b"})
	if err == nil {
		t.Error("expected error for duplicate ID")
	}
}

func TestAddProjectDuplicateRemote(t *testing.T) {
	cfg := &HushConfig{
		Projects: []ProjectEntry{{ID: "a", Remote: "github.com/a/a", Subdir: "sub"}},
	}
	err := cfg.AddProject(ProjectEntry{ID: "b", Remote: "github.com/a/a", Subdir: "sub"})
	if err == nil {
		t.Error("expected error for duplicate remote+subdir")
	}
}

func TestProjectsForHost(t *testing.T) {
	cfg := &HushConfig{
		Projects: []ProjectEntry{
			{ID: "global", Remote: "github.com/a/a"},
			{ID: "dev-only", Remote: "github.com/b/b", Hostname: "dev-box"},
			{ID: "prod-only", Remote: "github.com/c/c", Hostname: "prod-box"},
		},
	}

	devProjects := cfg.ProjectsForHost("dev-box")
	if len(devProjects) != 2 {
		t.Errorf("dev-box projects = %d, want 2", len(devProjects))
	}

	prodProjects := cfg.ProjectsForHost("prod-box")
	if len(prodProjects) != 2 {
		t.Errorf("prod-box projects = %d, want 2", len(prodProjects))
	}

	unknownProjects := cfg.ProjectsForHost("unknown")
	if len(unknownProjects) != 1 {
		t.Errorf("unknown host projects = %d, want 1", len(unknownProjects))
	}
}

func TestSaveCreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "hush.yaml")

	cfg := &HushConfig{}
	if err := cfg.SaveTo(path); err != nil {
		t.Fatalf("SaveTo: %v", err)
	}

	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file not created: %v", err)
	}
}
