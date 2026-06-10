#!/usr/bin/env bash
# Clipboard utilities — copy stdin to system clipboard

clipboard_copy() {
    local content
    content=$(cat)
    if command -v wl-copy &>/dev/null; then
        printf '%s' "$content" | wl-copy
    elif command -v xclip &>/dev/null; then
        printf '%s' "$content" | xclip -selection clipboard
    elif command -v pbcopy &>/dev/null; then
        printf '%s' "$content" | pbcopy
    else
        echo "No clipboard command found (wl-copy, xclip, pbcopy)" >&2
        return 1
    fi
}
