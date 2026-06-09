#!/usr/bin/env bash

# Navigation
alias ll='ls -lh --color=auto'
alias la='ls -lAh --color=auto'
alias l='ls -CF'
alias c='clear'
alias h='history | tail -50'
alias mkdir='mkdir -p'
alias rm='rm -i'
alias ..='cd ..'
alias ...='cd ../..'

# Home subdirs
alias lib="cd $HOME/library"
alias tosort="cd $HOME/library/to-sort"
alias docs="cd $HOME/docs"
alias notes="cd $HOME/submodules/zk/notes"
alias sand="cd $HOME/dev/sandbox"
alias curp="cd $HOME/dev/current"
alias fork="cd $HOME/dev/forks"
alias loc='cd ~/.local'
alias conf='cd ~/.config'

# Git
alias gs='git status'
alias ga='git add'
alias gb='git branch'
alias gp='git push'
alias gl='git pull'
alias glog='git lg1'
alias gco='git checkout'
alias gcb='git checkout -b'
alias grh='git reset --hard HEAD'

# Chezmoi
alias cm='chezmoi'
alias cmcd='chezmoi cd'
alias cme='chezmoi edit'
alias cmad='chezmoi add'
alias cmap='chezmoi apply'
alias cmt='chezmoi execute-template'
alias cmtf='chezmoi execute-template -f'
alias viconf='cd ~/.config/nvim-configs'

# Tools
alias kpxc='keepassxc-cli'
