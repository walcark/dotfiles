#!/usr/bin/env bash

set -Eeuo pipefail

if [ "${DOTFILES_DEBUG:-}" ]; then
    set -x
fi

function install_pixi() {
	# https://pixi.sh/latest/installation/
	curl -fsSL https://pixi.sh/install.sh | sh
}

function uninstall_pixi() {
	# https://pixi.sh/latest/installation/#uninstall
	pixi clean cache
	
}

function main() {
	install_pixi
}

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main
fi
