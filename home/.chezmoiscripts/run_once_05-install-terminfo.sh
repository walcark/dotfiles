#!/usr/bin/env bash
set -euo pipefail

echo "[chezmoi] Installing kitty terminfo..."

mkdir -p "$HOME/.terminfo"

curl -fsSL \
  "https://raw.githubusercontent.com/kovidgoyal/kitty/master/terminfo/kitty.terminfo" \
  | tic -x -o "$HOME/.terminfo" /dev/stdin

echo "[chezmoi] Terminfo installed to ~/.terminfo"
