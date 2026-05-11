#!/usr/bin/env bash
set -e

if gh auth status &>/dev/null; then
    echo "[chezmoi] gh already authenticated, skipping."
else
    echo "[chezmoi] GitHub authentication required."
    gh auth login
fi

gh auth setup-git
echo "[chezmoi] gh configured as git credential helper."
