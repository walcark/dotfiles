#!/usr/bin/env bash
set -Eeuo pipefail

if [[ -n "${DOTFILES_DEBUG:-}" ]]; then
    set -x
fi

# Debian/Ubuntu
function basics_apt() {
    sudo apt-get update -y
    sudo apt-get install -y \
        curl wget unzip tar git build-essential
}

# Fedora / RHEL
function basics_dnf() {
    sudo dnf install -y curl wget unzip tar git gcc gcc-c++ make
}

# Arch Linux
function basics_pacman() {
    sudo pacman -S --needed --noconfirm \
        curl wget unzip tar git base-devel
}

