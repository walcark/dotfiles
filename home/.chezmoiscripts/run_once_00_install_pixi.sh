#!/usr/bin/env bash
set -e

if command -v pixi >/dev/null 2>&1; then
    echo "[chezmoi] Pixi already installed"
    exit 0
fi

echo "[chezmoi] Installing Pixi..."

# Installation officielle
curl -fsSL https://pixi.sh/install.sh | bash

echo "[chezmoi] Pixi installed"

