#!/usr/bin/env bash

source "$HOME/bin/lib/path.sh"
source "$HOME/bin/lib/fzf.sh"

mkcd() {
    mkdir -p "$1" && cd "$1"
}
