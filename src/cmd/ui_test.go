package cmd

import (
	"errors"
	"strings"
	"testing"
)

func TestAPIMessage(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"failed to get repository: GET https://api.github.com/repos/a/b: 404 Not Found []", "failed to get repository: 404 Not Found"},
		{"GET https://api.github.com/user: 401 Bad credentials []", "401 Bad credentials"},
		{"repo_fullname variable is empty", "repo_fullname variable is empty"},
	}
	for _, tt := range tests {
		if got := apiMessage(errors.New(tt.in)); got != tt.want {
			t.Errorf("apiMessage(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestSplitRepoName(t *testing.T) {
	tests := []struct {
		in          string
		owner, name string
		wantErr     bool
	}{
		{"rwilgaard/alfred-github-search", "rwilgaard", "alfred-github-search", false},
		{"", "", "", true},
		{"bad", "", "", true},
		{"a/b/c", "", "", true},
		{"/b", "", "", true},
	}
	for _, tt := range tests {
		owner, name, err := splitRepoName(tt.in)
		if (err != nil) != tt.wantErr || owner != tt.owner || name != tt.name {
			t.Errorf("splitRepoName(%q) = %q, %q, %v", tt.in, owner, name, err)
		}
	}
}

func TestPreviewBody(t *testing.T) {
	long := strings.Repeat("line of text\n", maxPreviewBody/10)
	tests := []struct {
		name, in string
		check    func(string) bool
	}{
		{"empty", "  ", func(s string) bool { return s == "_No description provided._" }},
		{"only template comments", "<!-- fill this in\n-->\r\n", func(s string) bool { return s == "_No description provided._" }},
		{"strips comments and CRLF", "a\r\n<!-- hidden -->b", func(s string) bool { return s == "a\nb" }},
		{"task lists", "- [x] done\n  * [ ] todo\nnot [x] a task", func(s string) bool {
			return s == "- ☑ done\n  * ☐ todo\nnot [x] a task"
		}},
		{"lines with code spans untouched", "existing tc_redirect_* tests, `a_b` stays", func(s string) bool {
			return s == "existing tc_redirect_* tests, `a_b` stays"
		}},
		{"identifiers become code", "fix tc_redirect_* and tc_redirect_netdev.h, local_delivery. _em_ and [a_b](https://x/a_b)", func(s string) bool {
			return s == "fix `tcˍredirectˍ*` and `tcˍredirectˍnetdev.h`, `localˍdelivery`. _em_ and [a_b](https://x/a_b)"
		}},
		{"fenced blocks untouched", "```\na_b\n```", func(s string) bool { return s == "```\na_b\n```" }},
		{"truncates at a line break", long, func(s string) bool {
			return len(s) < maxPreviewBody+100 && strings.HasSuffix(s, "_… truncated. Press ↩ to read it on GitHub._") &&
				strings.Contains(s, "line of text\n\n_…")
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := previewBody(tt.in); !tt.check(got) {
				t.Errorf("previewBody(%.40q) = %.80q", tt.in, got)
			}
		})
	}
}
