#!/usr/bin/env bash

set -Eeuo pipefail

if [ "${DOTFILES_DEBUG:-}" ]; then
    set -x
fi

function install_pixi() {
    if command -v pixi >/dev/null 2>&1; then
        echo "[dotfiles] Pixi already installed."
        return 0
    fi

    echo "[dotfiles] Installing Pixi…"
    curl -fsSL https://pixi.sh/install.sh | sh
    echo "[dotfiles] Pixi installed."
}

function uninstall_pixi() {
    echo "[dotfiles] Uninstalling Pixi…"

    if command -v pixi >/dev/null 2>&1; then
        pixi clean cache || true
        pixi global clean || true
    fi

    rm -rf "$HOME/.pixi"
    rm -f "$HOME/.local/bin/pixi"
    rm -f "$HOME/.bash_completion.d/pixi" 2>/dev/null || true
    rm -f "$HOME/.zsh/completions/_pixi" 2>/dev/null || true

    echo "[dotfiles] Pixi uninstalled."
}

function main() {
    local cmd="${1:-install}"

    case "$cmd" in
        install) install_pixi ;;
        uninstall) uninstall_pixi ;;
        *) echo "Usage: $0 {install|uninstall}" ;;
    esac
}

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "${@}"
fi

