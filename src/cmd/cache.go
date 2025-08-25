package cmd

import (
	"fmt"
	"log"

	"github.com/ncruces/zenity"
	"github.com/rwilgaard/alfred-github-search/src/pkg/github"
	"github.com/spf13/cobra"
)

var (
    cacheCmd = &cobra.Command{
        Use:   "cache",
        Short: "refresh cache of user repositories",
        RunE: func(cmd *cobra.Command, args []string) error {
            log.Printf("[cache] fetching repositories...")

            token, err := wf.Keychain.Get(keychainAccount)
            if err != nil {
                zerr := zenity.Error(
                    fmt.Sprintf("Error retrieving credentials from Keychain: %s", err),
                    zenity.ErrorIcon,
                )
                if zerr != nil {
                    return zerr
                }
                return err
            }

            gh := github.NewTokenGithubService(token)
            repos, err := gh.GetAllUserRepositories()
            if err != nil {
                zerr := zenity.Error(fmt.Sprintf("Repository caching failed: %s", err), zenity.ErrorIcon)
                if zerr != nil {
                    return zerr
                }
                return err
            }

            if err := wf.Cache.StoreJSON(repoCacheName, repos); err != nil {
                return err
            }

            log.Printf("[cache] repositories fetched")
            return nil
        },
    }
)

func init() {
    rootCmd.AddCommand(cacheCmd)
}
