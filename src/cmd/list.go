package cmd

import (
	"time"

	aw "github.com/deanishe/awgo"
	"github.com/google/go-github/v89/github"
	"github.com/rwilgaard/go-alfredutils/alfredutils"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "list user repositories",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := alfredutils.CheckForUpdates(wf); err != nil {
			wf.FatalError(err)
		}

		if ok := alfredutils.HandleAuthentication(wf, keychainAccount); !ok {
			return
		}

		var repos []*github.Repository
		if err := alfredutils.LoadCache(wf, repoCacheName, &repos); err != nil {
			wf.FatalError(err)
		}

		maxCacheAge := time.Duration(cfg.CacheAge) * time.Minute
		if err := alfredutils.RefreshCache(wf, repoCacheName, maxCacheAge, []string{"cache"}); err != nil {
			wf.FatalError(err)
		}

		// RefreshCache adds its own item when the cache file doesn't exist yet, so
		// only the warm case — a stale cache refreshing behind results we can
		// already show — needs an indicator here.
		refreshing := wf.Cache.Exists(repoCacheName) && wf.IsRunning(repoCacheName)

		// UIDs opt items into Alfred's usage-based reordering, which would push
		// the refresh indicator below frequently-opened repos. Suppress them for
		// the duration of a refresh so the indicator stays on top.
		if refreshing {
			wf.Configure(aw.SuppressUIDs(true))
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

		// Added after Filter so the query can't discard it, then moved to the front.
		if refreshing {
			wf.NewItem("Updating repositories from GitHub...").
				Subtitle("Showing cached results in the meantime").
				Icon(aw.IconInfo).
				Valid(false)

			items := wf.Feedback.Items
			indicator := items[len(items)-1]
			copy(items[1:], items[:len(items)-1])
			items[0] = indicator
		}

		alfredutils.HandleFeedback(wf)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
