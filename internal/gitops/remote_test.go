package gitops

import "testing"

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"git@github.com:org/repo.git", "github.com/org/repo"},
		{"git@github.com:org/repo", "github.com/org/repo"},
		{"https://github.com/org/repo.git", "github.com/org/repo"},
		{"https://github.com/org/repo", "github.com/org/repo"},
		{"http://github.com/org/repo.git", "github.com/org/repo"},
		{"ssh://git@github.com/org/repo.git", "github.com/org/repo"},
		{"ssh://git@github.com/org/repo", "github.com/org/repo"},
		{"git@gitlab.com:group/subgroup/repo.git", "gitlab.com/group/subgroup/repo"},
		{"ssh://git@gitlab.com/group/subgroup/repo.git", "gitlab.com/group/subgroup/repo"},
		{"https://gitlab.com/group/subgroup/repo.git", "gitlab.com/group/subgroup/repo"},
		{"github.com/org/repo", "github.com/org/repo"}, // already normalized
		{"", ""},
		{"  git@github.com:org/repo.git  ", "github.com/org/repo"}, // whitespace
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := NormalizeURL(tt.input)
			if got != tt.want {
				t.Errorf("NormalizeURL(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestGenerateProjectID(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"github.com/openelf/getctx.org", "openelf-getctx.org"},
		{"github.com/user/repo", "user-repo"},
		{"gitlab.com/group/subgroup/repo", "subgroup-repo"},
		{"single", "single"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := GenerateProjectID(tt.input)
			if got != tt.want {
				t.Errorf("GenerateProjectID(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestGenerateProjectIDWithSubdir(t *testing.T) {
	tests := []struct {
		url    string
		subdir string
		want   string
	}{
		{"github.com/org/mono", "api", "org-mono-api"},
		{"github.com/org/mono", "services/api", "org-mono-services-api"},
		{"github.com/org/mono", "", "org-mono"},
	}

	for _, tt := range tests {
		t.Run(tt.url+"/"+tt.subdir, func(t *testing.T) {
			got := GenerateProjectIDWithSubdir(tt.url, tt.subdir)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}
