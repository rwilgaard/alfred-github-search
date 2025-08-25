package cmd

import (
	"context"
	"fmt"

	"github.com/google/go-github/v74/github"
	"github.com/maniartech/gotime"
	"github.com/rwilgaard/go-alfredutils/alfredutils"
	"github.com/spf13/cobra"
)

var (
	searchCmd = &cobra.Command{
		Use:   "search",
		Short: "search public repositories",
		Args:  cobra.ExactArgs(1),
		Run: func(_ *cobra.Command, args []string) {
			query := args[0]

			client := github.NewClient(nil)
			repos, _, err := client.Search.Repositories(context.Background(), query, nil)
			if err != nil {
				wf.FatalError(err)
			}

			for _, repo := range repos.Repositories {
				lastPushTime := gotime.TimeAgo(repo.PushedAt.Time)
				subtitle := fmt.Sprintf("%s  ·  ★ %d  ·  %s  ·  %s", repo.Owner.GetLogin(), repo.GetStargazersCount(), lastPushTime, repo.GetDescription())
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
)

func init() {
	rootCmd.AddCommand(searchCmd)
}
