package cmd

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/google/go-github/v89/github"
	"github.com/maniartech/gotime/v2"
	gh "github.com/rwilgaard/alfred-github-search/src/internal/github"
	"github.com/rwilgaard/go-alfredutils/v2/alfred/clipboard"
	alfredui "github.com/rwilgaard/go-alfredutils/v2/alfred/ui"
	"github.com/spf13/cobra"
)

// Component ids in workflow/repo.alfredui/ui.json.
const (
	uiTitle    = "title"
	uiSubtitle = "subtitle"
	uiMeta     = "meta"
	uiItems    = "items"
	uiFilter   = "filter"
	uiLoading  = "loading"
	uiIssues   = "issuesbtn"
	uiPulls    = "pullsbtn"
)

const (
	kindIssues = "issues"
	kindPulls  = "pulls"
)

// uiRepoCmd fills in the repo window when it opens: the header from the
// repository and the issue list.
var uiRepoCmd = &cobra.Command{
	Use:    "ui-repo",
	Short:  "fill in the repository window",
	Hidden: true,
	Run: func(_ *cobra.Command, _ []string) {
		emit(uiRepo(cfg.RepoFullName))
	},
}

// uiRepoItemsCmd switches the list to issues or pull requests. The window
// fetched both on open, so this is answered from the HTTP cache.
var uiRepoItemsCmd = &cobra.Command{
	Use:       "ui-repo-items issues|pulls",
	Short:     "show issues or pull requests in the repository window",
	Hidden:    true,
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{kindIssues, kindPulls},
	Run: func(_ *cobra.Command, args []string) {
		emit(uiRepoItems(cfg.RepoFullName, args[0]))
	},
}

// uiCopyCmd copies a clone URL and keeps the window open.
var uiCopyCmd = &cobra.Command{
	Use:       "ui-copy https|ssh",
	Short:     "copy a clone URL from the repository window",
	Hidden:    true,
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"https", "ssh"},
	Run: func(_ *cobra.Command, args []string) {
		emit(uiCopy(args[0]))
	},
}

// uiService returns an authenticated client, or an unauthenticated one for
// public repositories found with the global search before logging in.
func uiService() (*gh.GithubService, error) {
	service, err := setupGitHubClient(listMaxAge)
	if err == nil {
		return service, nil
	}
	log.Printf("[ui] falling back to unauthenticated client: %v", err)
	return gh.NewUnauthenticatedService(wf.Cache, listMaxAge)
}

func uiRepo(fullName string) alfredui.Actions {
	a := alfredui.NewUpdate()
	owner, name, err := splitRepoName(fullName)
	if err != nil {
		return failLoading(a, err)
	}
	service, err := uiService()
	if err != nil {
		return failLoading(a, err)
	}

	var (
		wg         sync.WaitGroup
		repo       *github.Repository
		issues     []*github.Issue
		rErr, iErr error
	)
	wg.Go(func() { repo, rErr = service.GetRepository(owner, name) })
	wg.Go(func() { issues, iErr = service.GetIssues(owner, name) })
	// Fetched only to fill the HTTP cache, so ⌘2 shows them without waiting.
	wg.Go(func() {
		if _, err := service.GetPullRequests(owner, name); err != nil {
			log.Printf("[ui] prefetch pull requests: %v", err)
		}
	})
	wg.Wait()

	if rErr != nil {
		return failLoading(a, rErr)
	}
	if err := wf.Usage.Record(fullName); err != nil {
		log.Printf("[ui] record usage of %s: %v", fullName, err)
	}

	a = a.Set(uiTitle, alfredui.Props{"text": repo.GetFullName()}).
		Set(uiSubtitle, alfredui.Props{"text": orSpace(repo.GetDescription())}).
		Set(uiMeta, alfredui.Props{"text": repoMeta(repo)})
	a.Update.Relayout = true // the header labels get their width from this text

	if iErr != nil {
		return failLoading(a, iErr)
	}
	return showRows(a, kindIssues, issueRows(issues))
}

func uiRepoItems(fullName, kind string) alfredui.Actions {
	a := alfredui.NewUpdate()
	owner, name, err := splitRepoName(fullName)
	if err != nil {
		return failLoading(a, err)
	}
	service, err := uiService()
	if err != nil {
		return failLoading(a, err)
	}

	switch kind {
	case kindPulls:
		prs, err := service.GetPullRequests(owner, name)
		if err != nil {
			return failLoading(a, err)
		}
		return showRows(a, kind, pullRows(prs))
	default:
		issues, err := service.GetIssues(owner, name)
		if err != nil {
			return failLoading(a, err)
		}
		return showRows(a, kindIssues, issueRows(issues))
	}
}

func uiCopy(which string) alfredui.Actions {
	value, label := cfg.RepoCloneURL, "HTTPS URL"
	if which == "ssh" {
		value, label = cfg.RepoSSHURL, "SSH URL"
	}
	if value == "" {
		return status("No "+label+" for this repository", alfredui.ColourError)
	}
	if _, err := clipboard.Copy(value); err != nil {
		return status(err.Error(), alfredui.ColourError)
	}
	return status("✓ Copied "+label, alfredui.ColourSuccess)
}

