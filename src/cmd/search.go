package cmd

import (
	"crypto/md5"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	aw "github.com/deanishe/awgo"
	"github.com/google/go-github/v89/github"

	gh "github.com/rwilgaard/alfred-github-search/src/internal/github"
	"github.com/rwilgaard/go-alfredutils/alfredutils"
	"github.com/spf13/cobra"
)

var (
	backgroundSearch bool
	searchCmd        = &cobra.Command{
		Use:   "search",
		Short: "search public repositories",
		Args:  cobra.ExactArgs(1),
		Run: func(_ *cobra.Command, args []string) {
			query := args[0]

			hasher := md5.New()
			hasher.Write([]byte(query))
			queryHash := fmt.Sprintf("%x", hasher.Sum(nil))

			cacheName := fmt.Sprintf("search_%s.json", queryHash)
			jobName := fmt.Sprintf("search_%s", queryHash)

			if backgroundSearch {
				// BACKGROUND WORKER
				service, err := setupGitHubClient()
				if err != nil {
					service = gh.NewUnauthenticatedService(wf.CacheDir())
				}

				repos, _, err := service.SearchRepositories(query)
				if err != nil {
					log.Printf("[background] Search failed: %v", err)
					return
				}

				if err := wf.Cache.StoreJSON(cacheName, repos); err != nil {
					log.Printf("[background] Storing cache failed: %v", err)
				}
				return
			}

			// INTERACTIVE SCRIPT FILTER
			var repos []*github.Repository
			cacheExists := wf.Cache.Exists(cacheName)
			cacheExpired := wf.Cache.Expired(cacheName, 10*time.Minute)
			jobRunning := wf.IsRunning(jobName)

			if cacheExists && !cacheExpired {
				// Cache is fresh! Load and display results instantly
				if err := wf.Cache.LoadJSON(cacheName, &repos); err != nil {
					wf.FatalError(err)
				}
			} else {
				// Cache is missing or expired
				if !jobRunning {
					// Spawn background worker to fetch fresh results
					cmd := exec.Command(os.Args[0], "search", query, "--background")
					if err := wf.RunInBackground(jobName, cmd); err != nil {
						wf.FatalError(err)
					}
					jobRunning = true
				}

				if jobRunning {
					// Tell Alfred to rerun the script in 0.25 seconds
					wf.Rerun(0.25)

					// If we have an expired cache, show it so the screen isn't blank
					if cacheExists {
						if err := wf.Cache.LoadJSON(cacheName, &repos); err != nil {
							wf.FatalError(err)
						}
						// Add a subtle loader at the top
						wf.NewItem("Updating results from GitHub...").
							Icon(aw.IconInfo).
							Valid(false)
					} else {
						// Blank state loading spinner
						wf.NewItem("Searching GitHub...").
							Subtitle(fmt.Sprintf("Fetching results for '%s'...", query)).
							Icon(aw.IconInfo).
							Valid(false)
					}
				} else {
					// Job finished but no cache exists (indicates a failure)
					wf.NewItem("GitHub Search Failed").
						Subtitle("Please check your network connection or API limits").
						Icon(aw.IconError).
						Valid(false)
				}
			}

			// Display repositories (either fresh cached or old expired cached)
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
)

func init() {
	searchCmd.Flags().BoolVarP(&backgroundSearch, "background", "b", false, "run search in background and cache results")
	rootCmd.AddCommand(searchCmd)
}
