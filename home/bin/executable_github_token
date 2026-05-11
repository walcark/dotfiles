#!/usr/bin/env bash
set -euo pipefail

DB="$HOME/Sync/pswd/mdps_keepass.kdbx"
ENTRY="Tokens/Github"

if ! command -v keepassxc-cli >/dev/null; then
  echo "keepassxc-cli not found" >&2
  exit 1
fi

if command -v pbcopy >/dev/null; then
  copy_cmd="pbcopy"
elif command -v xclip >/dev/null; then
  copy_cmd="xclip -selection clipboard"
elif command -v wl-copy >/dev/null; then
  copy_cmd="wl-copy"
else
  echo "No clipboard command found" >&2
  exit 1
fi

keepassxc-cli clip "$DB" "$ENTRY" | eval "$copy_cmd"

echo "GitHub token copied to clipboard."
