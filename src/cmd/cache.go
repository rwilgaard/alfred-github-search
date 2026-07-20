package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	gh "github.com/rwilgaard/alfred-github-search/src/internal/github"
)

// cacheCmd and its subcommands are invoked in the background by the Script
// Filter commands, never by the user. Alfred isn't listening for their output,
// so failures are reported via reportBackgroundError rather than wf.FatalError.
var cacheCmd = &cobra.Command{
	Use:    "cache",
	Short:  "refresh cached data in the background",
	Hidden: true,
}

var cacheReposCmd = &cobra.Command{
	Use:   "repos",
	Short: "refresh the cache of user repositories",
	RunE: func(_ *cobra.Command, _ []string) error {
		log.Printf("[cache] fetching repositories...")

		if err := fetchAndCacheRepositories(); err != nil {
			return err
		}

		log.Printf("[cache] repositories fetched")
		return nil
	},
}

var cacheSearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "run a repository search and cache the results",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		query := args[0]
		log.Printf("[cache] searching for %q...", query)

		if err := fetchAndCacheSearch(query); err != nil {
			return err
		}

		cleanupSearchCache(wf.CacheDir(), searchCacheTTL)
		log.Printf("[cache] search results fetched")
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

func fetchAndCacheSearch(query string) error {
	// Public search works without a token, so an auth failure degrades to an
	// unauthenticated (rate-limited) client rather than failing outright.
	service, authErr := setupGitHubClient()
	if authErr != nil {
		log.Printf("[cache] falling back to unauthenticated search: %v", authErr)
		service = gh.NewUnauthenticatedService(wf.CacheDir())
	}

	repos, _, err := service.SearchRepositories(query)
	if err != nil {
		msg := fmt.Sprintf("Search failed: %s", err)
		if authErr != nil {
			msg = fmt.Sprintf("%s (unauthenticated: %s)", msg, authErr)
		}
		reportBackgroundError(msg)
		return err
	}

	if err := wf.Cache.StoreJSON(searchCacheName(query), repos); err != nil {
		reportBackgroundError(fmt.Sprintf("Search failed: could not write cache: %s", err))
		return err
	}

	return nil
}

// Each distinct query gets its own cache file, so without this the cache
// directory grows without bound.
func cleanupSearchCache(cacheDir string, maxAge time.Duration) {
	files, err := os.ReadDir(cacheDir)
	if err != nil {
		log.Printf("[cleanup] Failed to read cache directory: %v", err)
		return
	}

	now := time.Now()
	for _, file := range files {
		if !file.Type().IsRegular() ||
			!strings.HasPrefix(file.Name(), searchCachePrefix) ||
			!strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		info, err := file.Info()
		if err != nil {
			continue
		}
		if now.Sub(info.ModTime()) <= maxAge {
			continue
		}

		filePath := filepath.Join(cacheDir, file.Name())
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			log.Printf("[cleanup] Failed to delete expired cache file %s: %v", file.Name(), err)
			continue
		}
		log.Printf("[cleanup] Deleted expired cache file: %s", file.Name())
	}
}

func init() {
	cacheCmd.AddCommand(cacheReposCmd)
	cacheCmd.AddCommand(cacheSearchCmd)
	rootCmd.AddCommand(cacheCmd)
}
