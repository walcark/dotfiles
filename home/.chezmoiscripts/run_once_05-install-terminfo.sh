#!/usr/bin/env bash
set -e

# Install xterm-kitty terminfo to ~/.terminfo so pixi-installed fish can use it
echo "[chezmoi] Installing kitty terminfo..."
curl -sL "https://raw.githubusercontent.com/kovidgoyal/kitty/master/terminfo/kitty.terminfo" | tic -x -
echo "[chezmoi] Terminfo installed."
