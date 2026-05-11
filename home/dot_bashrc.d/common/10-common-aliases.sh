#!/usr/bin/env bash


# Git commands
# -------------------------------------------------------------------
alias gs='git status'
alias ga='git add'
alias gb='git branch'
alias gc='git commit'
alias gp='git push'
alias gl='git pull'
alias glog='git lg1'
alias gco='git checkout'
alias gcb='git checkout -b'
alias grh='git reset --hard HEAD'


# General usage
# -------------------------------------------------------------------
alias vik='NVIM_APPNAME="nvim-kickstart" nvim'
alias vi='NVIM_APPNAME="astro-nvim" nvim'


# Chezmoi
# -------------------------------------------------------------------
alias cm='chezmoi'
alias cmcd='chezmoi cd'
alias cme='chezmoi edit'
alias cmad='chezmoi add'
alias cmap='chezmoi apply'
alias cmt='chezmoi execute-template'
alias cmtf='chezmoi execute-template -f'
alias viconf='chezmoi cd ; cd home/private_dot_config/astro-nvim'
