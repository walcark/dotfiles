# Export installed binaries to path
BIN_PATH="$HOME/.local/share"
export PATH="${BIN_PATH}/pixi:${PATH}"
export PATH="${BIN_PATH}/chezmoi:${PATH}"
export PATH="${BIN_PATH}/keepassxc:${PATH}"

# Export path for configurations files
export STARSHIP_CONFIG="$HOME/.config/starship.toml"

# Add alias for some binaries
alias kpxc="keepassxc-cli"
