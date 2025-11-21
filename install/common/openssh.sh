#!/usr/bin/env bash
set -Eeuo pipefail

if [[ -n "${DOTFILES_DEBUG:-}" ]]; then
    set -x
fi

function openssh_apt() {
    sudo apt-get update -y
    sudo apt-get install -y openssh-client
}

function openssh_dnf() {
    sudo dnf install -y openssh-clients
}

function openssh_pacman() {
    sudo pacman -S --needed --noconfirm openssh
}

