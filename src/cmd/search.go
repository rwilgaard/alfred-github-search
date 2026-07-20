package cmd

import (
	"crypto/md5"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
				service, authErr := setupGitHubClient()
				if authErr != nil {
					log.Printf("[background] Falling back to unauthenticated search: %v", authErr)
					service = gh.NewUnauthenticatedService(wf.CacheDir())
				}

				repos, _, err := service.SearchRepositories(query)
				if err != nil {
					msg := fmt.Sprintf("Search failed: %s", err)
					if authErr != nil {
						msg = fmt.Sprintf("%s (unauthenticated: %s)", msg, authErr)
					}
					reportBackgroundError(msg)
					log.Printf("[background] Search failed: %v", err)
					return
				}

				if err := wf.Cache.StoreJSON(cacheName, repos); err != nil {
					reportBackgroundError(fmt.Sprintf("Search failed: could not write cache: %s", err))
					log.Printf("[background] Storing cache failed: %v", err)
				}

				cleanupSearchCache(wf.CacheDir(), 24*time.Hour)
				return
			}

			var repos []*github.Repository
			cacheExists := wf.Cache.Exists(cacheName)
			cacheExpired := wf.Cache.Expired(cacheName, 10*time.Minute)
			jobRunning := wf.IsRunning(jobName)

			if cacheExists && !cacheExpired {
				if err := wf.Cache.LoadJSON(cacheName, &repos); err != nil {
					wf.FatalError(err)
				}
			} else {
				if !jobRunning {
					cmd := exec.Command(os.Args[0], "search", query, "--background")
					if err := wf.RunInBackground(jobName, cmd); err != nil {
						wf.FatalError(err)
					}
					jobRunning = true
				}

				if jobRunning {
					wf.Rerun(0.25)

					if cacheExists {
						if err := wf.Cache.LoadJSON(cacheName, &repos); err != nil {
							wf.FatalError(err)
						}
						wf.NewItem("Updating results from GitHub...").
							Icon(aw.IconInfo).
							Valid(false)
					} else {
						wf.NewItem("Searching GitHub...").
							Subtitle(fmt.Sprintf("Fetching results for '%s'...", query)).
							Icon(aw.IconInfo).
							Valid(false)
					}
				} else {
					wf.NewItem("GitHub Search Failed").
						Subtitle("Please check your network connection or API limits").
						Icon(aw.IconError).
						Valid(false)
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

			alfredutils.HandleFeedback(wf)
		},
	}
)

func init() {
	searchCmd.Flags().BoolVarP(&backgroundSearch, "background", "b", false, "run search in background and cache results")
	rootCmd.AddCommand(searchCmd)
}

func cleanupSearchCache(cacheDir string, maxAge time.Duration) {
	files, err := os.ReadDir(cacheDir)
	if err != nil {
		log.Printf("[cleanup] Failed to read cache directory: %v", err)
		return
	}

	now := time.Now()
	for _, file := range files {
		if file.Type().IsRegular() && strings.HasPrefix(file.Name(), "search_") && strings.HasSuffix(file.Name(), ".json") {
			info, err := file.Info()
			if err != nil {
				continue
			}
			if now.Sub(info.ModTime()) > maxAge {
				filePath := filepath.Join(cacheDir, file.Name())
				if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
					log.Printf("[cleanup] Failed to delete expired cache file %s: %v", file.Name(), err)
				} else if err == nil {
					log.Printf("[cleanup] Deleted expired cache file: %s", file.Name())
				}
			}
		}
	}
}
