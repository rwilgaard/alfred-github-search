package cmd

import (
	"time"

	"github.com/google/go-github/v78/github"
	"github.com/rwilgaard/go-alfredutils/alfredutils"
	"github.com/spf13/cobra"
)

var (
	listCmd = &cobra.Command{
		Use:   "list",
		Short: "list user repositories",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if ok := alfredutils.HandleAuthentication(wf, keychainAccount); !ok {
				return
			}

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
				subtitle := buildRepoSubtitle(repo)
				wf.NewItem(*repo.Name).
					UID(*repo.FullName).
					Subtitle(subtitle).
					Var("item_url", repo.GetHTMLURL()).
					Arg("repo").
					Valid(true)
			}

			var query string
			if len(args) > 0 {
				query = args[0]
			}

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
