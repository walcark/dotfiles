#!/usr/bin/env bash
# Description: KeePassXC helper functions
set -euo pipefail

function assert_keepassxc_cli_exists() {
  if ! command -v keepassxc-cli >/dev/null; then
    echo "keepassxc-cli not found" >&2
    exit 1
  fi
}

assert_keepassxc_cli_exists

DATABASE="$HOME/secrets/keepassxc/dtb.kdbx"

function copy_keepass_password() {
  local ENTRY="$1"
  keepassxc-cli clip -a password "$DATABASE" "$ENTRY"
}

function get_keepass_username() {
  local ENTRY="$1"
  keepassxc-cli show -a username "$DATABASE" "$ENTRY"
}

function get_keepass_password() {
  local ENTRY="$1"
  keepassxc-cli show -a password "$DATABASE" "$ENTRY"
}

function assert_keepassxc_cli_exists() {
  if ! command -v keepassxc-cli >/dev/null; then
    echo "keepassxc-cli not found" >&2
    exit 1
  fi
}
