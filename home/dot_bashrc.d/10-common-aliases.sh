#!/usr/bin/env bash

# Navigation
# ----------
alias ll='ls -alF'
alias la='ls -A'
alias l='ls -CF'
alias ll='ls -lh --color=auto'
alias la='ls -lAh --color=auto'
alias c='clear'
alias h='history | tail -50'
alias mkdir='mkdir -p'
alias rm='rm -i'
alias ..='cd ..'
alias ...='cd ../..'

# Git commands
# ------------
alias gs='git status'
alias ga='git add'
alias gb='git branch'
alias gc='git commit -m'
alias gp='git push'
alias gl='git pull'
alias gco='git checkout'
alias gcb='git checkout -b'
alias grh='git reset --hard HEAD'

# Personnal
# ---------
alias vi='nvim'
alias code='code .'
PARA_DIR="${HOME}/PARA"
alias proj='cd ~/PARA/projects'
alias arch='cd ~/PARA/archives'
alias conf='cd ~/.config'
alias loc='cd ~/.local'
alias vik='NVIM_APPNAME="nvim-kickstart" nvim'
alias via='NVIM_APPNAME="nvim-astro" nvim'

# Ressources aliases
# ------------------
alias ress="cd ${PARA_DIR}/ressources"
alias diary="cd ${PARA_DIR}/ressources/diary"
alias notes="vik ${PARA_DIR}/ressources/notes"
alias umld="cd ${PARA_DIR}/ressources/diagrams"
alias pyf="cd ${PARA_DIR}/ressources/python"
alias zot="cd ${PARA_DIR}/ressources/zotero"

# Chezmoi
# -------
alias cm='chezmoi'
alias cmcd='chezmoi cd'
alias cme='chezmoi edit'
alias cmad='chezmoi add'
alias cmap='chezmoi apply'
alias cmt='chezmoi execute-template'
alias cmtf='chezmoi execute-template -f'

# info "[$SCRIPT_NAME] Common aliases sourced to shell."