// status sets the status label. It relays out the window: a label keeps the
// width of its first layout, which for the status label is a single space.
func status(msg, colour string) alfredui.Actions {
	a := alfredui.NewUpdate().Status(uiStatus, msg, colour)
	a.Update.Relayout = true
	return a
}

func failLoading(a alfredui.Actions, err error) alfredui.Actions {
	log.Printf("[ui] %v", err)
	a = a.Set(uiLoading, alfredui.Props{"hidden": true}).
		Status(uiStatus, apiMessage(err), alfredui.ColourError)
	a.Update.Relayout = true // see status
	return a
}

// showRows puts rows in the list and marks which list it is: the active
// toggle button is bold, and the filter's placeholder names the list, its
// count and the shortcut to the other one. The filter is sent with the data
// so it survives the update.
func showRows(a alfredui.Actions, kind string, rows []alfredui.Props) alfredui.Actions {
	on, off := uiIssues, uiPulls
	noun, other := "issues", "⌘2 pull requests"
	if kind == kindPulls {
		on, off = uiPulls, uiIssues
		noun, other = "pull requests", "⌘1 issues"
	}

	loading := alfredui.Props{"hidden": true}
	if len(rows) == 0 {
		loading = alfredui.Props{"hidden": false, "text": "No open " + noun}
	}
	placeholder := fmt.Sprintf("Filter %s open %s   ·   %s   ·   ↑↓ select   ↩ open",
		countText(len(rows), 50), noun, other)

	return a.Set(uiItems, alfredui.Props{"data": map[string]any{
		"type": "basic",
		"src":  rows,
		"filter": map[string]any{
			"type":  "basic",
			"args":  []string{"rowtitle.text", "rowsubtitle.text"},
			"match": "{value:filter.text}",
		},
	}}).
		Set(on, alfredui.Props{"font": map[string]any{"weight": "bold"}}).
		Set(off, alfredui.Props{"font": nil}).
		Set(uiFilter, alfredui.Props{"placeholder": placeholder}).
		Set(uiLoading, loading).
		Status(uiStatus, " ", alfredui.ColourInfo)
}

func issueRows(issues []*github.Issue) []alfredui.Props {
	rows := make([]alfredui.Props, 0, len(issues))
	for _, iss := range issues {
		parts := []string{
			fmt.Sprintf("#%d", iss.GetNumber()),
			"by " + iss.GetUser().GetLogin(),
			timeAgo(iss.UpdatedAt),
		}
		switch n := iss.GetComments(); n {
		case 0:
		case 1:
			parts = append(parts, "1 comment")
		default:
			parts = append(parts, fmt.Sprintf("%d comments", n))
		}
		if labels := issueLabels(iss.Labels); labels != "" {
			parts = append(parts, labels)
		}
		rows = append(rows, row(iss.GetTitle(), strings.Join(parts, "  ·  "), iss.GetHTMLURL()))
	}
	return rows
}

func pullRows(prs []*github.PullRequest) []alfredui.Props {
	rows := make([]alfredui.Props, 0, len(prs))
	for _, pr := range prs {
		parts := []string{
			fmt.Sprintf("#%d", pr.GetNumber()),
			"by " + pr.GetUser().GetLogin(),
			timeAgo(pr.UpdatedAt),
			pr.GetHead().GetRef(),
		}
		if pr.GetDraft() {
			parts = append(parts, "Draft")
		}
		rows = append(rows, row(pr.GetTitle(), strings.Join(parts, "  ·  "), pr.GetHTMLURL()))
	}
	return rows
}

func row(title, subtitle, url string) alfredui.Props {
	return alfredui.Props{
		"rowtitle.text":    title,
		"rowsubtitle.text": subtitle,
		"url.text":         url,
	}
}

func issueLabels(labels []*github.Label) string {
	names := make([]string, 0, len(labels))
	for _, l := range labels {
		names = append(names, l.GetName())
	}
	return strings.Join(names, ", ")
}

func repoMeta(r *github.Repository) string {
	parts := []string{"Public"}
	if r.GetPrivate() {
		parts[0] = "Private"
	}
	if r.GetArchived() {
		parts = append(parts, "Archived")
	}
	if r.GetFork() {
		parts = append(parts, "Fork")
	}
	if lang := r.GetLanguage(); lang != "" {
		parts = append(parts, lang)
	}
	parts = append(parts,
		fmt.Sprintf("★ %d", r.GetStargazersCount()),
		fmt.Sprintf("%d forks", r.GetForksCount()),
		"pushed "+strings.ToLower(timeAgo(r.PushedAt)),
	)
	if topics := r.Topics; len(topics) > 0 {
		parts = append(parts, strings.Join(topics, ", "))
	}
	return strings.Join(parts, "  ·  ")
}

func timeAgo(t *github.Timestamp) string {
	if t == nil {
		return "never"
	}
	return gotime.TimeAgo(t.Time)
}

// countText shows "50+" when a list hit the page limit.
func countText(n, limit int) string {
	if n >= limit {
		return fmt.Sprintf("%d+", limit)
	}
	return fmt.Sprint(n)
}

// orSpace keeps a label's height when its text is empty.
func orSpace(s string) string {
	if s == "" {
		return " "
	}
	return s
}

func init() {
	rootCmd.AddCommand(uiRepoCmd, uiRepoItemsCmd, uiCopyCmd)
}
