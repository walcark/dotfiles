#!/usr/bin/env bash

command -v fzf >/dev/null || return 0

# Use fd as default command — faster than find, respects .gitignore
export FZF_DEFAULT_COMMAND='fd --type f --hidden --follow --exclude .git'
export FZF_CTRL_T_COMMAND="$FZF_DEFAULT_COMMAND"
export FZF_ALT_C_COMMAND='fd --type d --hidden --follow --exclude .git'

# Default options — Tokyo Night colors, layout
export FZF_DEFAULT_OPTS="
  --height 40%
  --layout reverse
  --border rounded
  --prompt '  '
  --pointer '▶'
  --marker '✓'
  --color 'bg+:#24283b,bg:#1f2335,spinner:#7dcfff,hl:#ff9e64'
  --color 'fg:#c0caf5,header:#ff9e64,info:#7dcfff,pointer:#7dcfff'
  --color 'marker:#7dcfff,fg+:#c0caf5,prompt:#7dcfff,hl+:#ff9e64'
  --bind 'ctrl-u:half-page-up'
  --bind 'ctrl-d:half-page-down'
"

# Ctrl+T — preview file content with bat
export FZF_CTRL_T_OPTS="
  --preview 'bat --style=numbers --color=always --line-range :100 {}'
  --preview-window right:50%:wrap
"

# Alt+C — preview directory tree
export FZF_ALT_C_OPTS="
  --preview 'tree -C {} | head -50'
"

# Ctrl+R — full command visible, most recent first
export FZF_CTRL_R_OPTS="
  --preview 'echo {}'
  --preview-window up:3:wrap
  --bind 'ctrl-y:execute-silent(echo -n {2..} | wl-copy 2>/dev/null || echo -n {2..} | xclip -selection clipboard 2>/dev/null)+abort'
"

# Shell integrations — keybindings (Ctrl+T, Ctrl+R, Alt+C) + completion (**)
eval "$(fzf --bash)"
