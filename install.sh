#!/usr/bin/env bash
# 
# Bootstrap a new machine. Usage:
#   bash <(curl -fsSL https://raw.githubusercontent.com/walcark/dotfiles/main/install.sh)
#
# During the reorganization of the dotfiles, the command becomes the following:
#   DOTFILES_BRANCH=reorganize bash <(curl -fsSL https://raw.githubusercontent.com/walcark/dotfiles/main/install.sh)

set -Eeuo pipefail

# Dotfiles location and branch
DOTFILES_URL="${DOTFILES_URL:-https://github.com/walcark/dotfiles.git}"
DOTFILES_BRANCH="${DOTFILES_BRANCH:-main}"
PULL_DIR="$HOME/.ansible/pull"


log() { printf '\n[Dotfiles Bootstrap] %s\n' "$*"; }

# ansible-pull asks for sudo on stdin, a real terminal is thus required
if [ ! -t 0 ]; then
  exec </dev/tty 2>/dev/null \
      || { log "stdin is not a terminal, please use: bash <(curl -fsSL ...)" >&2; exit 1; }
fi

# 1. Install pixi - user space based - no sudo required
if ! command -v pixi >/dev/null 2>&1; then
  log "Installing pixi..."
  curl -fsSL https://pixi.sh/install.sh | bash
fi
export PATH="$HOME/.pixi/bin:$PATH"

# 2. Install ansible / chezmoi and git via pixi
log "Installing ansible, chezmoi and git via pixi..."
pixi global install ansible chezmoi git

# pixi only auto-exposes binaries matching the package name: expose the
# rest of the ansible suite (ansible-doc is also what gives the ansible
# language server its module completion and docs in neovim)
pixi global expose add \
  ansible-doc ansible-playbook ansible-galaxy ansible-vault \
  ansible-console ansible-config ansible-inventory ansible-pull \
  --environment ansible

# 3. Full convergence — everything else lives in the playbook.
# The -e overrides propagate URL/branch to the chezmoi task in roles/core,
# so a bootstrap from a test branch is consistent end to end.
log "Converge to required state with ansible-pull (branch: ${DOTFILES_BRANCH})..."
ansible-pull \
  --url "$DOTFILES_URL" \
  --checkout "$DOTFILES_BRANCH" \
  --directory "$PULL_DIR" \
  --inventory "$PULL_DIR/ansible/inventory.ini" \
  -e "dotfiles_url=$DOTFILES_URL" \
  -e "dotfiles_branch=$DOTFILES_BRANCH" \
  ansible/playbook.yml

log "Finished."
log "Usage: chezmoi apply for dotfiles - ansible-pull for the environment."
