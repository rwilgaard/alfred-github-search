package cmd

import (
	"fmt"
	"strings"

	"github.com/rwilgaard/alfred-github-search/src/internal/util"
	"github.com/rwilgaard/go-alfredutils/alfredutils"
	"github.com/spf13/cobra"
)

var detailsCmd = &cobra.Command{
	Use:   "details",
	Short: "list actions for a repository",
	Run: func(_ *cobra.Command, _ []string) {
		repoName := cfg.RepoFullName
		if repoName == "" {
			wf.FatalError(fmt.Errorf("repo_fullname variable is empty"))
		}

		parts := strings.Split(repoName, "/")
		repoShortName := parts[len(parts)-1]

		wf.NewItem("Open in Browser").
			Subtitle("Open " + repoName + " home page in your browser").
			Icon(util.GetIcon("browser")).
			Arg("open_browser").
			Valid(true)

		wf.NewItem("View Open Issues").
			Subtitle("Show issues for " + repoShortName).
			Icon(util.GetIcon("issue-open")).
			Arg("view_issues").
			Valid(true)

		wf.NewItem("View Open Pull Requests").
			Subtitle("Show pull requests for " + repoShortName).
			Icon(util.GetIcon("pr-open")).
			Arg("view_prs").
			Valid(true)

		if cfg.RepoCloneURL != "" {
			wf.NewItem("Copy HTTPS Clone URL").
				Subtitle(cfg.RepoCloneURL).
				Icon(util.GetIcon("copy")).
				Var("action", "copy").
				Var("notification_title", "HTTPS Clone URL Copied").
				Var("notification_text", cfg.RepoFullName).
				Arg(cfg.RepoCloneURL).
				Valid(true)
		}

		if cfg.RepoSSHURL != "" {
			wf.NewItem("Copy SSH Clone URL").
				Subtitle(cfg.RepoSSHURL).
				Icon(util.GetIcon("copy")).
				Var("action", "copy").
				Var("notification_title", "SSH Clone URL Copied").
				Var("notification_text", cfg.RepoFullName).
				Arg(cfg.RepoSSHURL).
				Valid(true)
		}

		wf.NewItem("Back to Repositories").
			Subtitle("Return to your repository list").
			Icon(util.GetIcon("back")).
			Var("list_query", cfg.ListQuery).
			Arg("back").
			Valid(true)

		alfredutils.HandleFeedback(wf)
	},
}

func init() {
	rootCmd.AddCommand(detailsCmd)
}
