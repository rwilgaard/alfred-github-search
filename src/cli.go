package main

import "flag"

var (
    opts = &options{}
    cli  = flag.NewFlagSet("alfred-github-search", flag.ContinueOnError)
)

type options struct {
    // Arguments
    Query string

    // Commands
    Update    bool
    UserRepos bool
    Search    bool
    Auth      bool
    Cache     bool
}

func init() {
    cli.BoolVar(&opts.Update, "update", false, "check for updates")
    cli.BoolVar(&opts.UserRepos, "user-repos", false, "only list repos for the authenticated user")
    cli.BoolVar(&opts.Search, "search", false, "search for repos globally")
    cli.BoolVar(&opts.Auth, "auth", false, "authenticate")
    cli.BoolVar(&opts.Cache, "cache", false, "refresh cache")
}
