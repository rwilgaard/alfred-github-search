package github

import (
	"context"
	"net/http"

	"github.com/google/go-github/v78/github"
	"github.com/gregjones/httpcache"
	"golang.org/x/oauth2"
)

type GithubService struct {
    Client *github.Client
}

func newCachedClient(baseTransport http.RoundTripper) *http.Client {
	cacheTransport := httpcache.NewMemoryCacheTransport()
	cacheTransport.Transport = baseTransport

	return cacheTransport.Client()
}

func NewUnauthenticatedService() *GithubService {
	httpClient := newCachedClient(nil)

	return &GithubService{
		Client: github.NewClient(httpClient),
	}
}

func NewAuthenticatedService(token string) *GithubService {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})

	oauthTransport := &oauth2.Transport{
		Source: ts,
	}

	httpClient := newCachedClient(oauthTransport)
	client := github.NewClient(httpClient)

	return &GithubService{
		Client: client,
	}
}

func (gh *GithubService) TestAuthentication() error {
    _, _, err := gh.Client.Users.Get(context.Background(), "")
    return err
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
        repos, resp, err := gh.Client.Repositories.ListByAuthenticatedUser(context.Background(), &opts)
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
