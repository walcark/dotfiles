#!/usr/bin/env bash

SCRIPT_NAME="$(basename "${BASH_SOURCE[0]}")"

__conda_setup="$('/home/kwalcarius/miniconda3/bin/conda' 'shell.bash' 'hook' 2> /dev/null)"

if [ $? -eq 0 ]; then
    eval "$__conda_setup"
    # info "[$SCRIPT_NAME] Conda shell hook successfully evaluated."
else
    warn "[$SCRIPT_NAME] Conda shell hook failed. Trying fallback methods..."

    if [ -f "/home/kwalcarius/miniconda3/etc/profile.d/conda.sh" ]; then
        . "/home/kwalcarius/miniconda3/etc/profile.d/conda.sh"
        # info "[$SCRIPT_NAME] Loaded conda.sh fallback script."
    else
        export PATH="/home/kwalcarius/miniconda3/bin:$PATH"
        warn "[$SCRIPT_NAME] conda.sh not found. Added bin directory to PATH directly."
    fi
fi

unset __conda_setup

