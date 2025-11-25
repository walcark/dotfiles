#!/usr/bin/env bash

SCRIPT_NAME="$(basename "${BASH_SOURCE[0]}")"

if [ -x /usr/bin/lesspipe ]; then
  eval "$(SHELL=/bin/sh lesspipe)"
	# info "[$SCRIPT_NAME] Evaluated lesspipe."
else
	warn "[$SCRIPT_NAME] lesspipe not found or not executable."
fi
