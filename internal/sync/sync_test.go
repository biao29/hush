package sync

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/getctx/hush/internal/config"
)

func TestSyncToProject(t *testing.T) {
	hushHome := t.TempDir()
	t.Setenv("HUSH_HOME", hushHome)

	projDir := t.TempDir()

	// Create source override in private repo.
	repoProjectDir := config.ProjectDir("test-proj")
	os.MkdirAll(repoProjectDir, 0o755)
	os.WriteFile(filepath.Join(repoProjectDir, "agents.override.md"), []byte("# Private\nSecret stuff\n"), 0o644)

	entry := &config.ProjectEntry{
		ID:   "test-proj",
		Path: projDir,
	}

	result, err := SyncToProject(entry)
	if err != nil {
		t.Fatalf("SyncToProject: %v", err)
	}
	if !result.Changed {
		t.Error("expected changed=true")
	}
	if len(result.Files) != 1 || result.Files[0] != "AGENTS.override.md" {
		t.Errorf("files = %v", result.Files)
	}

	// Verify file was copied.
	data, err := os.ReadFile(filepath.Join(projDir, "AGENTS.override.md"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(data) != "# Private\nSecret stuff\n" {
		t.Errorf("content = %q", string(data))
	}
}

func TestSyncToProjectIdempotent(t *testing.T) {
	hushHome := t.TempDir()
	t.Setenv("HUSH_HOME", hushHome)

	projDir := t.TempDir()
	content := "# Private\n"

	repoProjectDir := config.ProjectDir("test-proj")
	os.MkdirAll(repoProjectDir, 0o755)
	os.WriteFile(filepath.Join(repoProjectDir, "agents.override.md"), []byte(content), 0o644)
	os.WriteFile(filepath.Join(projDir, "AGENTS.override.md"), []byte(content), 0o644)

	entry := &config.ProjectEntry{ID: "test-proj", Path: projDir}

	result, err := SyncToProject(entry)
	if err != nil {
		t.Fatalf("SyncToProject: %v", err)
	}
	if result.Changed {
		t.Error("expected changed=false when content identical")
	}
}

func TestSyncToProjectMissingSource(t *testing.T) {
	hushHome := t.TempDir()
	t.Setenv("HUSH_HOME", hushHome)

	projDir := t.TempDir()
	os.MkdirAll(config.ProjectDir("empty-proj"), 0o755)

	entry := &config.ProjectEntry{ID: "empty-proj", Path: projDir}

	result, err := SyncToProject(entry)
	if err != nil {
		t.Fatalf("SyncToProject: %v", err)
	}
	if result.Changed {
		t.Error("expected no changes when source missing")
	}
}

func TestSyncFromProject(t *testing.T) {
	hushHome := t.TempDir()
	t.Setenv("HUSH_HOME", hushHome)

	projDir := t.TempDir()
	os.WriteFile(filepath.Join(projDir, "AGENTS.override.md"), []byte("# My private rules\n"), 0o644)
	os.WriteFile(filepath.Join(projDir, "AGENTS.md"), []byte("# Public rules\n"), 0o644)

	entry := &config.ProjectEntry{ID: "test-proj", Path: projDir}

	result, err := SyncFromProject(entry)
	if err != nil {
		t.Fatalf("SyncFromProject: %v", err)
	}
	if !result.Changed {
		t.Error("expected changed=true")
	}
	if len(result.Files) != 2 {
		t.Errorf("files = %v, want 2 files", result.Files)
	}

	// Verify files in private repo.
	repoDir := config.ProjectDir("test-proj")
	data, _ := os.ReadFile(filepath.Join(repoDir, "agents.override.md"))
	if string(data) != "# My private rules\n" {
		t.Errorf("override content = %q", string(data))
	}
	data, _ = os.ReadFile(filepath.Join(repoDir, "agents.md"))
	if string(data) != "# Public rules\n" {
		t.Errorf("agents.md content = %q", string(data))
	}
}

func TestSyncFromProjectIdempotent(t *testing.T) {
	hushHome := t.TempDir()
	t.Setenv("HUSH_HOME", hushHome)

	projDir := t.TempDir()
	content := "# Same\n"
	os.WriteFile(filepath.Join(projDir, "AGENTS.override.md"), []byte(content), 0o644)
	os.WriteFile(filepath.Join(projDir, "AGENTS.md"), []byte(content), 0o644)

	repoDir := config.ProjectDir("test-proj")
	os.MkdirAll(repoDir, 0o755)
	os.WriteFile(filepath.Join(repoDir, "agents.override.md"), []byte(content), 0o644)
	os.WriteFile(filepath.Join(repoDir, "agents.md"), []byte(content), 0o644)

	entry := &config.ProjectEntry{ID: "test-proj", Path: projDir}

	result, err := SyncFromProject(entry)
	if err != nil {
		t.Fatalf("SyncFromProject: %v", err)
	}
	if result.Changed {
		t.Error("expected no changes when content identical")
	}
}
