package cmd

import (
	"fmt"
	"log"

	gh "github.com/rwilgaard/alfred-github-search/src/internal/github"
	"github.com/spf13/cobra"
)

// searchCmd searches GitHub on every query. Alfred waits for typing to pause
// before running it, and responses are cached for searchMaxAge, so repeating
// or backspacing to an earlier query makes no request.
var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "search public repositories",
	Args:  cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		query := args[0]

		// Public search works without a token, so an auth failure degrades to
		// an unauthenticated (rate-limited) client rather than failing outright.
		service, authErr := setupGitHubClient(searchMaxAge)
		if authErr != nil {
			log.Printf("falling back to unauthenticated search: %v", authErr)
			var err error
			service, err = gh.NewUnauthenticatedService(wf.Cache, searchMaxAge)
			if err != nil {
				wf.FatalError(err)
			}
		}

		repos, err := service.SearchRepositories(query)
		if err != nil {
			if authErr != nil {
				err = fmt.Errorf("%w (unauthenticated: %w)", err, authErr)
			}
			wf.FatalError(err)
		}

		addRepoItems(repos)
		wf.Send()
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
