#!/usr/bin/env bash
# fzf utilities — sourced by bashrc.d/core/25-functions.sh

fdf() { fd "$@" | fzf; }

fdo() {
    local open_cmd="xdg-open"
    [[ "${1:-}" == "-o" ]] && { open_cmd="$2"; shift 2; }
    local result
    result=$(fd "$@" | fzf) || return 0
    [[ -n "$result" ]] && "$open_cmd" "$result"
}
