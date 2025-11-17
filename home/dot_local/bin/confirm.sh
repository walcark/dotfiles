#!/usr/bin/env bash

confirm() {
    read -r -p "$1 (y/N) " response
    case "$response" in
        [yY][eE][sS]|[yY]) return 0 ;;
        *) return 1 ;;
    esac
}


