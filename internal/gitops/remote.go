// Package gitops provides Git command wrappers and URL normalization.
package gitops

import (
	"regexp"
	"strings"
)

// scpPattern matches git@host:org/repo.git style URLs (SCP syntax).
var scpPattern = regexp.MustCompile(`^git@([^:/]+):(.+?)(?:\.git)?$`)

// sshURLPattern matches ssh://git@host/org/repo.git style URLs.
var sshURLPattern = regexp.MustCompile(`^ssh://git@([^:/]+)/(.+?)(?:\.git)?$`)

// httpsPattern matches https://host/org/repo.git style URLs.
var httpsPattern = regexp.MustCompile(`^https?://([^/]+)/(.+?)(?:\.git)?$`)

// NormalizeURL converts a git remote URL to a canonical form: "host/path"
// with no protocol, no .git suffix. Works for both SSH and HTTPS remotes.
//
// Examples:
//
//	git@github.com:org/repo.git     → github.com/org/repo
//	https://github.com/org/repo.git → github.com/org/repo
//	ssh://git@github.com/org/repo   → github.com/org/repo
func NormalizeURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	// ssh://git@host/path (URL-style SSH) — must check before SCP pattern.
	if m := sshURLPattern.FindStringSubmatch(raw); m != nil {
		return m[1] + "/" + m[2]
	}
	// git@host:path (SCP-style SSH)
	if m := scpPattern.FindStringSubmatch(raw); m != nil {
		return m[1] + "/" + m[2]
	}
	if m := httpsPattern.FindStringSubmatch(raw); m != nil {
		return m[1] + "/" + m[2]
	}

	// Already normalized or unrecognized — return as-is.
	return raw
}

// GenerateProjectID extracts the last two path segments from a normalized URL
// and joins them with a hyphen.
//
// Examples:
//
//	github.com/acme/widgets → acme-widgets
//	github.com/user/repo          → user-repo
func GenerateProjectID(normalizedURL string) string {
	parts := strings.Split(normalizedURL, "/")
	if len(parts) < 2 {
		return normalizedURL
	}
	org := parts[len(parts)-2]
	repo := parts[len(parts)-1]
	return org + "-" + repo
}

// GenerateProjectIDWithSubdir appends the subdir to make monorepo IDs unique.
//
// Examples:
//
//	(github.com/org/mono, "api")    → org-mono-api
//	(github.com/org/mono, "")       → org-mono
func GenerateProjectIDWithSubdir(normalizedURL, subdir string) string {
	id := GenerateProjectID(normalizedURL)
	if subdir != "" {
		// Replace path separators with hyphens.
		subdir = strings.ReplaceAll(subdir, "/", "-")
		subdir = strings.ReplaceAll(subdir, "\\", "-")
		id += "-" + subdir
	}
	return id
}
