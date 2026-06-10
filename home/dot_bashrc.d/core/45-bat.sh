#!/usr/bin/env bash

command -v bat >/dev/null || return 0

export BAT_THEME="OneHalfDark"

alias bat='bat --style=numbers,changes,header'
