#!/usr/bin/env bash
# ZK shared utilities — sourced by tuto, note, snip, tmpl

source "$HOME/bin/lib/shell.sh"

ZK_DIR="${ZK_DIR:-$HOME/submodules/zk}"

zk_push() {
    local file="$1" msg="$2"
    git -C "$ZK_DIR" add "$file"
    git -C "$ZK_DIR" commit -m "$msg"
    git -C "$ZK_DIR" push
}

zk_show() {
    local file="$1" lang="${2:-markdown}"
    if [[ "$lang" == "markdown" ]] && command -v glow &>/dev/null; then
        glow "$file"
    elif command -v bat &>/dev/null; then
        bat --language="$lang" --style=plain "$file"
    else
        less "$file"
    fi
}

zk_clipboard() {
    local file="$1"
    if command -v wl-copy &>/dev/null; then
        wl-copy < "$file"
    elif command -v xclip &>/dev/null; then
        xclip -selection clipboard < "$file"
    elif command -v pbcopy &>/dev/null; then
        pbcopy < "$file"
    else
        error "No clipboard command found (wl-copy, xclip, pbcopy)"
        return 1
    fi
}

zk_editor() {
    NVIM_APPNAME="astro-nvim" ${VISUAL:-nvim} "$1"
}

zk_normalize_name() {
    echo "$1" | tr '[:upper:]' '[:lower:]' | tr ' ' '-'
}

zk_capitalize_title() {
    echo "$1" | tr '-' ' ' | awk '{for(i=1;i<=NF;i++) $i=toupper(substr($i,1,1)) substr($i,2); print}'
}

# Returns 0 if file matches ALL given tags, 1 otherwise
# Usage: zk_match_tags <file> <comma-separated-tags>
zk_match_tags() {
    local file="$1" filter_tags="$2"
    local tags normalized_tags
    tags=$(grep -m1 '^Tags:' "$file" 2>/dev/null | sed 's/^Tags: *//' || true)
    normalized_tags=$(echo "$tags" | tr -d ' ')
    IFS=',' read -ra ftags <<< "$filter_tags"
    for ft in "${ftags[@]}"; do
        [[ ",$normalized_tags," == *",${ft},"* ]] || return 1
    done
    return 0
}

# Print all unique tags found in *.md files of a directory
# Usage: zk_list_tags <dir>
zk_list_tags() {
    local dir="$1"
    find "$dir" -maxdepth 1 -name '*.md' -exec grep -h '^Tags:' {} \; \
        | sed 's/^Tags: *//' \
        | tr ',' '\n' \
        | tr -d ' ' \
        | sort -u \
        | grep -v '^$'
}
