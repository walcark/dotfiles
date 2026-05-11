#!/usr/bin/env bash

SCRIPT_NAME="$(basename "${BASH_SOURCE[0]}")"

if ! shopt -oq posix; then
  if [ -f /usr/share/bash-completion/bash_completion ]; then
    . /usr/share/bash-completion/bash_completion
    #info "[$SCRIPT_NAME] Loaded bash completion from /usr/share/bash-completion/bash_completion."
  elif [ -f /etc/bash_completion ]; then
    . /etc/bash_completion
    #info "[$SCRIPT_NAME] Loaded bash completion from /etc/bash_completion."
  else
    warn "[$SCRIPT_NAME] Bash completion file not found in expected locations."
  fi
else
  warn "[$SCRIPT_NAME] POSIX mode is enabled — bash completion skipped."
fi


