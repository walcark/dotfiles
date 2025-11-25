#!/usr/bin/env bash
set -e

# Test pixi installation
if command -v pixi >/dev/null 2>&1; then
    echo "[chezmoi] Pixi already installed"
    exit 0
fi
echo "[chezmoi] Installing Pixi..."

# Set installer location
TMP_SCRIPT="/tmp/install_pixi.sh"
FALLBACK_SCRIPT="$HOME/.local/share/chezmoi/install/common/pixi_install_script.sh"

# Installing pixi
if curl -fsSL https://pixi.sh/install.sh -o "$TMP_SCRIPT"; then
  echo "[chezmoi] Install pixi from curled file."
  bash "$TMP_SCRIPT"
else
  echo "[chezmoi] Failed to download from pixi.sh (blocked or unreachable)."
  echo "[chezmoi] Trying fallback installer."

  if [[ -f "$FALLBACK_SCRIPT" ]]; then
    bash "$FALLBACK_SCRIPT"
  else
    echo "[chezmoi] ERROR: fallback script not found: $FALLBACK_SCRIPT"
    exit 1
  fi 
fi 

echo "[chezmoi] Pixi installation completed."

# Add Pixi to path
export PATH="$HOME/"

# Add pixi to path
export PATH="$HOME/.pixi/bin:$PATH"

echo "[chezmoi] Pixi installed"

