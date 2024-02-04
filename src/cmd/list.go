package cmd

import (
    "fmt"
    "time"

    "github.com/google/go-github/v58/github"
    "github.com/rwilgaard/alfred-github-search/src/pkg/alfred"
    "github.com/spf13/cobra"
)

var (
    listCmd = &cobra.Command{
        Use:   "list",
        Short: "list user repositories",
        Args:  cobra.ExactArgs(1),
        Run: func(cmd *cobra.Command, args []string) {
            alfred.HandleAuthentication(wf, keychainAccount)

            var repos []*github.Repository
            err := alfred.LoadCache(wf, repoCacheName, &repos)
            if err != nil {
                wf.FatalError(err)
            }

            maxCacheAge := time.Duration(cfg.CacheAge * int(time.Minute))
            if err := alfred.RefreshCache(wf, repoCacheName, maxCacheAge); err != nil {
                wf.FatalError(err)
            }

            for _, repo := range repos {
                updatedAt := repo.UpdatedAt.Time.Format("02-01-2006 15:04")
                wf.NewItem(*repo.FullName).
                    UID(*repo.FullName).
                    Subtitle(fmt.Sprintf("%s  •  Updated: %s", *repo.HTMLURL, updatedAt)).
                    Var("item_url", *repo.HTMLURL).
                    Arg("repo").
                    Valid(true)
            }

            query := args[0]
            if len(query) > 0 {
                wf.Filter(query)
            }
            alfred.HandleFeedback(wf)
        },
    }
)

func init() {
    rootCmd.AddCommand(listCmd)
}
