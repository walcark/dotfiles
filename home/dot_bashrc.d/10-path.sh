#!/usr/bin/env bash

SCRIPT_NAME="$(basename "${BASH_SOURCE[0]}")"

[ -d "$HOME/bin" ] && PATH="$HOME/bin:$PATH"
[ -d "$HOME/.local/bin" ] && PATH="$HOME/.local/bin:$PATH"
[ -d "$HOME/.pixi/bin" ] && PATH="$HOME/.pixi/bin:$PATH" 
#info "[$SCRIPT_NAME] $HOME/bin and $HOME/.local/bin added to PATH."
