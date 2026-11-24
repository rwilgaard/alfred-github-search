# GitHub Search

A workflow for quickly finding repositories in GitHub.

## Requirements

- [Alfred](https://www.alfredapp.com/) 6 with Powerpack
- A GitHub personal access token for the `repo` keyword

## Installation

- [Download the latest release](https://github.com/rwilgaard/alfred-github-search/releases)
- Open the downloaded file in Finder.
- If running on macOS Catalina or later, you _**MUST**_ add Alfred to the list of security exceptions for running unsigned software. See [this guide](https://github.com/deanishe/awgo/wiki/Catalina) for instructions on how to do this.

## Authentication

The `repo` keyword requires a GitHub personal access token. The first time you use it, press ⏎ on "You're not logged in" to open the login window, paste a token and press ⏎ or **Save**. The token is checked against GitHub before it's stored in the macOS Keychain. **Create a token…** opens GitHub's token page with the `repo` scope selected.

To remove a stored token, type `workflow:clearauth` into Alfred.

Global search (`gh`) works without a token, but is subject to stricter API rate limits when unauthenticated.

## Keywords

- `gh` searches repositories globally in GitHub. Responses are cached for 10 minutes, so repeating a search makes no request.
- `repo` searches repositories for the authenticated user. The list is cached and refreshed in the background; repositories you open often rank higher.

## Actions

On a highlighted repository:

| Action | Result |
|--------|--------|
| ⏎ | Open the repository in your browser |
| ⌘⏎ | Open the repository window |

### Repository window

Shows the repository's description, visibility, language, stars, forks, last push and topics, with a list of open issues or pull requests below.

| Key | Result |
|-----|--------|
| type | Filter the list |
| ↑ ↓ | Select an issue or pull request |
| ⏎ / double-click | Open the selected issue or pull request |
| ⌘Y | Preview the selected issue or pull request: title, state, branches, labels, assignees and description. ⏎ opens it on GitHub, Esc closes the preview |
| ⌘1 / ⌘2 | Show open issues / open pull requests |
| ⌘O | Open the repository in your browser |
| ⌘C | Copy the HTTPS clone URL |
| ⌥⌘C | Copy the SSH clone URL |
| Esc | Back to the search results |

Lists show up to 50 open issues and pull requests.

## Configuration

Set in the workflow's User Configuration:

| Option | Default | Description |
| --- | --- | --- |
| Keyword for global search | `gh` | Keyword for searching repositories globally. |
| Keyword for user search | `repo` | Keyword for searching repositories for the authenticated user. |
| Cache age | `360` | Max age in minutes for the user repository cache. |
| Check for updates | on | Check GitHub for a new version once a day. |

## Updates

The workflow checks for new releases once a day and shows an "Update available!" row when one is found. Type `workflow:update` to check now.

## Development

```sh
make build          # build arch-specific binaries
make dev            # build, install into Alfred in place, and reload
make test           # run tests
make package-alfred # build + zip into .alfredworkflow
make zip-alfred     # zip into .alfredworkflow without rebuilding (for CI/CD)
make fmt            # format with gofumpt
make release V=x.y.z        # bump version, package, and tag locally
```

Requires Go 1.27.4+ and `golangci-lint`. Built on [go-alfredutils](https://github.com/rwilgaard/go-alfredutils) v2.
