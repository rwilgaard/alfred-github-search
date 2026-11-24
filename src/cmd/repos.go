package cmd

import (
	"time"

	"github.com/google/go-github/v89/github"
	"github.com/rwilgaard/alfred-github-search/src/internal/util"
	"github.com/spf13/cobra"
)

var reposCmd = &cobra.Command{
	Use:   "repos [query]",
	Short: "list the authenticated user's repositories",
	Args:  cobra.MaximumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		if !wf.RequireAuth(keychainAccount) {
			return
		}

		var repos []*github.Repository
		if err := wf.Cache.LoadJSONIfExists(repoCacheName, &repos); err != nil {
			wf.FatalError(err)
		}

		maxCacheAge := time.Duration(cfg.CacheAge) * time.Minute
		refreshing, err := wf.RefreshIfStale(repoCacheName, maxCacheAge, repoCacheJob)
		if err != nil {
			wf.FatalError(err)
		}

		var query string
		if len(args) > 0 {
			query = args[0]
		}

		if refreshing && len(repos) > 0 {
			wf.NewItem("Updating repositories from GitHub...").
				Subtitle("Showing cached results in the meantime").
				Icon(util.IconRefresh).
				Pin()
		}

		addRepoItems(repos)

		wf.Filter(query)
		wf.Send()
	},
}

func init() {
	rootCmd.AddCommand(reposCmd)
}
