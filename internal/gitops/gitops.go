package gitops

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const defaultTimeout = 30 * time.Second

// Run executes a git command in the given directory with a timeout.
func Run(dir string, args ...string) (string, error) {
	return RunWithTimeout(dir, defaultTimeout, args...)
}

// RunWithTimeout executes a git command with a custom timeout.
func RunWithTimeout(dir string, timeout time.Duration, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == "" {
			errMsg = err.Error()
		}
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), errMsg)
	}

	return strings.TrimSpace(stdout.String()), nil
}

// Clone clones a git repository to the destination directory.
func Clone(url, dest string) error {
	_, err := RunWithTimeout(".", 120*time.Second, "clone", url, dest)
	return err
}

// Pull runs git pull --ff-only in the given directory.
func Pull(dir string) error {
	_, err := Run(dir, "pull", "--ff-only")
	return err
}

// CommitAll stages all changes and commits with the given message.
func CommitAll(dir, message string) error {
	if _, err := Run(dir, "add", "-A"); err != nil {
		return err
	}
	_, err := Run(dir, "commit", "-m", message)
	return err
}

// Push pushes to the remote.
func Push(dir string) error {
	_, err := RunWithTimeout(dir, 60*time.Second, "push")
	return err
}

// IsClean returns true if the working tree has no uncommitted changes.
func IsClean(dir string) (bool, error) {
	out, err := Run(dir, "status", "--porcelain")
	if err != nil {
		return false, err
	}
	return out == "", nil
}

// RemoteURL returns the origin remote URL for the repo at dir.
func RemoteURL(dir string) (string, error) {
	return Run(dir, "remote", "get-url", "origin")
}

// RepoRoot returns the git repository root for the given directory.
func RepoRoot(dir string) (string, error) {
	return Run(dir, "rev-parse", "--show-toplevel")
}

// IsGitRepo returns true if dir is inside a git repository.
func IsGitRepo(dir string) bool {
	_, err := Run(dir, "rev-parse", "--git-dir")
	return err == nil
}

// InitBare initializes a bare git repository.
func InitBare(dir string) error {
	_, err := Run(dir, "init", "--bare")
	return err
}
