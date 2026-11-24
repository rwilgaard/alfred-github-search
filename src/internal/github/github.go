package github

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v89/github"
	"github.com/rwilgaard/go-alfredutils/v2/alfred"
	"github.com/rwilgaard/go-alfredutils/v2/alfred/httpcache"
)

type GithubService struct {
	client *github.Client
}

// newClient stores GET responses in store: younger than maxAge they are used
// without a request, older ones are revalidated with ETags (GitHub doesn't
// count a 304 against the rate limit). The token is set outside the cache,
// so cached responses are keyed by it.
func newClient(store *alfred.Store, maxAge time.Duration, opts ...github.ClientOptionsFunc) (*GithubService, error) {
	transport := &httpcache.Transport{Store: store, MaxAge: maxAge, StaleIfError: true}
	opts = append([]github.ClientOptionsFunc{github.WithTransport(transport)}, opts...)
	client, err := github.NewClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}
	return &GithubService{client: client}, nil
}

func NewUnauthenticatedService(store *alfred.Store, maxAge time.Duration) (*GithubService, error) {
	return newClient(store, maxAge)
}

func NewAuthenticatedService(token string, store *alfred.Store, maxAge time.Duration) (*GithubService, error) {
	return newClient(store, maxAge, github.WithAuthToken(token))
}

// TestAuthentication verifies the token by fetching the authenticated user.
func (gh *GithubService) TestAuthentication() error {
	_, _, err := gh.client.Users.Get(context.Background(), "")
	return err
}

func (gh *GithubService) SearchRepositories(query string) ([]*github.Repository, error) {
	result, _, err := gh.client.Search.Repositories(context.Background(), query, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to search repositories: %w", err)
	}
	return result.Repositories, nil
}

func (gh *GithubService) GetAllUserRepositories() ([]*github.Repository, error) {
	var allRepos []*github.Repository
	opts := github.RepositoryListByAuthenticatedUserOptions{
		Sort: "updated",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	for {
		repos, resp, err := gh.client.Repositories.ListByAuthenticatedUser(context.Background(), &opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list repositories: %w", err)
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allRepos, nil
}

func (gh *GithubService) GetIssues(owner, repo string) ([]*github.Issue, error) {
	const want = 50
	opts := &github.IssueListByRepoOptions{
		State: "open",
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	var list []*github.Issue
	for len(list) < want {
		issues, resp, err := gh.client.Issues.ListByRepo(context.Background(), owner, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list issues: %w", err)
		}
		for _, iss := range issues {
			if !iss.IsPullRequest() {
				list = append(list, iss)
			}
		}
		if resp.NextPage == 0 {
			break
		}
		opts.ListOptions.Page = resp.NextPage
	}

	if len(list) > want {
		list = list[:want]
	}
	return list, nil
}

func (gh *GithubService) GetPullRequests(owner, repo string) ([]*github.PullRequest, error) {
	opts := &github.PullRequestListOptions{
		State: "open",
		ListOptions: github.ListOptions{
			PerPage: 50,
		},
	}
	prs, _, err := gh.client.PullRequests.List(context.Background(), owner, repo, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list pull requests: %w", err)
	}
	return prs, nil
}

func (gh *GithubService) GetRepository(owner, repo string) (*github.Repository, error) {
	r, _, err := gh.client.Repositories.Get(context.Background(), owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}
	return r, nil
}
