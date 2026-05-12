if command -q fzf
    fzf --fish | source

    set -gx FZF_DEFAULT_COMMAND 'fd --type f --hidden --follow --exclude .git'
    set -gx FZF_CTRL_T_COMMAND $FZF_DEFAULT_COMMAND
    set -gx FZF_ALT_C_COMMAND 'fd --type d --hidden --follow --exclude .git'
    set -gx FZF_DEFAULT_OPTS "
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
end
