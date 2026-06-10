#!/usr/bin/env bash
set -euo pipefail

echo "[chezmoi] Creating home directory layout..."

dirs=(
    "$HOME/docs"
    "$HOME/dev"
    "$HOME/library"
    "$HOME/tmp"
    "$HOME/submodules"
    "$HOME/Sync"
)

for d in "${dirs[@]}"; do
    if [[ ! -d "$d" ]]; then
        mkdir -p "$d"
        echo "  created $d"
    fi
done

echo "[chezmoi] Home layout ready."
