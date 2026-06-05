#!/usr/bin/env bash
# Generic fzf utilities


function fzf_so() {
    local file
    printf '%q\n' fzf "${FZF_SO_OPTS[@]}" "$@" >&2
    file=$(
        fzf \
            --style=full \
            --height=100% \
            --layout=reverse \
            --border \
            --preview-window=right \
            --preview 'bat -n --color=always {}' \
            "$@"
    ) || return

    vi "$file"
}
