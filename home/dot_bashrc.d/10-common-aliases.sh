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
alias para='cd ~/PARA'
alias proj='cd ~/PARA/projects'
alias arch='cd ~/PARA/archives'
alias ress='cd ~/PARA/ressources'
alias conf='cd ~/.config'
alias loc='cd ~/.local'


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
