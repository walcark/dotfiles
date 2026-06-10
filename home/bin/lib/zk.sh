#!/usr/bin/env bash
# ZK shared utilities — sourced by note, tuto
# Caller must set ZK_SECTIONS=(...) before calling zk_run

source "$HOME/bin/lib/shell.sh"

ZK_DIR="${ZK_DIR:-$HOME/submodules/zk}"

# --- Internals ---

_zk_normalize() { echo "$1" | tr '[:upper:]' '[:lower:]' | tr ' ' '-'; }

_zk_title() {
    echo "$1" | tr '-' ' ' \
        | awk '{for(i=1;i<=NF;i++) $i=toupper(substr($i,1,1)) substr($i,2); print}'
}

_zk_push() {
    local file="$1" msg="$2"
    git -C "$ZK_DIR" add "$file"
    git -C "$ZK_DIR" commit -m "$msg"
    git -C "$ZK_DIR" push
}

_zk_show_file() {
    if command -v glow &>/dev/null; then
        glow "$1"
    elif command -v bat &>/dev/null; then
        bat --language=markdown --style=plain "$1"
    else
        less "$1"
    fi
}

_zk_editor() { NVIM_APPNAME="astro-nvim" ${VISUAL:-nvim} "$1"; }

_zk_match_tags() {
    local file="$1" filter="$2"
    local tags
    tags=$(grep -m1 '^Tags:' "$file" 2>/dev/null | sed 's/^Tags: *//' | tr -d ' ')
    IFS=',' read -ra wanted <<< "$filter"
    for t in "${wanted[@]}"; do
        [[ ",$tags," == *",${t},"* ]] || return 1
    done
}

# --- Commands ---

