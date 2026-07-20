# GitHub Search

A workflow for quickly finding repositories in GitHub.

## Installation

- [Download the latest release](https://github.com/rwilgaard/alfred-github-search/releases)
- Open the downloaded file in Finder.
- If running on macOS Catalina or later, you _**MUST**_ add Alfred to the list of security exceptions for running unsigned software. See [this guide](https://github.com/deanishe/awgo/wiki/Catalina) for instructions on how to do this.

## Authentication

The `repo` keyword requires a GitHub personal access token. The first time you use it, you'll be prompted for one. The token is stored in the macOS Keychain.

To remove a stored token, type `workflow:clearauth` into Alfred.

Global search (`gh`) works without a token, but is subject to stricter API rate limits when unauthenticated.

## Keywords

- `gh` searches repositories globally in GitHub. Results are fetched in the background and cached for 10 minutes.
- `repo` searches repositories for the authenticated user.

## Actions

The following actions can be used on a highlighted repository:
- `⏎` opens the repository in your browser.

## Configuration

Set in the workflow's User Configuration:

| Option | Default | Description |
| --- | --- | --- |
| Keyword for global search | `gh` | Keyword for searching repositories globally. |
| Keyword for user search | `repo` | Keyword for searching repositories for the authenticated user. |
| Cache age | `360` | Max age in minutes for the user repository cache. |

## Updates

The workflow checks for new releases and notifies you when one is available.
