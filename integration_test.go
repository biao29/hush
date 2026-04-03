package integration_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/getctx/hush/internal/config"
	"github.com/getctx/hush/internal/gitops"
	hsync "github.com/getctx/hush/internal/sync"
)

// TestFullLifecycle simulates: init → link → edit → push → sync on another "device".
func TestFullLifecycle(t *testing.T) {
	// Skip if git is not available.
	if _, err := gitops.Run(".", "version"); err != nil {
		t.Skipf("git not available: %v", err)
	}

	// Setup: create a bare repo to act as the "remote".
	bareDir := t.TempDir()
	if _, err := gitops.Run(bareDir, "init", "--bare"); err != nil {
		t.Fatalf("init bare: %v", err)
	}

	// === Device 1: Init ===
	hushHome1 := t.TempDir()
	t.Setenv("HUSH_HOME", hushHome1)

	repoDir := filepath.Join(hushHome1, "repo")
	if err := gitops.Clone(bareDir, repoDir); err != nil {
		t.Fatalf("clone: %v", err)
	}

	// Create initial structure (simulates hush init on empty repo).
	os.MkdirAll(filepath.Join(repoDir, "projects"), 0o755)
	cfg := &config.HushConfig{}
	if err := cfg.SaveTo(filepath.Join(repoDir, "hush.yaml")); err != nil {
		t.Fatalf("save config: %v", err)
	}
	gitops.Run(repoDir, "config", "user.email", "test@test.com")
	gitops.Run(repoDir, "config", "user.name", "Test")
	gitops.CommitAll(repoDir, "initial")
	gitops.Push(repoDir)

	// === Device 1: Create a project and link ===
	projectDir := t.TempDir()
	gitops.Run(projectDir, "init")
	gitops.Run(projectDir, "config", "user.email", "test@test.com")
	gitops.Run(projectDir, "config", "user.name", "Test")
	gitops.Run(projectDir, "remote", "add", "origin", "git@github.com:testorg/testrepo.git")

	// Scaffold the project.
	r, err := hsync.ScaffoldProject(projectDir, projectDir, false)
	if err != nil {
		t.Fatalf("scaffold: %v", err)
	}
	if len(r.Changes) < 4 {
		t.Errorf("scaffold changes = %d, want >= 4", len(r.Changes))
	}

	// Verify scaffolded files.
	agentsData, _ := os.ReadFile(filepath.Join(projectDir, "AGENTS.md"))
	if !strings.Contains(string(agentsData), "@AGENTS.override.md") {
		t.Error("AGENTS.md should contain @AGENTS.override.md")
	}

	claudeData, _ := os.ReadFile(filepath.Join(projectDir, "CLAUDE.md"))
	if !strings.Contains(string(claudeData), "@AGENTS.md") {
		t.Error("CLAUDE.md should contain @AGENTS.md")
	}

	gitignoreData, _ := os.ReadFile(filepath.Join(projectDir, ".gitignore"))
	if !strings.Contains(string(gitignoreData), "AGENTS.override.md") {
		t.Error(".gitignore should contain AGENTS.override.md")
	}

	if _, err := os.Stat(filepath.Join(projectDir, "AGENTS.override.md")); err != nil {
		t.Error("AGENTS.override.md should exist")
	}

	// Register the project.
	cfg, _ = config.LoadFrom(filepath.Join(repoDir, "hush.yaml"))
	entry := config.ProjectEntry{
		ID:     "testorg-testrepo",
		Remote: "github.com/testorg/testrepo",
		Path:   projectDir,
	}
	if err := cfg.AddProject(entry); err != nil {
		t.Fatalf("add project: %v", err)
	}
	cfg.SaveTo(filepath.Join(repoDir, "hush.yaml"))
	os.MkdirAll(config.ProjectDir("testorg-testrepo"), 0o755)

	// === Device 1: Edit private rules ===
	os.WriteFile(filepath.Join(projectDir, "AGENTS.override.md"), []byte("# Private\nAPI_KEY=secret\n"), 0o644)

	// === Device 1: Push ===
	pushResult, err := hsync.SyncFromProject(&entry)
	if err != nil {
		t.Fatalf("push: %v", err)
	}
	if !pushResult.Changed {
		t.Error("push should report changes")
	}

	// Verify files in private repo.
	overrideData, _ := os.ReadFile(filepath.Join(config.ProjectDir("testorg-testrepo"), "agents.override.md"))
	if string(overrideData) != "# Private\nAPI_KEY=secret\n" {
		t.Errorf("override in repo = %q", string(overrideData))
	}

	gitops.CommitAll(repoDir, "push changes")
	gitops.Push(repoDir)

	// === Device 2: Init + Sync ===
	hushHome2 := t.TempDir()
	t.Setenv("HUSH_HOME", hushHome2)

	repoDir2 := filepath.Join(hushHome2, "repo")
	if err := gitops.Clone(bareDir, repoDir2); err != nil {
		t.Fatalf("clone device 2: %v", err)
	}

	// Create project dir on device 2.
	projectDir2 := t.TempDir()
	gitops.Run(projectDir2, "init")
	gitops.Run(projectDir2, "remote", "add", "origin", "git@github.com:testorg/testrepo.git")

	// Update entry path for device 2.
	cfg2, _ := config.LoadFrom(filepath.Join(repoDir2, "hush.yaml"))
	entry2 := cfg2.FindProject("testorg-testrepo")
	if entry2 == nil {
		t.Fatal("project not found in device 2 config")
	}
	entry2.Path = projectDir2
	cfg2.SaveTo(filepath.Join(repoDir2, "hush.yaml"))

	// Sync to project.
	syncResult, err := hsync.SyncToProject(entry2)
	if err != nil {
		t.Fatalf("sync: %v", err)
	}
	if !syncResult.Changed {
		t.Error("sync should report changes")
	}

	// Verify override arrived on device 2.
	data, err := os.ReadFile(filepath.Join(projectDir2, "AGENTS.override.md"))
	if err != nil {
		t.Fatalf("read override on device 2: %v", err)
	}
	if string(data) != "# Private\nAPI_KEY=secret\n" {
		t.Errorf("device 2 override = %q", string(data))
	}

	// === Idempotency: sync again should be no-op ===
	syncResult2, err := hsync.SyncToProject(entry2)
	if err != nil {
		t.Fatalf("sync again: %v", err)
	}
	if syncResult2.Changed {
		t.Error("second sync should be no-op (idempotent)")
	}
}

// TestScaffoldIdempotency verifies running scaffold twice produces no changes.
func TestScaffoldIdempotency(t *testing.T) {
	dir := t.TempDir()

	// First run.
	r1, err := hsync.ScaffoldProject(dir, dir, false)
	if err != nil {
		t.Fatalf("scaffold 1: %v", err)
	}
	if len(r1.Changes) == 0 {
		t.Error("first scaffold should have changes")
	}

	// Second run.
	r2, err := hsync.ScaffoldProject(dir, dir, false)
	if err != nil {
		t.Fatalf("scaffold 2: %v", err)
	}
	if len(r2.Changes) != 0 {
		t.Errorf("second scaffold should be noop, got changes: %v", r2.Changes)
	}
}
