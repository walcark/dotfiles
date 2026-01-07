#!/usr/bin/env bash
set -e

echo "[chezmoi] Installing Pixi global tools..."

# Universal tools
pixi global install \
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
    stylua

echo "[chezmoi] CLI installed via Pixi"

