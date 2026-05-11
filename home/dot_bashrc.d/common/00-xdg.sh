#!/usr/bin/env bash

# XDG Base Directory Specification
export XDG_CONFIG_HOME="${XDG_CONFIG_HOME:-$HOME/.config}"
export XDG_CACHE_HOME="${XDG_CACHE_HOME:-$HOME/.cache}"
export XDG_DATA_HOME="${XDG_DATA_HOME:-$HOME/.local/share}"
export XDG_STATE_HOME="${XDG_STATE_HOME:-$HOME/.local/state}"
export XDG_RUNTIME_DIR="${XDG_RUNTIME_DIR:-/run/user/$(id -u)}"

# For tools that still rely on legacy locations but support XDG if set
export LESSHISTFILE="$XDG_CACHE_HOME/less/history"
export ZDOTDIR="$XDG_CONFIG_HOME/zsh"
export HISTFILE="$XDG_STATE_HOME/shell/history"
export PYTHONHISTFILE="$XDG_STATE_HOME/python/history"

# Create dirs if missing
mkdir -p \
    "$XDG_CONFIG_HOME" \
    "$XDG_CACHE_HOME" \
    "$XDG_DATA_HOME" \
    "$XDG_STATE_HOME" \
    "$XDG_CACHE_HOME/less" \
    "$(dirname "$HISTFILE")" \
    "$(dirname "$PYTHONHISTFILE")"

SCRIPT_NAME="$(basename "${BASH_SOURCE[0]}")"
# info "[$SCRIPT_NAME] Loaded XDG base directories"
