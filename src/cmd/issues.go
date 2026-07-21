package cmd

import (
	"fmt"
	"strings"

	"github.com/maniartech/gotime/v2"
	"github.com/rwilgaard/alfred-github-search/src/internal/util"
	"github.com/rwilgaard/go-alfredutils/alfredutils"
	"github.com/spf13/cobra"
)

var issuesCmd = &cobra.Command{
	Use:   "issues [query]",
	Short: "list open issues for a repository",
	Args:  cobra.MaximumNArgs(1),
	Run: func(_ *cobra.Command, args []string) {
		repoName := cfg.RepoFullName
		if repoName == "" {
			wf.FatalError(fmt.Errorf("repo_fullname variable is empty"))
		}

		parts := strings.Split(repoName, "/")
		if len(parts) != 2 {
			wf.FatalError(fmt.Errorf("invalid repo_fullname: %s", repoName))
		}
		owner, repo := parts[0], parts[1]

		wf.NewItem("Back to Actions").
			Subtitle("Return to the actions menu for " + repo).
			Icon(util.GetIcon("back")).
			Arg("back_to_details").
			Valid(true)

		service, err := setupGitHubClient()
		if err != nil {
			wf.FatalError(err)
		}

		issues, err := service.GetIssues(owner, repo)
		if err != nil {
			wf.FatalError(err)
		}

		for _, iss := range issues {
			updated := "never"
			if iss.UpdatedAt != nil {
				updated = gotime.TimeAgo(*iss.UpdatedAt.GetTime())
			}
			subtitle := fmt.Sprintf("#%d  ·  By %s  ·  %s", iss.GetNumber(), iss.GetUser().GetLogin(), updated)

			wf.NewItem(iss.GetTitle()).
				Subtitle(subtitle).
				Icon(util.GetIcon("issue-open")).
				Var("item_url", iss.GetHTMLURL()).
				Arg("issue").
				Valid(true)
		}

		var query string
		if len(args) > 0 {
			query = args[0]
		}
		if len(query) > 0 {
			wf.Filter(query)
		}

		alfredutils.HandleFeedback(wf)
	},
}

func init() {
	rootCmd.AddCommand(issuesCmd)
}
