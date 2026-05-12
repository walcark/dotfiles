#!/usr/bin/env bash
# Generic shell utilities — sourced by dot_init.sh and bin scripts

function info()  { echo -e "\e[32m* ${*}\e[39m"; }
function warn()  { echo -e "\e[33m* ${*}\e[39m"; }
function error() { echo -e "\e[31m* ${*}\e[39m"; }

case "${OSTYPE}" in
    solaris*) OSNAME="SOLARIS"  ;;
    darwin*)  OSNAME="MACOSX"   ;;
    linux*)   OSNAME="LINUX"    ;;
    bsd*)     OSNAME="BSD"      ;;
    msys*)    OSNAME="WINDOWS"  ;;
    *)        OSNAME="${OSTYPE}" ;;
esac
