#!/usr/bin/env bash

SCRIPT_NAME="$(basename "${BASH_SOURCE[0]}")"

if [ -x "$(command -v pixi)" ]; then
    eval "$(pixi completion --shell bash)"
    # info "[$SCRIPT_NAME] Pixi autocompletion enabled."
else
    warn "[$SCRIPT_NAME] Pixi not found or not executable."
fi
