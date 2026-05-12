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

_lspath_help() {
    echo "lspath — manage colon-separated path variables"
    echo ""
    echo "Usage: lspath         [VAR]     list entries (default: PATH)"
    echo "       lspath -a <entry> [VAR]  add after (append)"
    echo "       lspath -b <entry> [VAR]  add before (prepend)"
    echo "       lspath -r <entry> [VAR]  remove all occurrences"
    echo "       lspath -c         [VAR]  remove non-existent + duplicates"
    echo "       lspath -h                show this help"
}

# --- Public interface ---

lspath() {
    local opt="${1:-}"

    case "$opt" in
        -a)
            local entry="${2:?usage: lspath -a <entry> [VAR]}" var="${3:-PATH}"
            _path_has "$entry" "$var" && return 0
            export "$var"="${!var:+${!var}:}$entry"
            _path_pretty "$var"
            ;;
        -b)
            local entry="${2:?usage: lspath -b <entry> [VAR]}" var="${3:-PATH}"
            local cleaned
            cleaned=$(_path_entries "$var" | grep -vxF "$entry" | paste -sd ':')
            export "$var"="$entry${cleaned:+:$cleaned}"
            _path_pretty "$var"
            ;;
        -r)
            local entry="${2:?usage: lspath -r <entry> [VAR]}" var="${3:-PATH}"
            local cleaned
            cleaned=$(_path_entries "$var" | grep -vxF "$entry" | paste -sd ':')
            export "$var"="$cleaned"
            _path_pretty "$var"
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
            _path_pretty "$var"
            ;;
        -h)
            _lspath_help
            ;;
        -*)
            error "lspath: unknown option '$opt'" ; return 1
            ;;
        *)
            _path_pretty "${opt:-PATH}"
            ;;
    esac
}
