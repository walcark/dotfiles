# XDG Base Directories
set -gx XDG_CONFIG_HOME ~/.config
set -gx XDG_CACHE_HOME ~/.cache
set -gx XDG_DATA_HOME ~/.local/share
set -gx XDG_STATE_HOME ~/.local/state

# Editor
set -gx EDITOR nvim
set -gx VISUAL nvim
set -gx NVIM_APPNAME "nvim-configs/astro-nvim"

# Tools
set -gx BAT_THEME "OneHalfDark"
set -gx RIPGREP_CONFIG_PATH ~/.config/ripgrep/config
set -gx STARSHIP_CONFIG ~/.config/starship.toml

# History
set -gx fish_history_max_lines 10000
