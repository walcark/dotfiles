#!/usr/bin/env bash

source "$HOME/bin/lib/path.sh"

mkcd() {
    mkdir -p "$1" && cd "$1"
}

fdf() { fd "$@" | fzf; }

# fd + fzf + open — pick a file and open it
# Usage: fdo [fd args]           open with xdg-open (system default)
#        fdo -o <program> [args] open with specific program
fdo() {
    local open_cmd="xdg-open"
    [[ "${1:-}" == "-o" ]] && { open_cmd="$2"; shift 2; }
    local result
    result=$(fd "$@" | fzf) || return 0
    [[ -n "$result" ]] && "$open_cmd" "$result"
}
