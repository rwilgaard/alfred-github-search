package cmd

import (
	"errors"
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
