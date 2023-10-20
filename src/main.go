package main

import (
	"log"
	"os"
	"os/exec"

	aw "github.com/deanishe/awgo"
	"github.com/deanishe/awgo/update"
	"github.com/google/go-github/v56/github"
	"github.com/gregjones/httpcache"
	"go.deanishe.net/fuzzy"
)

type workflowConfig struct {
    Username string `env:"username"`
    APIToken string
    CacheAge int `env:"cache_age"`
}

const (
    repo            = "rwilgaard/alfred-github-search"
    updateJobName   = "checkForUpdates"
    repoCacheName   = "repositories.json"
    keychainAccount = "alfred-github-search"
)

var (
    cfg *workflowConfig
    wf  *aw.Workflow
)

func init() {
    sopts := []fuzzy.Option{
        fuzzy.AdjacencyBonus(10.0),
        fuzzy.LeadingLetterPenalty(-0.1),
        fuzzy.MaxLeadingLetterPenalty(-3.0),
        fuzzy.UnmatchedLetterPenalty(-0.5),
    }
    wf = aw.New(
        aw.SortOptions(sopts...),
        aw.AddMagic(magicAuth{wf}),
        update.GitHub(repo),
    )
}

func run() {
    if err := cli.Parse(wf.Args()); err != nil {
        wf.FatalError(err)
    }
    opts.Query = cli.Arg(0)

    if opts.Update {
        wf.Configure(aw.TextErrors(true))
        log.Println("Checking for updates...")
        if err := wf.CheckForUpdate(); err != nil {
            wf.FatalError(err)
        }
        return
    }

    if wf.UpdateCheckDue() && !wf.IsRunning(updateJobName) {
        log.Println("Running update check in background...")
        cmd := exec.Command(os.Args[0], "-update")
        if err := wf.RunInBackground(updateJobName, cmd); err != nil {
            log.Printf("Error starting update check: %s", err)
        }
    }

    if wf.UpdateAvailable() {
        wf.Configure(aw.SuppressUIDs(true))
        wf.NewItem("Update Available!").
            Subtitle("Press ⏎ to install").
            Autocomplete("workflow:update").
            Valid(false).
            Icon(aw.IconInfo)
    }

    cfg = &workflowConfig{}
    if err := wf.Config.To(cfg); err != nil {
        wf.FatalError(err)
    }

    if opts.Auth {
        runAuth()
    }

    token, err := wf.Keychain.Get(keychainAccount)
    if err != nil {
        wf.NewItem("You're not logged in.").
            Subtitle("Press ⏎ to authenticate").
            Icon(aw.IconInfo).
            Arg("auth").
            Valid(true)
        wf.SendFeedback()
        return
    }

    cfg.APIToken = token
    client := github.NewClient(httpcache.NewMemoryCacheTransport().Client()).WithAuthToken(cfg.APIToken)

    if opts.UserRepos {
        runUserRepos(client)
        if len(opts.Query) > 0 {
            wf.Filter(opts.Query)
        }
        wf.SendFeedback()
        return
    }

    runSearch(client)

    if wf.IsEmpty() {
        wf.NewItem("No results found...").
            Subtitle("Try a different query?").
            Icon(aw.IconInfo)
    }
    wf.SendFeedback()
}

func main() {
    wf.Run(run)
}
