package cmd

import (
	"fmt"

	aw "github.com/deanishe/awgo"
	"github.com/deanishe/awgo/update"
	"github.com/google/go-github/v78/github"
	"github.com/maniartech/gotime"
	gh "github.com/rwilgaard/alfred-github-search/src/internal/github"
	"github.com/rwilgaard/go-alfredutils/alfredutils"

	"github.com/spf13/cobra"
	"go.deanishe.net/fuzzy"
)

type workflowConfig struct {
	CacheAge int `env:"cache_age"`
}

const (
	repo            = "rwilgaard/alfred-github-search"
	repoCacheName   = "repositories.json"
	keychainAccount = "alfred-github-search"
)

var (
	wf      *aw.Workflow
	cfg     = &workflowConfig{}
	rootCmd = &cobra.Command{
		Use:           "github",
		Short:         "github is a CLI to be used by Alfred for searching Github repositories",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
)

func Execute() {
	wf.Run(run)
}

func run() {
	alfredutils.AddClearAuthMagic(wf, keychainAccount)

	if err := alfredutils.InitWorkflow(wf, cfg); err != nil {
		wf.FatalError(err)
	}

	if err := alfredutils.CheckForUpdates(wf); err != nil {
		wf.FatalError(err)
	}

	if err := rootCmd.Execute(); err != nil {
		wf.FatalError(err)
	}
}

func setupGitHubClient() (*gh.GithubService, error) {
	token, err := wf.Keychain.Get(keychainAccount)
	if err != nil {
		return nil, fmt.Errorf("failed to get token from keychain: %w", err)
	}

	return gh.NewAuthenticatedService(token), nil
}

func buildRepoSubtitle(repo *github.Repository) string {
	var lastPushTime string
	if repo.PushedAt != nil {
		lastPushTime = gotime.TimeAgo(*repo.PushedAt.GetTime())
	} else {
		lastPushTime = "never"
	}

	owner := repo.Owner.GetLogin()
	stars := repo.GetStargazersCount()
	desc := repo.GetDescription()

	return fmt.Sprintf("%s  ·  ★ %d  ·  %s  ·  %s", owner, stars, lastPushTime, desc)
}

func init() {
	sopts := []fuzzy.Option{
		fuzzy.AdjacencyBonus(10.0),
		fuzzy.LeadingLetterPenalty(-0.1),
		fuzzy.MaxLeadingLetterPenalty(-3.0),
		fuzzy.UnmatchedLetterPenalty(-0.5),
	}
	wf = aw.New(
		aw.SortOptions(sopts...),
		update.GitHub(repo),
	)
}
