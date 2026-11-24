package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:          "update",
	Short:        "check for updates",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, _ []string) error {
		log.Println("Checking for updates...")
		return updater.Check(cmd.Context())
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
}
