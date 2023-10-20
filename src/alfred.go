package main

import (
    "context"
    "fmt"
    "log"
    "time"

    aw "github.com/deanishe/awgo"
    "github.com/google/go-github/v56/github"
    "github.com/ncruces/zenity"
)

type magicAuth struct {
    wf *aw.Workflow
}

func (a magicAuth) Keyword() string     { return "clearauth" }
func (a magicAuth) Description() string { return "Clear credentials for GitHub." }
func (a magicAuth) RunText() string     { return "Credentials cleared!" }
func (a magicAuth) Run() error          { return clearAuth() }

func clearAuth() error {
    if err := wf.Keychain.Delete(keychainAccount); err != nil {
        return err
    }
    return nil
}

func cacheRepositories(client *github.Client) error {
    log.Printf("[cache] fetching repositories...")

    repos, err := getAllUserRepositories(client)
    if err != nil {
        return err
    }

    if err := wf.Cache.StoreJSON(repoCacheName, repos); err != nil {
        return err
    }

    log.Printf("[cache] repositories fetched")
    return nil
}

func runUserRepos(client *github.Client) {
    var repos []*github.Repository
    if wf.Cache.Exists(repoCacheName) {
        if err := wf.Cache.LoadJSON(repoCacheName, &repos); err != nil {
            wf.FatalError(err)
        }
    }

    maxCacheAge := cfg.CacheAge * int(time.Minute)
    if wf.Cache.Expired(repoCacheName, time.Duration(maxCacheAge)) {
        if err := cacheRepositories(client); err != nil {
            wf.FatalError(err)
        }
        wf.Rerun(0.3)
    }

    for _, repo := range repos {
        wf.NewItem(*repo.FullName).
            Subtitle(fmt.Sprintf("%s  •  Updated: %s", *repo.HTMLURL, repo.UpdatedAt.String())).
            Var("item_url", *repo.HTMLURL).
            Arg("repo").
            Valid(true)
    }
}

func runSearch(client *github.Client) {
    repos, _, err := client.Search.Repositories(context.Background(), opts.Query, nil)
    if err != nil {
        wf.FatalError(err)
    }

    for _, repo := range repos.Repositories {
        wf.NewItem(*repo.FullName).
            Subtitle(fmt.Sprintf("%s  •  Updated: %s", *repo.HTMLURL, repo.UpdatedAt.String())).
            Var("item_url", *repo.HTMLURL).
            Arg("repo").
            Valid(true)
    }
}

func runAuth() {
    _, pwd, err := zenity.Password(
        zenity.Title(fmt.Sprintf("Enter API Token for %s", cfg.Username)),
    )
    if err != nil {
        wf.FatalError(err)
    }
    if err := wf.Keychain.Set(keychainAccount, pwd); err != nil {
        wf.FatalError(err)
    }
}
