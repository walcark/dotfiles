#!/usr/bin/env bash
# Path variable manipulation — sourced by 25-functions.sh

# --- Private helpers ---

_path_entries() {
    local var="${1:-PATH}"
    echo "${!var}" | tr ':' '\n'
}


_path_has() {
    local entry="$1" var="${2:-PATH}"
    [[ ":${!var}:" == *":$entry:"* ]]
}

_path_pretty() {
    local var="${1:-PATH}" value
    value="${!var:-}"

    local GREEN='\e[32m' RED='\e[31m' YELLOW='\e[33m' BOLD='\e[1m' RESET='\e[0m'

    if [[ -z "$value" ]]; then
        echo "$var is not set or empty"
        return 0
    fi

    mapfile -t entries < <(echo "$value" | tr ':' '\n')
    local total=${#entries[@]}
    printf "${BOLD}%s${RESET} (%d %s)\n" "$var" "$total" \
        "$([ "$total" -eq 1 ] && echo entry || echo entries)"

    local seen=() i=1
    for entry in "${entries[@]}"; do
        local dup=0
        for s in "${seen[@]:-}"; do [[ "$s" == "$entry" ]] && { dup=1; break; }; done

        if [[ -z "$entry" ]]; then
            printf "  %2d  ${YELLOW}(empty)${RESET}\n" "$i"
        elif [[ "$dup" -eq 1 ]]; then
            printf "  %2d  ${YELLOW}%s  dup${RESET}\n" "$i" "$entry"
        elif [[ -d "$entry" ]]; then
            printf "  %2d  ${GREEN}%s${RESET}\n" "$i" "$entry"
        else
            printf "  %2d  ${RED}%s  ✗${RESET}\n" "$i" "$entry"
        fi

        seen+=("$entry")
        i=$((i + 1))
    done
}

_pathtool_help() {
    echo "pathtool — manage colon-separated path variables"
    echo ""
    echo "Usage: pathtool           [VAR]     list entries (default: PATH)"
    echo "       pathtool -a <entry> [VAR]  add after (append)"
    echo "       pathtool -b <entry> [VAR]  add before (prepend)"
    echo "       pathtool -r <entry> [VAR]  remove all occurrences"
    echo "       pathtool -c         [VAR]  remove non-existent + duplicates"
    echo "       pathtool -h                show this help"
    echo ""
    echo "Mutations are silent by default (Unix style). Pass -v / --verbose"
    echo "before the action to reprint the result, or just run 'lspath'."
    echo ""
    echo "lspath always lists/prints — it is the display-oriented front end."
}

# --- Public interface ---

pathtool() {
    local verbose=
    if [[ "${1:-}" == "-v" || "${1:-}" == "--verbose" ]]; then
        verbose=1
        shift
    fi

    local opt="${1:-}"

    case "$opt" in
        -a)
            local entry="${2:?usage: pathtool -a <entry> [VAR]}" var="${3:-PATH}"
            _path_has "$entry" "$var" && { [[ -n "$verbose" ]] && _path_pretty "$var"; return 0; }
            export "$var"="${!var:+${!var}:}$entry"
            [[ -z "$verbose" ]] || _path_pretty "$var"
            ;;
        -b)
            local entry="${2:?usage: pathtool -b <entry> [VAR]}" var="${3:-PATH}"
            local cleaned
            cleaned=$(_path_entries "$var" | grep -vxF "$entry" | paste -sd ':')
            export "$var"="$entry${cleaned:+:$cleaned}"
            [[ -z "$verbose" ]] || _path_pretty "$var"
            ;;
        -r)
            local entry="${2:?usage: pathtool -r <entry> [VAR]}" var="${3:-PATH}"
            local cleaned
            cleaned=$(_path_entries "$var" | grep -vxF "$entry" | paste -sd ':')
            export "$var"="$cleaned"
            [[ -z "$verbose" ]] || _path_pretty "$var"
            ;;
        -c)
            local var="${2:-PATH}" seen=()
            while IFS= read -r p; do
                [[ -z "$p" || ! -d "$p" ]] && continue
                local dup=0
                for s in "${seen[@]:-}"; do [[ "$s" == "$p" ]] && { dup=1; break; }; done
                [[ "$dup" -eq 0 ]] && seen+=("$p")
            done < <(_path_entries "$var")
            export "$var"="$(printf '%s\n' "${seen[@]:-}" | paste -sd ':')"
            [[ -z "$verbose" ]] || _path_pretty "$var"
            ;;
        -h)
            _pathtool_help
            ;;
        -*)
            error "pathtool: unknown option '$opt'" ; return 1
            ;;
        *)
            _path_pretty "${opt:-PATH}"
            ;;
    esac
}

# Backward-compatible alias — forwards everything to pathtool.
lspath() {
    pathtool "$@"
}
