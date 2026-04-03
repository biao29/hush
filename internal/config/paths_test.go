package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHomeDirDefault(t *testing.T) {
	os.Unsetenv("HUSH_HOME")
	home, _ := os.UserHomeDir()
	want := filepath.Join(home, ".hush")
	if got := HomeDir(); got != want {
		t.Errorf("HomeDir() = %q, want %q", got, want)
	}
}

func TestHomeDirOverride(t *testing.T) {
	t.Setenv("HUSH_HOME", "/tmp/test-hush")
	if got := HomeDir(); got != "/tmp/test-hush" {
		t.Errorf("HomeDir() = %q, want %q", got, "/tmp/test-hush")
	}
}

func TestRepoDir(t *testing.T) {
	t.Setenv("HUSH_HOME", "/tmp/h")
	if got := RepoDir(); got != "/tmp/h/repo" {
		t.Errorf("RepoDir() = %q, want %q", got, "/tmp/h/repo")
	}
}

func TestConfigPath(t *testing.T) {
	t.Setenv("HUSH_HOME", "/tmp/h")
	if got := ConfigPath(); got != "/tmp/h/repo/hush.yaml" {
		t.Errorf("ConfigPath() = %q, want %q", got, "/tmp/h/repo/hush.yaml")
	}
}

func TestProjectDir(t *testing.T) {
	t.Setenv("HUSH_HOME", "/tmp/h")
	want := "/tmp/h/repo/projects/my-project"
	if got := ProjectDir("my-project"); got != want {
		t.Errorf("ProjectDir() = %q, want %q", got, want)
	}
}
