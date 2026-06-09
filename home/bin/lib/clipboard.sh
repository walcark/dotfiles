#!/usr/bin/env bash
# Clipboard utilities — copy stdin to system clipboard

clipboard_copy() {
    if command -v wl-copy &>/dev/null; then
        wl-copy
    elif command -v xclip &>/dev/null; then
        xclip -selection clipboard
    elif command -v pbcopy &>/dev/null; then
        pbcopy
    else
        echo "No clipboard command found (wl-copy, xclip, pbcopy)" >&2
        return 1
    fi
}
