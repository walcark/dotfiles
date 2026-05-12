#!/usr/bin/env bash

source "$HOME/bin/lib/path.sh"

mkcd() {
    mkdir -p "$1" && cd "$1"
}
