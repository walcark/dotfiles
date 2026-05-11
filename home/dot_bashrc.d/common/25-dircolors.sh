#!/usr/bin/env bash

SCRIPT_NAME="$(basename "${BASH_SOURCE[0]}")"

if [ -x /usr/bin/dircolors ]; then
  if [ -r ~/.dircolors ]; then
    eval "$(dircolors -b ~/.dircolors)"
    # info "[$SCRIPT_NAME] Loaded user ~/.dircolors configuration."
  else
    eval "$(dircolors -b)"
    # info "[$SCRIPT_NAME] Loaded default dircolors configuration."
  fi

  alias ls='ls --color=auto'
  alias grep='grep --color=auto'
  alias fgrep='fgrep --color=auto'
  alias egrep='egrep --color=auto'
  # info "[$SCRIPT_NAME] Colorized aliases set for ls, grep, fgrep, and egrep."
else
  warn "[$SCRIPT_NAME] /usr/bin/dircolors not found or not executable — skipping color setup."
fi

