#!/usr/bin/env bash

SCRIPT_NAME="$(basename "${BASH_SOURCE[0]}")"

if [ -x "$(command -v starship)" ]; then
    eval "$(starship init bash)"
	#info "[$SCRIPT_NAME] Starship PS1 sourced."
	export STARSHIP_CONFIG="$HOME/.config/starship.toml"
  #info "[$SCRIPT_NAME] STARSHIP_CONFIG set to $STARSHIP_CONFIG"
else
	warn "[$SCRIPT_NAME] Starship not found or not executable."
fi
