package cmd

import (
	"fmt"
	"time"

	"github.com/google/go-github/v74/github"
	"github.com/maniartech/gotime"
	"github.com/rwilgaard/go-alfredutils/alfredutils"
	"github.com/spf13/cobra"
)

var (
    listCmd = &cobra.Command{
        Use:   "list",
        Short: "list user repositories",
        Args:  cobra.ExactArgs(1),
        Run: func(cmd *cobra.Command, args []string) {
            alfredutils.HandleAuthentication(wf, keychainAccount)

            var repos []*github.Repository
            err := alfredutils.LoadCache(wf, repoCacheName, &repos)
            if err != nil {
                wf.FatalError(err)
            }

            maxCacheAge := time.Duration(cfg.CacheAge * int(time.Minute))
            if err := alfredutils.RefreshCache(wf, repoCacheName, maxCacheAge, []string{"cache"}); err != nil {
                wf.FatalError(err)
            }

            for _, repo := range repos {
				lastPushTime := gotime.TimeAgo(repo.PushedAt.Time)
				subtitle := fmt.Sprintf("%s  ·  ★ %d  ·  %s  ·  %s", repo.Owner.GetLogin(), repo.GetStargazersCount(), lastPushTime, repo.GetDescription())
				wf.NewItem(*repo.Name).
					UID(*repo.FullName).
					Subtitle(subtitle).
					Var("item_url", repo.GetHTMLURL()).
					Arg("repo").
					Valid(true)
            }

            query := args[0]
            if len(query) > 0 {
                wf.Filter(query)
            }
            alfredutils.HandleFeedback(wf)
        },
    }
)

func init() {
    rootCmd.AddCommand(listCmd)
}
