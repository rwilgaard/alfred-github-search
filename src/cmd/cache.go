package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var cacheCmd = &cobra.Command{
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

func fetchAndCacheRepositories() error {
	gh, err := setupGitHubClient()
	if err != nil {
		reportBackgroundError(fmt.Sprintf("Cache failed: Error retrieving credentials: %s", err))
		return err
	}

	repos, err := gh.GetAllUserRepositories()
	if err != nil {
		reportBackgroundError(fmt.Sprintf("Cache failed: Could not fetch repos: %s", err))
		return err
	}

	if err := wf.Cache.StoreJSON(repoCacheName, repos); err != nil {
		reportBackgroundError(fmt.Sprintf("Cache failed: Could not write cache file: %s", err))
		return err
	}

	return nil
}

func init() {
	rootCmd.AddCommand(cacheCmd)
}
