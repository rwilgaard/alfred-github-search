package cmd

import (
	"github.com/spf13/cobra"
)

// recordCmd records a use of the repository in repo_fullname for ranking.
// Alfred runs it next to the Open URL action, which runs no Go code.
var recordCmd = &cobra.Command{
	Use:    "record",
	Short:  "record a use of a repository",
	Hidden: true,
	Args:   cobra.NoArgs,
	RunE: func(_ *cobra.Command, _ []string) error {
		if cfg.RepoFullName == "" {
			return nil
		}
		if err := wf.Usage.Record(cfg.RepoFullName); err != nil {
			wf.ReportError(err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(recordCmd)
}
