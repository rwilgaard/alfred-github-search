package cmd

import (
    aw "github.com/deanishe/awgo"
    "github.com/deanishe/awgo/update"
    "github.com/rwilgaard/go-alfredutils/alfredutils"
    "github.com/spf13/cobra"
    "go.deanishe.net/fuzzy"
)

type workflowConfig struct {
    CacheAge int `env:"cache_age"`
}

const (
    repo            = "rwilgaard/alfred-github-search"
    repoCacheName   = "repositories.json"
    keychainAccount = "alfred-github-search"
)

var (
    wf      *aw.Workflow
    cfg     = &workflowConfig{}
    rootCmd = &cobra.Command{
        Use:   "github",
        Short: "github is a CLI to be used by Alfred for searching Github repositories",
    }
)

func Execute() {
    wf.Run(run)
}

func run() {
    alfredutils.AddClearAuthMagic(wf, keychainAccount)

    if err := alfredutils.InitWorkflow(wf, cfg); err != nil {
        wf.FatalError(err)
    }

    if err := alfredutils.CheckForUpdates(wf); err != nil {
        wf.FatalError(err)
    }

    if err := rootCmd.Execute(); err != nil {
        wf.FatalError(err)
    }
}

func init() {
    sopts := []fuzzy.Option{
        fuzzy.AdjacencyBonus(10.0),
        fuzzy.LeadingLetterPenalty(-0.1),
        fuzzy.MaxLeadingLetterPenalty(-3.0),
        fuzzy.UnmatchedLetterPenalty(-0.5),
    }
    wf = aw.New(
        aw.SortOptions(sopts...),
        update.GitHub(repo),
    )
}
