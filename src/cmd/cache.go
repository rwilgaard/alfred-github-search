package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/rwilgaard/go-alfredutils/v2/alfred/httpcache"
	"github.com/spf13/cobra"
)

// Stored API responses older than this are deleted by the cache job. Search
// stores one per query, so without pruning the folder grows without bound.
const httpCacheTTL = 24 * time.Hour

// cacheCmd runs in the background, started by the repos Script Filter. Alfred
// isn't listening for its output, so failures go through wf.ReportError.
var cacheCmd = &cobra.Command{
	Use:    "cache",
	Short:  "refresh the cache of user repositories",
	Hidden: true,
	RunE: func(_ *cobra.Command, _ []string) error {
		log.Println("[cache] fetching repositories...")
		if err := fetchAndCacheRepositories(); err != nil {
			wf.ReportError(fmt.Errorf("cache failed: %w", err))
			return nil
		}
		log.Println("[cache] repositories fetched")

		pruneHTTPCache(httpCacheTTL)
		removeLegacyCache()
		return nil
	},
}

func fetchAndCacheRepositories() error {
	// MaxAge 0: always revalidate, so the list is current; unchanged pages
	// come back as 304s.
	service, err := setupGitHubClient(0)
	if err != nil {
		return fmt.Errorf("error retrieving credentials: %w", err)
	}

	repos, err := service.GetAllUserRepositories()
	if err != nil {
		return fmt.Errorf("could not fetch repos: %w", err)
	}

	if err := wf.Cache.StoreJSON(repoCacheName, repos); err != nil {
		return fmt.Errorf("could not write cache: %w", err)
	}
	return nil
}

func pruneHTTPCache(maxAge time.Duration) {
	dir := wf.Cache.Path(httpcache.DefaultDir)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("[cache] read %s: %v", dir, err)
		}
		return
	}
	for _, e := range entries {
		key := httpcache.DefaultDir + "/" + e.Name()
		if e.Type().IsRegular() && wf.Cache.Expired(key, maxAge) {
			if err := wf.Cache.Delete(key); err != nil {
				log.Printf("[cache] delete %s: %v", key, err)
			}
		}
	}
}

// removeLegacyCache deletes what versions before 1.3 cached: a disk cache of
// HTTP responses and one file of results per search query.
func removeLegacyCache() {
	if err := os.RemoveAll(wf.Cache.Path("http-cache")); err != nil {
		log.Printf("[cache] remove legacy http-cache: %v", err)
	}
	entries, err := os.ReadDir(wf.Cache.Path(""))
	if err != nil {
		return
	}
	for _, e := range entries {
		if name := e.Name(); strings.HasPrefix(name, "search_") {
			if err := wf.Cache.Delete(name); err != nil {
				log.Printf("[cache] delete %s: %v", name, err)
			}
		}
	}
}

func init() {
	rootCmd.AddCommand(cacheCmd)
}
