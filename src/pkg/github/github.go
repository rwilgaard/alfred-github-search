package github

import (
    "context"

    "github.com/google/go-github/v58/github"
    "github.com/gregjones/httpcache"
)

type GithubService struct {
    Client *github.Client
}

func NewGithubService() *GithubService {
    return &GithubService{
        Client: github.NewClient(
            httpcache.NewMemoryCacheTransport().Client(),
        ),
    }
}

func NewTokenGithubService(token string) *GithubService {
    return &GithubService{
        Client: github.NewClient(
            httpcache.NewMemoryCacheTransport().Client(),
        ).WithAuthToken(token),
    }
}

func (gh *GithubService) TestAuthentication() (statusCode int, err error) {
    _, resp, err := gh.Client.Users.Get(context.Background(), "")
    if err != nil {
        return resp.StatusCode, err
    }
    return resp.StatusCode, nil
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
