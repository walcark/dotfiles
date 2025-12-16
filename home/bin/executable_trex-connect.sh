#!/usr/bin/env bash
set -euo pipefail

keepassxc-cli clip ~/Sync/pswd/mdps_keepass.kdbx "Server/Trex"
sleep 0.2
echo "Password for Trex was copied to clipboard for 15 seconds..."
ssh -X walcark@trex.cnes.fr
sleep 15
xclip -selection clipboard /dev/null