_zk_list() {
    local prefix="$1" dir="$2"; shift 2
    local filter_tags=""
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --tags) filter_tags="$2"; shift 2 ;;
            *) error "Unknown option: $1"; exit 1 ;;
        esac
    done

    mapfile -d '' files < <(find "$dir" -maxdepth 1 -name '*.md' -print0 2>/dev/null | sort -z)
    [[ ${#files[@]} -gt 0 ]] || { echo "(no ${prefix}s yet)"; return; }

    local found=0
    for f in "${files[@]}"; do
        [[ -n "$filter_tags" ]] && ! _zk_match_tags "$f" "$filter_tags" && continue
        local name title tags
        name=$(basename "$f" .md)
        title=$(grep -m1 '^# ' "$f" 2>/dev/null | sed 's/^# //' || echo "$name")
        tags=$(grep -m1 '^Tags:' "$f" 2>/dev/null | sed 's/^Tags: *//' || true)
        [[ -n "$tags" ]] \
            && printf "%-30s  %-35s  [%s]\n" "$name" "$title" "$tags" \
            || printf "%-30s  %s\n"           "$name" "$title"
        found=1
    done
    [[ $found -eq 1 ]] || echo "(no ${prefix}s found for given tags)"
}

_zk_show() {
    local prefix="$1" dir="$2" name="${3%.md}"
    local file="$dir/$name.md"
    [[ -f "$file" ]] || { error "$prefix '$name' not found"; exit 1; }
    _zk_show_file "$file"
}

_zk_new() {
    local prefix="$1" dir="$2" name
    name=$(_zk_normalize "$3")
    local file="$dir/$name.md"
    mkdir -p "$dir"
    [[ ! -f "$file" ]] || { error "$prefix '$name' already exists"; exit 1; }
    read -rp "Tags (comma-separated, optional): " tags
    {
        echo "# $(_zk_title "$name")"
        echo ""
        echo "Date: $(date +%Y-%m-%d)"
        echo "Tags: $tags"
        echo ""
        for section in "${ZK_SECTIONS[@]}"; do
            echo "## $section"
            echo ""
        done
    } > "$file"
    _zk_editor "$file"
    _zk_push "$file" "$prefix: add $name"
}

_zk_edit() {
    local prefix="$1" dir="$2" name="${3%.md}"
    local file="$dir/$name.md"
    [[ -f "$file" ]] || { error "$prefix '$name' not found"; exit 1; }
    _zk_editor "$file"
    _zk_push "$file" "$prefix: update $name"
}

_zk_del() {
    local prefix="$1" dir="$2" name="${3%.md}"
    local file="$dir/$name.md"
    [[ -f "$file" ]] || { error "$prefix '$name' not found"; exit 1; }
    read -rp "Delete '$name'? [y/N] " confirm
    [[ "$confirm" =~ ^[Yy]$ ]] || exit 0
    git -C "$ZK_DIR" rm "$file"
    git -C "$ZK_DIR" commit -m "$prefix: delete $name"
    git -C "$ZK_DIR" push
}

_zk_tags() {
    local dir="$1"
    find "$dir" -maxdepth 1 -name '*.md' -exec grep -h '^Tags:' {} \; 2>/dev/null \
        | sed 's/^Tags: *//' | tr ',' '\n' | tr -d ' ' | sort -u | grep -v '^$' \
        || echo "(no tags yet)"
}

_zk_pick() {
    local prefix="$1" dir="$2"

    mapfile -d '' files < <(find "$dir" -maxdepth 1 -name '*.md' -print0 2>/dev/null | sort -z)
    [[ ${#files[@]} -gt 0 ]] || { echo "(no ${prefix}s yet)"; return; }

    local items=()
    for f in "${files[@]}"; do
        local name title tags
        name=$(basename "$f" .md)
        title=$(grep -m1 '^# ' "$f" 2>/dev/null | sed 's/^# //' || echo "$name")
        tags=$(grep -m1 '^Tags:' "$f" 2>/dev/null | sed 's/^Tags: *//' || true)
        [[ -n "$tags" ]] \
            && items+=("$(printf '%-30s  %-35s  [%s]' "$name" "$title" "$tags")") \
            || items+=("$(printf '%-30s  %s'           "$name" "$title")")
    done

    local result key name
    result=$(printf '%s\n' "${items[@]}" | fzf \
        --height=80% --layout=reverse --border \
        --header="$prefix — enter:show  ctrl-e:edit  ctrl-d:delete  ctrl-n:new" \
        --expect=ctrl-e,ctrl-d,ctrl-n \
        --preview="bat --language=markdown --style=plain --color=always $dir/{1}.md 2>/dev/null" \
        --preview-window=right:60%:wrap) || return 0

    IFS=$'\n' read -r key name <<< "$result"
    name=$(awk '{print $1}' <<< "$name")
    [[ -n "$name" ]] || return 0

    case "$key" in
        ctrl-e) _zk_edit "$prefix" "$dir" "$name" ;;
        ctrl-d) _zk_del  "$prefix" "$dir" "$name" ;;
        ctrl-n)
            read -rp "New $prefix name: " new_name
            [[ -n "$new_name" ]] && _zk_new "$prefix" "$dir" "$new_name"
            ;;
        *) _zk_show_file "$dir/$name.md" ;;
    esac
}

# --- Public entry point ---
# Usage: zk_run <prefix> <dir> [cmd] [args...]
# Requires ZK_SECTIONS=(...) set by caller (defines ## sections in `new` template)

_zk_usage() {
    cat <<EOF
$1 — manage ${1}s in $2

Usage: $1                           interactive fzf picker
       $1 list [--tags <t1,t2>]     list
       $1 show <name>               display
       $1 new  <name>               create and open
       $1 edit <name>               edit and commit
       $1 del  <name>               delete
       $1 tags                      list all tags
EOF
    exit 1
}

zk_run() {
    local prefix="$1" dir="$2"; shift 2
    local cmd="${1:-}"; [[ -n "$cmd" ]] && shift

    case "$cmd" in
        "")    _zk_pick "$prefix" "$dir" ;;
        list)  _zk_list "$prefix" "$dir" "$@" ;;
        show)  [[ $# -ge 1 ]] || _zk_usage "$prefix" "$dir"; _zk_show "$prefix" "$dir" "$@" ;;
        new)   [[ $# -ge 1 ]] || _zk_usage "$prefix" "$dir"; _zk_new  "$prefix" "$dir" "$@" ;;
        edit)  [[ $# -ge 1 ]] || _zk_usage "$prefix" "$dir"; _zk_edit "$prefix" "$dir" "$@" ;;
        del)   [[ $# -ge 1 ]] || _zk_usage "$prefix" "$dir"; _zk_del  "$prefix" "$dir" "$@" ;;
        tags)  _zk_tags "$dir" ;;
        *)     _zk_usage "$prefix" "$dir" ;;
    esac
}
