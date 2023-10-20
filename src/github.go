package main

import (
    "context"

    "github.com/google/go-github/v56/github"
)

func getAllUserRepositories(client *github.Client) ([]*github.Repository, error) {
    var allRepos []*github.Repository
    params := github.RepositoryListOptions{
        Sort: "updated",
        ListOptions: github.ListOptions{
            PerPage: 100,
        },
    }

    for {
        repos, resp, err := client.Repositories.List(context.Background(), "", &params)
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
