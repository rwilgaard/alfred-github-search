package cmd

import (
    "fmt"

    "github.com/ncruces/zenity"
    "github.com/rwilgaard/alfred-github-search/src/internal/github"
    "github.com/spf13/cobra"
)

var (
    authCmd = &cobra.Command{
        Use:   "auth",
        Short: "authenticate",
        RunE: func(cmd *cobra.Command, args []string) error {
            _, token, err := zenity.Password(zenity.Title("Enter API Token"))
            if err != nil {
                return err
            }

            gh := github.NewAuthenticatedService(token)
			if err := gh.TestAuthentication(); err != nil {
                zerr := zenity.Error(
                    fmt.Sprintf("Error authenticating: %s", err),
                    zenity.ErrorIcon,
                )
                if zerr != nil {
                    return err
                }
                return err
            }

            if err := wf.Keychain.Set(keychainAccount, token); err != nil {
                return err
            }
            return nil
        },
    }
)

func init() {
    rootCmd.AddCommand(authCmd)
}
