#!/usr/bin/env bash

# Machine-specific overrides — not versioned, managed directly on each machine
[[ -f "$HOME/.env.local" ]] && source "$HOME/.env.local"

# Clean PATH after all modules and overrides have been loaded
lspath -c PATH
