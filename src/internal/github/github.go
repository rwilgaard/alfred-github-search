package github

import (
	"context"
	"net/http"
	"path/filepath"

	"github.com/google/go-github/v89/github"
	"github.com/gregjones/httpcache"
	"github.com/gregjones/httpcache/diskcache"
	"golang.org/x/oauth2"
)

type GithubService struct {
	client *github.Client
}

func newCachedClient(baseTransport http.RoundTripper, cacheDir string) *http.Client {
	var cacheTransport *httpcache.Transport

	if cacheDir != "" {
		diskDir := filepath.Join(cacheDir, "http-cache")
		cache := diskcache.New(diskDir)
		cacheTransport = httpcache.NewTransport(cache)
	} else {
		cacheTransport = httpcache.NewMemoryCacheTransport()
	}

	cacheTransport.Transport = baseTransport
	return cacheTransport.Client()
}

func NewUnauthenticatedService(cacheDir string) *GithubService {
	httpClient := newCachedClient(nil, cacheDir)
	client, err := github.NewClient(github.WithHTTPClient(httpClient))
	if err != nil {
		panic(err)
	}

	return &GithubService{
		client: client,
	}
}

func NewAuthenticatedService(token string, cacheDir string) *GithubService {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})

	oauthTransport := &oauth2.Transport{
		Source: ts,
	}

	httpClient := newCachedClient(oauthTransport, cacheDir)
	client, err := github.NewClient(github.WithHTTPClient(httpClient))
	if err != nil {
		panic(err)
	}

	return &GithubService{
		client: client,
	}
}

func (gh *GithubService) TestAuthentication() error {
	_, _, err := gh.client.Users.Get(context.Background(), "")
	return err
}

func (gh *GithubService) SearchRepositories(query string) ([]*github.Repository, *github.Response, error) {
	result, resp, err := gh.client.Search.Repositories(context.Background(), query, nil)
	if err != nil {
		return nil, resp, err
	}
	return result.Repositories, resp, nil
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
			return nil, err
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
			return nil, err
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
		return nil, err
	}
	return prs, nil
}
