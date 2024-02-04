package cmd

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/rwilgaard/alfred-github-search/src/pkg/github"
	"github.com/rwilgaard/go-alfredutils/alfredutils"
	"github.com/spf13/cobra"
)

var (
    searchCmd = &cobra.Command{
        Use:   "search",
        Short: "search repositories globally",
        Args:  cobra.ExactArgs(1),
        Run: func(cmd *cobra.Command, args []string) {
            query := args[0]
            if !strings.Contains(query, "in:name") {
                query += " in:name"
            }

            var gh *github.GithubService
            token, err := wf.Keychain.Get(keychainAccount)
            if err != nil {
                log.Printf("Using unauthenticated client")
                gh = github.NewGithubService()
            } else {
                log.Printf("Using authenticated client")
                gh = github.NewTokenGithubService(token)
            }

            repos, _, err := gh.Client.Search.Repositories(context.Background(), query, nil)
            if err != nil {
                wf.FatalError(err)
            }

            for _, repo := range repos.Repositories {
                updatedAt := repo.UpdatedAt.Time.Format("02-01-2006 15:04")
                wf.NewItem(*repo.FullName).
                    UID(*repo.FullName).
                    Subtitle(fmt.Sprintf("%s  •  Updated: %s", *repo.HTMLURL, updatedAt)).
                    Var("item_url", *repo.HTMLURL).
                    Arg("repo").
                    Valid(true)
            }

            alfredutils.HandleFeedback(wf)
        },
    }
)

func init() {
    rootCmd.AddCommand(searchCmd)
}
