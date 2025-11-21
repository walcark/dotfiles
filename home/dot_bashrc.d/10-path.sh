#!/usr/bin/env bash

SCRIPT_NAME="$(basename "${BASH_SOURCE[0]}")"

# set PATH so it includes user's private bin if it exists
[ -d "$HOME/bin" ] && PATH="$HOME/bin:$PATH"

# set PATH so it includes user's private bin if it exists
[ -d "$HOME/.local/bin" ] && PATH="$HOME/.local/bin:$PATH"

# add all pixi downloaded binaries to path
[ -d "$HOME/.pixi/bin" ] && PATH="$HOME/.pixi/bin:$PATH" 

info "[$SCRIPT_NAME] $HOME/bin and $HOME/.local/bin added to PATH."
