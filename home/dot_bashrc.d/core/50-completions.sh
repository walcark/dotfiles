#!/usr/bin/env bash

# Avoid POSIX shells
if ! shopt -oq posix; then

    # Fedora / modern RHEL
    if [ -f /etc/profile.d/bash_completion.sh ]; then
        . /etc/profile.d/bash_completion.sh

    # Debian / Arch / generic
    elif [ -f /usr/share/bash-completion/bash_completion ]; then
        . /usr/share/bash-completion/bash_completion

    # legacy fallback
    elif [ -f /etc/bash_completion ]; then
        . /etc/bash_completion
    fi
fi
