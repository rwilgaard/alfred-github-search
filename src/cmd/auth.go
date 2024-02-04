package cmd

import (
    "fmt"

    "github.com/ncruces/zenity"
    "github.com/rwilgaard/alfred-github-search/src/pkg/github"
    "github.com/spf13/cobra"
)

var (
    authCmd = &cobra.Command{
        Use:   "auth",
        Short: "authenticate",
        RunE: func(cmd *cobra.Command, args []string) error {
            _, pw, err := zenity.Password(zenity.Title("Enter API Token"))
            if err != nil {
                return err
            }

            gh := github.NewTokenGithubService(pw)
            sc, err := gh.TestAuthentication()
            if err != nil {
                zerr := zenity.Error(
                    fmt.Sprintf("Error authenticating: HTTP %d", sc),
                    zenity.ErrorIcon,
                )
                if zerr != nil {
                    return err
                }
                return err
            }

            if err := wf.Keychain.Set(keychainAccount, pw); err != nil {
                return err
            }
            return nil
        },
    }
)

func init() {
    rootCmd.AddCommand(authCmd)
}
