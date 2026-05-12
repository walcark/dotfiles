#!/usr/bin/env bash

# Machine-specific overrides — not versioned, managed directly on each machine
[[ -f "$HOME/.env.local" ]] && source "$HOME/.env.local"
