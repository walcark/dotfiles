#!/usr/bin/env bash
set -e

echo "[chezmoi] Installing Pixi global tools..."

# Universal tools
pixi global install \
    git-delta \
    starship \
    ripgrep \
    fd-find \
    fzf \
    jq \
    bat \
    tree \
    gh \
    nvim \
    ruff \
    typst \
    typst-lsp \
    plantuml \
    stylua \
    lazygit \
    restic \
    zoxide \
    tldr \
    tmux \
    python=3.11

echo "[chezmoi] CLI installed via Pixi"

