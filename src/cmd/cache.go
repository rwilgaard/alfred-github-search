package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var (
	cacheCmd = &cobra.Command{
		Use:   "cache",
		Short: "refresh cache of user repositories",
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Printf("[cache] fetching repositories...")

			if err := fetchAndCacheRepositories(); err != nil {
				return err
			}

			log.Printf("[cache] repositories fetched")
			return nil
		},
	}
)

func fetchAndCacheRepositories() error {
	gh, err := setupGitHubClient()
	if err != nil {
		msg := fmt.Sprintf("Cache failed: Error retrieving credentials: %s", err)
		if zerr := wf.Alfred.RunTrigger("error", msg); zerr != nil {
			log.Printf("Alfred error trigger failed: %v", zerr)
		}
		return err
	}

	repos, err := gh.GetAllUserRepositories()
	if err != nil {
		msg := fmt.Sprintf("Cache failed: Could not fetch repos: %s", err)
		if zerr := wf.Alfred.RunTrigger("error", msg); zerr != nil {
			log.Printf("Alfred error trigger failed: %v", zerr)
		}
		return err
	}

	if err := wf.Cache.StoreJSON(repoCacheName, repos); err != nil {
		msg := fmt.Sprintf("Cache failed: Could not write cache file: %s", err)
		if zerr := wf.Alfred.RunTrigger("error", msg); zerr != nil {
			log.Printf("Alfred error trigger failed: %v", zerr)
		}
		return err
	}

	return nil
}

func init() {
	rootCmd.AddCommand(cacheCmd)
}
