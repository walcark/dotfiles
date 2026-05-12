#!/usr/bin/env bash

command -v bat >/dev/null || return 0

export BAT_THEME="OneHalfDark"

alias cat='bat --paging=never'
alias bat='bat --style=numbers,changes,header'
