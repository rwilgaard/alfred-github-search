package cmd

import (
	"time"

	aw "github.com/deanishe/awgo"
	"github.com/google/go-github/v89/github"
	"github.com/rwilgaard/go-alfredutils/alfredutils"
	"github.com/spf13/cobra"
)

var reposCmd = &cobra.Command{
	Use:   "repos [query]",
	Short: "list the authenticated user's repositories",
	Args:  cobra.MaximumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
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
		if err := alfredutils.RefreshCache(wf, repoCacheName, maxCacheAge, []string{"cache", "repos"}); err != nil {
			wf.FatalError(err)
		}

		refreshing := wf.Cache.Exists(repoCacheName) && wf.IsRunning(repoCacheName)

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

		if refreshing {
			prependItem("Updating repositories from GitHub...", "Showing cached results in the meantime")
		}

		alfredutils.HandleFeedback(wf)
	},
}

func init() {
	rootCmd.AddCommand(reposCmd)
}
