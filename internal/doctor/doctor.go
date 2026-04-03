// Package doctor provides diagnostic checks for hush configuration.
package doctor

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/getctx/hush/internal/config"
	"github.com/getctx/hush/internal/gitops"
	"github.com/getctx/hush/internal/sync"
)

// Check represents a single diagnostic check result.
type Check struct {
	Name   string
	Status string // "ok", "warn", "fail"
	Detail string
}

// RunChecks performs all diagnostic checks and returns results.
func RunChecks() []Check {
	var checks []Check

	checks = append(checks, checkGit())
	checks = append(checks, checkRepoDir())
	checks = append(checks, checkConfig()...)

	return checks
}

func checkGit() Check {
	_, err := exec.LookPath("git")
	if err != nil {
		return Check{Name: "git", Status: "fail", Detail: "git not found in PATH"}
	}
	return Check{Name: "git", Status: "ok", Detail: "git available"}
}

func checkRepoDir() Check {
	dir := config.RepoDir()
	if _, err := os.Stat(dir); err != nil {
		return Check{Name: "private repo", Status: "fail", Detail: "not initialized — run 'hush init <repo-url>'"}
	}
	return Check{Name: "private repo", Status: "ok", Detail: dir}
}

func checkConfig() []Check {
	var checks []Check

	cfg, err := config.Load()
	if err != nil {
		checks = append(checks, Check{Name: "hush.yaml", Status: "fail", Detail: err.Error()})
		return checks
	}
	checks = append(checks, Check{
		Name:   "hush.yaml",
		Status: "ok",
		Detail: fmtCount(len(cfg.Projects), "project"),
	})

	for _, p := range cfg.Projects {
		checks = append(checks, checkProject(p))
	}

	return checks
}

func checkProject(p config.ProjectEntry) Check {
	name := "project:" + p.ID

	// Check path exists.
	if _, err := os.Stat(p.Path); err != nil {
		return Check{Name: name, Status: "warn", Detail: "path not found: " + p.Path}
	}

	// Check AGENTS.override.md exists.
	overridePath := filepath.Join(p.Path, "AGENTS.override.md")
	if _, err := os.Stat(overridePath); err != nil {
		return Check{Name: name, Status: "warn", Detail: "AGENTS.override.md missing — run 'hush sync'"}
	}

	// Check .gitignore at the actual git root (important for monorepos).
	gitRoot, err := gitops.RepoRoot(p.Path)
	if err != nil {
		gitRoot = p.Path // fallback if not a git repo
	}
	gitignorePath := filepath.Join(gitRoot, ".gitignore")
	data, err := os.ReadFile(gitignorePath)
	if err != nil {
		return Check{Name: name, Status: "warn", Detail: ".gitignore missing or unreadable at " + gitRoot}
	}
	r, _ := sync.EnsureGitignore(gitRoot, true) // dry-run check
	if len(r.Changes) > 0 {
		return Check{Name: name, Status: "warn", Detail: "AGENTS.override.md not in .gitignore at " + gitRoot}
	}
	_ = data

	return Check{Name: name, Status: "ok", Detail: p.Path}
}

func fmtCount(n int, noun string) string {
	if n == 1 {
		return "1 " + noun
	}
	return fmtInt(n) + " " + noun + "s"
}

func fmtInt(n int) string {
	s := ""
	if n == 0 {
		return "0"
	}
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}
