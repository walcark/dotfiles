#!/usr/bin/env bash

export NVIM_APPNAME="nvim-configs/mini-nvim"
export VISUAL=nvim
export EDITOR=nvim

# Go defaults GOPATH to $HOME/go with nothing set — XDG-clean like
# everything else in core/00-xdg.sh instead of dropping a bare "go/"
# straight into $HOME.
export GOPATH="$XDG_DATA_HOME/go"
export GOMODCACHE="$GOPATH/pkg/mod"
export GOCACHE="$XDG_CACHE_HOME/go-build"

source "$HOME/.config/nvim-configs/bin/utils"
