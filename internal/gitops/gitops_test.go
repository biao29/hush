package gitops

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsGitRepo(t *testing.T) {
	// Create a temp dir with git init.
	dir := t.TempDir()
	if _, err := Run(dir, "init"); err != nil {
		t.Skipf("git not available: %v", err)
	}

	if !IsGitRepo(dir) {
		t.Error("expected IsGitRepo=true after init")
	}

	noGitDir := t.TempDir()
	if IsGitRepo(noGitDir) {
		t.Error("expected IsGitRepo=false for non-git dir")
	}
}

func TestRepoRoot(t *testing.T) {
	dir := t.TempDir()
	if _, err := Run(dir, "init"); err != nil {
		t.Skipf("git not available: %v", err)
	}

	root, err := RepoRoot(dir)
	if err != nil {
		t.Fatalf("RepoRoot: %v", err)
	}

	// Resolve symlinks for macOS /private/tmp.
	wantDir, _ := filepath.EvalSymlinks(dir)
	gotDir, _ := filepath.EvalSymlinks(root)
	if gotDir != wantDir {
		t.Errorf("RepoRoot = %q, want %q", gotDir, wantDir)
	}
}

func TestIsClean(t *testing.T) {
	dir := t.TempDir()
	if _, err := Run(dir, "init"); err != nil {
		t.Skipf("git not available: %v", err)
	}

	clean, err := IsClean(dir)
	if err != nil {
		t.Fatalf("IsClean: %v", err)
	}
	if !clean {
		t.Error("expected clean after init")
	}

	// Create an untracked file.
	os.WriteFile(filepath.Join(dir, "test.txt"), []byte("hello"), 0o644)

	clean, err = IsClean(dir)
	if err != nil {
		t.Fatalf("IsClean: %v", err)
	}
	if clean {
		t.Error("expected dirty after adding file")
	}
}

func TestCommitAll(t *testing.T) {
	dir := t.TempDir()
	if _, err := Run(dir, "init"); err != nil {
		t.Skipf("git not available: %v", err)
	}
	// Set user for commit.
	Run(dir, "config", "user.email", "test@test.com")
	Run(dir, "config", "user.name", "Test")

	os.WriteFile(filepath.Join(dir, "file.txt"), []byte("content"), 0o644)

	if err := CommitAll(dir, "initial"); err != nil {
		t.Fatalf("CommitAll: %v", err)
	}

	clean, _ := IsClean(dir)
	if !clean {
		t.Error("expected clean after commit")
	}
}

func TestCloneAndPull(t *testing.T) {
	// Create a bare repo.
	bareDir := t.TempDir()
	if _, err := Run(bareDir, "init", "--bare"); err != nil {
		t.Skipf("git not available: %v", err)
	}

	// Clone it.
	cloneDir := filepath.Join(t.TempDir(), "clone")
	if err := Clone(bareDir, cloneDir); err != nil {
		t.Fatalf("Clone: %v", err)
	}

	if !IsGitRepo(cloneDir) {
		t.Error("cloned dir should be a git repo")
	}

	// Set user for commit.
	Run(cloneDir, "config", "user.email", "test@test.com")
	Run(cloneDir, "config", "user.name", "Test")

	// Create a file, commit, push.
	os.WriteFile(filepath.Join(cloneDir, "test.txt"), []byte("hello"), 0o644)
	CommitAll(cloneDir, "add test")
	Push(cloneDir)

	// Clone again and verify.
	clone2Dir := filepath.Join(t.TempDir(), "clone2")
	Clone(bareDir, clone2Dir)

	data, err := os.ReadFile(filepath.Join(clone2Dir, "test.txt"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(data) != "hello" {
		t.Errorf("content = %q, want %q", string(data), "hello")
	}
}
