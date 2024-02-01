package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "os/exec"
    "strings"
    "time"

    aw "github.com/deanishe/awgo"
    "github.com/google/go-github/v58/github"
    "github.com/gregjones/httpcache"
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
        zerr := zenity.Error(fmt.Sprintf("Repository caching failed: %s", err), zenity.ErrorIcon)
        if zerr != nil {
            return zerr
        }
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
        wf.Rerun(2)
        if !wf.IsRunning("cache") {
            cmd := exec.Command(os.Args[0], "-cache")
            if err := wf.RunInBackground("cache", cmd); err != nil {
                wf.FatalError(err)
            }
        } else {
            log.Printf("cache job already running.")
        }

        if len(repos) == 0 {
            wf.NewItem("Refreshing repository cache…").
                Icon(aw.IconInfo)
            wf.SendFeedback()
            return
        }
    }

    for _, repo := range repos {
        updatedAt := repo.UpdatedAt.Time.Format("02-01-2006 15:04")
        wf.NewItem(*repo.FullName).
            UID(*repo.FullName).
            Subtitle(fmt.Sprintf("%s  •  Updated: %s", *repo.HTMLURL, updatedAt)).
            Var("item_url", *repo.HTMLURL).
            Arg("repo").
            Valid(true)
    }
}

func runSearch(client *github.Client) {
    q := opts.Query
    if !strings.Contains(q, "in:name") {
        q = opts.Query + " in:name"
    }

    repos, _, err := client.Search.Repositories(context.Background(), q, nil)
    if err != nil {
        wf.FatalError(err)
    }

    for _, repo := range repos.Repositories {
        updatedAt := repo.UpdatedAt.Time.Format("02-01-2006 15:04")
        wf.NewItem(*repo.FullName).
            UID(*repo.FullName).
            Subtitle(fmt.Sprintf("%s  •  Updated: %s", *repo.HTMLURL, updatedAt)).
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

    client := github.NewClient(httpcache.NewMemoryCacheTransport().Client()).WithAuthToken(pwd)
    sc, err := testAuthentication(client)

    if err != nil {
        zerr := zenity.Error(
            fmt.Sprintf("Error authenticating: HTTP %d", sc),
            zenity.ErrorIcon,
        )
        if zerr != nil {
            wf.FatalError(err)
        }
        wf.FatalError(err)
    }

    if err := wf.Keychain.Set(keychainAccount, pwd); err != nil {
        wf.FatalError(err)
    }
}
