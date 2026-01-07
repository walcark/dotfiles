#!/usr/bin/env bash

set -Eeuo pipefail


function get_os_type() {
  uname
}


function install_chezmoi() {
  echo "Installing chezmoi locally..."
  }


function setup_chezmoi_linux() {
  local bin_dir="$HOME/.local/bin"

  if CHEZMOI_PATH=$(command -v chezmoi); then
    echo "Chezmoi exists on the OS."
    export PATH="$(CHEZMOI_PATH):$PATH"
  else
    echo "Chezmoi does not exist on the OS. Installing..."
    sh -c "$(curl -fsLS get.chezmoi.io)" -- -b "$bin_dir"
    export PATH="$(command -v chezmoi):$PATH"
  fi

  local chezmoi_cmd="${bin_dir}/chezmoi"
  
  # Run chezmoi init
  "${chezmoi_cmd}" init "${DOTFILE_REPO_URL}" \
    --force \
    --branch ""


}


function setup_chezmoi() {
  local ostype 
  ostype=$(get_os_type)

  if [ "$ostype" == "Linux" ]; then
    setup_chezmoi_linux
  else
    echo "Invalid OS type: ${ostype}" >&2
    exit 1
  fi
}


function run_chezmoi() {
  local bin_dir=
}
