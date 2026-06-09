#!/usr/bin/env bash

command -v zoxide >/dev/null || return 0

# Init zoxide — replaces cd with z, adds zi (interactive with fzf)
eval "$(zoxide init bash)"
