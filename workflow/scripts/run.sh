#!/usr/bin/env bash
# Runs the workflow binary for Alfred UI script actions (all .alfredui
# packages use this one script via ../scripts/run.sh). These get no
# alfred_workflow_bundleid, which the binary needs to find its cache and data.
set -euo pipefail

wfdir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)" # workflow/: info.plist and the binary

: "${alfred_workflow_bundleid:=$(plutil -extract bundleid raw "$wfdir/info.plist")}"
export alfred_workflow_bundleid

cd "$wfdir"
exec "$wfdir/alfred-github-search" "$@"
