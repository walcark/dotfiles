#!/usr/bin/env bash

command -v rg >/dev/null || return 0

export RIPGREP_CONFIG_PATH="$HOME/.config/ripgrep/config"
