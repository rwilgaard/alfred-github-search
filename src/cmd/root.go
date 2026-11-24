package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/go-github/v89/github"
	"github.com/maniartech/gotime/v2"
	gh "github.com/rwilgaard/alfred-github-search/src/internal/github"
	"github.com/rwilgaard/alfred-github-search/src/internal/util"
	"github.com/rwilgaard/go-alfredutils/v2/alfred"
	"github.com/rwilgaard/go-alfredutils/v2/alfred/update"
	"github.com/spf13/cobra"
	"go.deanishe.net/fuzzy"
)

type workflowConfig struct {
	CacheAge     int    `env:"cache_age"`
	GitHubToken  string `env:"github_token"`
	RepoFullName string `env:"repo_fullname"`
	RepoCloneURL string `env:"repo_clone_url"`
	RepoSSHURL   string `env:"repo_ssh_url"`
}

const (
	repo            = "rwilgaard/alfred-github-search"
	repoCacheName   = "repositories.json"
	repoCacheJob    = "cache"
	keychainAccount = "alfred-github-search"

	// How long API responses are reused without asking GitHub again.
	searchMaxAge = 10 * time.Minute
	listMaxAge   = time.Minute
)

var (
	wf      *alfred.Workflow
	updater *update.Updater
	cfg     = &workflowConfig{CacheAge: 360}

	rootCmd = &cobra.Command{
		Use:           "alfred-github-search",
		Short:         "alfred-github-search is a CLI to be used by Alfred for searching GitHub repositories",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
)

func Execute() {
	var err error
	wf, err = alfred.New(
		alfred.WithUsage(),
		alfred.WithFuzzy(
			fuzzy.AdjacencyBonus(10.0),
			fuzzy.LeadingLetterPenalty(-0.1),
			fuzzy.MaxLeadingLetterPenalty(-3.0),
			fuzzy.UnmatchedLetterPenalty(-0.5),
		),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err) // not run by Alfred: nobody reads Script Filter JSON
		os.Exit(1)
	}
	// Both register "workflow:" magic queries, which wf.Run handles before
	// run(), so they must come first.
	updater = update.GitHub(wf, repo)
	wf.AddClearAuth(keychainAccount)
	wf.Run(run)
}

func run() {
	if err := wf.Config(cfg); err != nil {
		wf.FatalError(err)
	}
	if err := updater.Run(); err != nil {
		wf.FatalError(err)
	}
	if err := rootCmd.Execute(); err != nil {
		wf.FatalError(err)
	}
}

// setupGitHubClient returns a client for the stored token, with API responses
// cached for maxAge. The token is passed on as the github_token variable, so
// the sub-menus opened from these results don't read the Keychain again.
func setupGitHubClient(maxAge time.Duration) (*gh.GithubService, error) {
	token := cfg.GitHubToken
	if token == "" {
		var err error
		token, err = wf.LoadAuth(keychainAccount)
		if err != nil {
			return nil, fmt.Errorf("failed to get token from keychain: %w", err)
		}
	}
	wf.Var("github_token", token)

	return gh.NewAuthenticatedService(token, wf.Cache, maxAge)
}

func splitRepoName(fullName string) (owner, name string, err error) {
	if fullName == "" {
		return "", "", fmt.Errorf("repo_fullname variable is empty")
	}
	owner, name, ok := strings.Cut(fullName, "/")
	if !ok || owner == "" || name == "" || strings.Contains(name, "/") {
		return "", "", fmt.Errorf("invalid repo_fullname: %s", fullName)
	}
	return owner, name, nil
}

// addRepoItems adds one item per repository. ⏎ opens it in the browser, ⌘⏎
// opens the repository window (workflow/repo.alfredui).
func addRepoItems(repos []*github.Repository) {
	for _, r := range repos {
		item := wf.NewItem(r.GetName()).
			UID(r.GetFullName()).
			Subtitle(buildRepoSubtitle(r)).
			Icon(repoIcon(r)).
			MatchOn(alfred.MatchSecondary, r.GetFullName()).
			MatchOn(alfred.MatchMeta, r.GetDescription()).
			Var("item_url", r.GetHTMLURL()).
			Var("repo_fullname", r.GetFullName()).
			Arg("repo").
			Valid(true)

		item.NewModifier(alfred.ModCmd).
			Subtitle("Show details, issues and pull requests").
			Var("repo_html_url", r.GetHTMLURL()).
			Var("repo_clone_url", r.GetCloneURL()).
			Var("repo_ssh_url", r.GetSSHURL()).
			Arg("details").
			Valid(true)
	}
}

func repoIcon(r *github.Repository) alfred.Icon {
	if r.GetPrivate() {
		return util.IconRepoPrivate
	}
	return util.IconRepoPublic
}

func buildRepoSubtitle(r *github.Repository) string {
	lastPushTime := "never"
	if r.PushedAt != nil {
		lastPushTime = gotime.TimeAgo(*r.PushedAt.GetTime())
	}

	parts := []string{
		r.GetOwner().GetLogin(),
		fmt.Sprintf("★ %d", r.GetStargazersCount()),
		lastPushTime,
	}

	if desc := r.GetDescription(); desc != "" {
		parts = append(parts, desc)
	}

	return strings.Join(parts, "  ·  ")
}
