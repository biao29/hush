package project

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/getctx/hush/internal/config"
	"github.com/getctx/hush/internal/gitops"
)

func setupGitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if _, err := gitops.Run(dir, "init"); err != nil {
		t.Skipf("git not available: %v", err)
	}
	gitops.Run(dir, "config", "user.email", "test@test.com")
	gitops.Run(dir, "config", "user.name", "Test")
	gitops.Run(dir, "remote", "add", "origin", "git@github.com:openelf/getctx.org.git")
	return dir
}

func TestDetect(t *testing.T) {
	dir := setupGitRepo(t)

	det, err := Detect(dir)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}

	if det.Remote != "github.com/openelf/getctx.org" {
		t.Errorf("Remote = %q, want %q", det.Remote, "github.com/openelf/getctx.org")
	}
	if det.Subdir != "" {
		t.Errorf("Subdir = %q, want empty", det.Subdir)
	}
}

func TestDetectSubdir(t *testing.T) {
	dir := setupGitRepo(t)

	subdir := filepath.Join(dir, "services", "api")
	os.MkdirAll(subdir, 0o755)

	det, err := Detect(subdir)
	if err != nil {
		t.Fatalf("Detect: %v", err)
	}

	if det.Subdir != "services/api" {
		t.Errorf("Subdir = %q, want %q", det.Subdir, "services/api")
	}
}

func TestDetectNotGit(t *testing.T) {
	dir := t.TempDir()
	_, err := Detect(dir)
	if err == nil {
		t.Error("expected error for non-git directory")
	}
}

func TestRegister(t *testing.T) {
	// Setup HUSH_HOME.
	hushHome := t.TempDir()
	t.Setenv("HUSH_HOME", hushHome)

	// Create repo/projects dir.
	os.MkdirAll(config.ProjectsDir(), 0o755)

	cfg := &config.HushConfig{}
	det := &DetectedProject{
		GitRoot:  "/tmp/repo",
		Remote:   "github.com/org/repo",
		FullPath: "/tmp/repo",
	}

	err := Register(cfg, det, "org-repo", "myhost")
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	if len(cfg.Projects) != 1 {
		t.Fatalf("projects count = %d, want 1", len(cfg.Projects))
	}

	p := cfg.Projects[0]
	if p.ID != "org-repo" {
		t.Errorf("ID = %q", p.ID)
	}
	if p.Hostname != "myhost" {
		t.Errorf("Hostname = %q", p.Hostname)
	}

	// Check directory was created.
	if _, err := os.Stat(config.ProjectDir("org-repo")); err != nil {
		t.Errorf("project dir not created: %v", err)
	}
}

func TestRegisterDuplicate(t *testing.T) {
	hushHome := t.TempDir()
	t.Setenv("HUSH_HOME", hushHome)
	os.MkdirAll(config.ProjectsDir(), 0o755)

	cfg := &config.HushConfig{}
	det := &DetectedProject{Remote: "github.com/org/repo", FullPath: "/tmp/repo"}

	Register(cfg, det, "org-repo", "")

	err := Register(cfg, det, "org-repo", "")
	if err == nil {
		t.Error("expected error for duplicate registration")
	}
}
