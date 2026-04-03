package project

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/getctx/hush/internal/config"
	"github.com/getctx/hush/internal/gitops"
)

func TestResolveByName(t *testing.T) {
	cfg := &config.HushConfig{
		Projects: []config.ProjectEntry{
			{ID: "my-proj", Remote: "github.com/org/repo"},
		},
	}

	p, err := Resolve(cfg, "my-proj")
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if p.ID != "my-proj" {
		t.Errorf("ID = %q", p.ID)
	}
}

func TestResolveByNameMissing(t *testing.T) {
	cfg := &config.HushConfig{}
	_, err := Resolve(cfg, "missing")
	if err == nil {
		t.Error("expected error for missing project")
	}
}

// TestResolveMonorepoSubdir verifies that CWD-based resolution picks the
// correct project entry when multiple subdir projects share the same remote.
func TestResolveMonorepoSubdir(t *testing.T) {
	// Create a git repo simulating a monorepo.
	root := t.TempDir()
	if _, err := gitops.Run(root, "init"); err != nil {
		t.Skipf("git not available: %v", err)
	}
	gitops.Run(root, "config", "user.email", "test@test.com")
	gitops.Run(root, "config", "user.name", "Test")
	gitops.Run(root, "remote", "add", "origin", "git@github.com:org/mono.git")

	apiDir := filepath.Join(root, "services", "api")
	webDir := filepath.Join(root, "services", "web")
	os.MkdirAll(apiDir, 0o755)
	os.MkdirAll(webDir, 0o755)

	cfg := &config.HushConfig{
		Projects: []config.ProjectEntry{
			{ID: "mono-api", Remote: "github.com/org/mono", Subdir: "services/api", Path: apiDir},
			{ID: "mono-web", Remote: "github.com/org/mono", Subdir: "services/web", Path: webDir},
		},
	}

	// Resolve from web directory — should NOT pick api.
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)

	os.Chdir(webDir)
	p, err := Resolve(cfg, "")
	if err != nil {
		t.Fatalf("Resolve from web: %v", err)
	}
	if p.ID != "mono-web" {
		t.Errorf("from web dir: got %q, want %q", p.ID, "mono-web")
	}

	// Resolve from api directory.
	os.Chdir(apiDir)
	p, err = Resolve(cfg, "")
	if err != nil {
		t.Fatalf("Resolve from api: %v", err)
	}
	if p.ID != "mono-api" {
		t.Errorf("from api dir: got %q, want %q", p.ID, "mono-api")
	}

	// Resolve from a deeper nested directory inside api.
	deepDir := filepath.Join(apiDir, "internal", "handlers")
	os.MkdirAll(deepDir, 0o755)
	os.Chdir(deepDir)
	p, err = Resolve(cfg, "")
	if err != nil {
		t.Fatalf("Resolve from deep: %v", err)
	}
	if p.ID != "mono-api" {
		t.Errorf("from deep dir: got %q, want %q", p.ID, "mono-api")
	}
}
