package sync

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureAgentsMDCreatesFile(t *testing.T) {
	dir := t.TempDir()
	r, err := EnsureAgentsMD(dir, false)
	if err != nil {
		t.Fatalf("EnsureAgentsMD: %v", err)
	}
	if len(r.Changes) != 1 {
		t.Errorf("changes = %d, want 1", len(r.Changes))
	}

	data, _ := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if !strings.Contains(string(data), "@AGENTS.override.md") {
		t.Error("created AGENTS.md should contain @AGENTS.override.md")
	}
}

func TestEnsureAgentsMDAppendsInclude(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("# Rules\nSome rules here.\n"), 0o644)

	r, err := EnsureAgentsMD(dir, false)
	if err != nil {
		t.Fatalf("EnsureAgentsMD: %v", err)
	}
	if len(r.Changes) != 1 {
		t.Errorf("changes = %d, want 1", len(r.Changes))
	}

	data, _ := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	content := string(data)
	if !strings.HasPrefix(content, "# Rules\n") {
		t.Error("original content should be preserved")
	}
	if !strings.Contains(content, "@AGENTS.override.md") {
		t.Error("should contain @AGENTS.override.md")
	}
}

func TestEnsureAgentsMDNoopIfPresent(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("# Rules\n@AGENTS.override.md\n"), 0o644)

	r, err := EnsureAgentsMD(dir, false)
	if err != nil {
		t.Fatalf("EnsureAgentsMD: %v", err)
	}
	if len(r.Changes) != 0 {
		t.Error("should be noop when include already present")
	}
}

func TestEnsureAgentsMDDryRun(t *testing.T) {
	dir := t.TempDir()
	r, err := EnsureAgentsMD(dir, true)
	if err != nil {
		t.Fatalf("EnsureAgentsMD dry-run: %v", err)
	}
	if len(r.Changes) != 1 {
		t.Errorf("dry-run changes = %d, want 1", len(r.Changes))
	}
	// File should NOT be created.
	if _, err := os.Stat(filepath.Join(dir, "AGENTS.md")); err == nil {
		t.Error("file should not exist in dry-run mode")
	}
}

func TestEnsureClaudeMDCreates(t *testing.T) {
	dir := t.TempDir()
	r, err := EnsureClaudeMD(dir, false)
	if err != nil {
		t.Fatalf("EnsureClaudeMD: %v", err)
	}
	if len(r.Changes) != 1 {
		t.Errorf("changes = %d, want 1", len(r.Changes))
	}

	data, _ := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	if strings.TrimSpace(string(data)) != "@AGENTS.md" {
		t.Errorf("CLAUDE.md content = %q, want %q", string(data), "@AGENTS.md\n")
	}
}

func TestEnsureClaudeMDPrependsToExisting(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("# Existing rules\n"), 0o644)

	r, err := EnsureClaudeMD(dir, false)
	if err != nil {
		t.Fatalf("EnsureClaudeMD: %v", err)
	}
	if len(r.Changes) != 1 {
		t.Errorf("changes = %d, want 1", len(r.Changes))
	}

	data, _ := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	content := string(data)
	if !strings.HasPrefix(content, "@AGENTS.md\n") {
		t.Error("should start with @AGENTS.md")
	}
	if !strings.Contains(content, "# Existing rules") {
		t.Error("existing content should be preserved")
	}
}

func TestEnsureClaudeMDNoopIfPresent(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("@AGENTS.md\n# Rules\n"), 0o644)

	r, err := EnsureClaudeMD(dir, false)
	if err != nil {
		t.Fatalf("EnsureClaudeMD: %v", err)
	}
	if len(r.Changes) != 0 {
		t.Error("should be noop when @AGENTS.md already present")
	}
}

func TestEnsureGitignore(t *testing.T) {
	dir := t.TempDir()

	r, err := EnsureGitignore(dir, false)
	if err != nil {
		t.Fatalf("EnsureGitignore: %v", err)
	}
	if len(r.Changes) != 1 {
		t.Errorf("changes = %d, want 1", len(r.Changes))
	}

	data, _ := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if !strings.Contains(string(data), "AGENTS.override.md") {
		t.Error(".gitignore should contain AGENTS.override.md")
	}
}

func TestEnsureGitignoreNoopIfPresent(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("node_modules\nAGENTS.override.md\n"), 0o644)

	r, err := EnsureGitignore(dir, false)
	if err != nil {
		t.Fatalf("EnsureGitignore: %v", err)
	}
	if len(r.Changes) != 0 {
		t.Error("should be noop when already present")
	}
}

func TestEnsureGitignoreAppendsToExisting(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("node_modules\n"), 0o644)

	r, err := EnsureGitignore(dir, false)
	if err != nil {
		t.Fatalf("EnsureGitignore: %v", err)
	}
	if len(r.Changes) != 1 {
		t.Errorf("changes = %d, want 1", len(r.Changes))
	}

	data, _ := os.ReadFile(filepath.Join(dir, ".gitignore"))
	content := string(data)
	if !strings.HasPrefix(content, "node_modules\n") {
		t.Error("original content should be preserved")
	}
	if !strings.Contains(content, "AGENTS.override.md") {
		t.Error("should contain AGENTS.override.md")
	}
}

func TestEnsureOverride(t *testing.T) {
	dir := t.TempDir()
	r, err := EnsureOverride(dir, false)
	if err != nil {
		t.Fatalf("EnsureOverride: %v", err)
	}
	if len(r.Changes) != 1 {
		t.Errorf("changes = %d, want 1", len(r.Changes))
	}

	if _, err := os.Stat(filepath.Join(dir, "AGENTS.override.md")); err != nil {
		t.Error("AGENTS.override.md should be created")
	}
}

func TestEnsureOverrideNoopIfExists(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "AGENTS.override.md"), []byte("# Existing\n"), 0o644)

	r, err := EnsureOverride(dir, false)
	if err != nil {
		t.Fatalf("EnsureOverride: %v", err)
	}
	if len(r.Changes) != 0 {
		t.Error("should be noop when file exists")
	}
}

func TestScaffoldProject(t *testing.T) {
	dir := t.TempDir()
	r, err := ScaffoldProject(dir, dir, false)
	if err != nil {
		t.Fatalf("ScaffoldProject: %v", err)
	}

	// Should have created AGENTS.md, CLAUDE.md, AGENTS.override.md, .gitignore entry.
	if len(r.Changes) < 4 {
		t.Errorf("changes = %d, want >= 4", len(r.Changes))
	}
}
