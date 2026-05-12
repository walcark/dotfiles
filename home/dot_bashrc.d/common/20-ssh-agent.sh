#!/usr/bin/env bash

# Start ssh-agent if not already running.
# Keys are loaded automatically on first use via AddKeysToAgent yes in ~/.ssh/config.
# KeePassXC SSH Agent integration handles passphrase on database unlock.

_agent_env="$HOME/.ssh/agent.env"

_agent_load() { [[ -f "$_agent_env" ]] && source "$_agent_env" >/dev/null; }

_agent_start() {
    (umask 077; ssh-agent > "$_agent_env")
    source "$_agent_env" >/dev/null
}

_agent_load

if [[ -z "$SSH_AUTH_SOCK" ]] || ! ssh-add -l >/dev/null 2>&1; then
    _agent_start
fi

unset _agent_env
