package cmd

import (
	"fmt"
	"time"

	aw "github.com/deanishe/awgo"
	"github.com/google/go-github/v78/github"

	gh "github.com/rwilgaard/alfred-github-search/src/internal/github"
	"github.com/rwilgaard/go-alfredutils/alfredutils"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "search public repositories",
	Args:  cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		query := args[0]

		service, err := setupGitHubClient()
		if err != nil {
			service = gh.NewUnauthenticatedService()
		}

		repos, _, err := service.SearchRepositories(query)

		if rateLimitErr, ok := err.(*github.RateLimitError); ok {
			handleRateLimitError(rateLimitErr)
			alfredutils.HandleFeedback(wf)
			return
		}

		if err != nil {
			wf.FatalError(err)
		}

		for _, repo := range repos {
			subtitle := buildRepoSubtitle(repo)
			wf.NewItem(*repo.Name).
				UID(*repo.FullName).
				Subtitle(subtitle).
				Var("item_url", repo.GetHTMLURL()).
				Arg("repo").
				Valid(true)
		}

		alfredutils.HandleFeedback(wf)
	},
}

func handleRateLimitError(err *github.RateLimitError) {
	resetTime := err.Rate.Reset.Time
	minutesUntil := time.Until(resetTime).Round(time.Minute)

	wf.NewItem("GitHub API Rate Limit Hit").
		Subtitle(fmt.Sprintf("Try again in %s (at %s)", minutesUntil, resetTime.Local().Format("3:04 PM"))).
		Icon(aw.IconError)
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
