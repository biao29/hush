package doctor

import "testing"

func TestCheckGit(t *testing.T) {
	c := checkGit()
	// Git should be available in CI/dev environments.
	if c.Status != "ok" {
		t.Skipf("git not available: %s", c.Detail)
	}
}

func TestCheckRepoDirMissing(t *testing.T) {
	t.Setenv("HUSH_HOME", "/nonexistent/path")
	c := checkRepoDir()
	if c.Status != "fail" {
		t.Errorf("status = %q, want %q", c.Status, "fail")
	}
}

func TestFmtCount(t *testing.T) {
	tests := []struct {
		n    int
		noun string
		want string
	}{
		{0, "project", "0 projects"},
		{1, "project", "1 project"},
		{5, "file", "5 files"},
	}
	for _, tt := range tests {
		got := fmtCount(tt.n, tt.noun)
		if got != tt.want {
			t.Errorf("fmtCount(%d, %q) = %q, want %q", tt.n, tt.noun, got, tt.want)
		}
	}
}
