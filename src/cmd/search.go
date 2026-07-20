package cmd

import (
	"os"
	"os/exec"

	aw "github.com/deanishe/awgo"
	"github.com/google/go-github/v89/github"

	"github.com/rwilgaard/go-alfredutils/alfredutils"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "search public repositories",
	Args:  cobra.ExactArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		query := args[0]
		cacheName := searchCacheName(query)
		jobName := searchJobName(query)

		var repos []*github.Repository
		cacheExists := wf.Cache.Exists(cacheName)
		fetching := wf.Cache.Expired(cacheName, searchCacheMaxAge)

		if cacheExists && !fetching {
			if err := wf.Cache.LoadJSON(cacheName, &repos); err != nil {
				wf.FatalError(err)
			}
		} else {
			if !wf.IsRunning(jobName) {
				exe, err := os.Executable()
				if err != nil {
					wf.FatalError(err)
				}
				cmd := exec.Command(exe, "cache", "search", query)
				if err := wf.RunInBackground(jobName, cmd); err != nil {
					wf.FatalError(err)
				}
			}

			wf.Rerun(searchRerunInterval)

			wf.Configure(aw.SuppressUIDs(true))

			if cacheExists {
				if err := wf.Cache.LoadJSON(cacheName, &repos); err != nil {
					wf.FatalError(err)
				}
			}
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

		if fetching {
			if cacheExists {
				prependItem("Updating results from GitHub...", "Showing cached results in the meantime")
			} else {
				prependItem("Searching GitHub...", "Fetching results for '"+query+"'...")
			}
		}

		alfredutils.HandleFeedback(wf)
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
