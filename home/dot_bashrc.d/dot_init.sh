#!/usr/bin/env bash

function info() {
    echo -e "\e[32m* ${*}\e[39m"
}

function warn() {
    echo -e "\e[33m* ${*}\e[39m"
}

function error() {
    echo -e "\e[31m* ${*}\e[39m"
}

function nln() {
    echo ""
}

case "${OSTYPE}" in
  solaris*) OSNAME="SOLARIS" ;;
  darwin*)  OSNAME="MACOSX" ;;
  linux*)   OSNAME="LINUX" ;;
  bsd*)     OSNAME="BSD" ;;
  msys*)    OSNAME="WINDOWS" ;;
  *)        OSNAME="${OSTYPE}" ;;
esac

THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &>/dev/null && pwd )"

for BASHRC_D_FILE in "$THIS_DIR"/*.sh; do
  source "${BASHRC_D_FILE}"
done

