package main

import (
	"context"

	"github.com/google/go-github/v58/github"
)

func testAuthentication(client *github.Client) (statusCode int, err error) {
    _, resp, err := client.Users.Get(context.Background(), "")
    if err != nil {
        return resp.StatusCode, err
    }
    return resp.StatusCode, nil
}

func getAllUserRepositories(client *github.Client) ([]*github.Repository, error) {
    var allRepos []*github.Repository
    params := github.RepositoryListByAuthenticatedUserOptions{
        Sort: "updated",
        ListOptions: github.ListOptions{
            PerPage: 100,
        },
    }

    for {
        repos, resp, err := client.Repositories.ListByAuthenticatedUser(context.Background(), &params)
        if err != nil {
            return nil, err
        }
        allRepos = append(allRepos, repos...)
        if resp.NextPage == 0 {
            break
        }
        params.Page = resp.NextPage
    }

    return allRepos, nil
}
