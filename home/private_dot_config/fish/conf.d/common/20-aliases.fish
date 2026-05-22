# Navigation
alias ll='ls -lh --color=auto'
alias la='ls -lAh --color=auto'
alias l='ls -CF'
alias c='clear'
alias h='history | tail -50'
alias mkdir='mkdir -p'
alias ..='cd ..'
alias ...='cd ../..'

# Home subdirs
alias docs="cd $HOME/docs"
alias notes="cd $HOME/submodules/zk/notes"
alias sand="cd $HOME/dev/sandbox"
alias curp="cd $HOME/dev/current"
alias conf="cd $HOME/.config"

# Git
alias gs='git status'
alias ga='git add'
alias gb='git branch'
alias gc='git commit'
alias gp='git push'
alias gl='git pull'
alias glog='git lg1'
alias gco='git checkout'
alias gcb='git checkout -b'
alias grh='git reset --hard HEAD'

# Editor
alias vi='NVIM_APPNAME="nvim-configs/astro-nvim" nvim'
alias viconf='cd ~/.config/nvim-configs'

# Chezmoi
alias cm='chezmoi'
alias cmcd='chezmoi cd'
alias cme='chezmoi edit'
alias cmad='chezmoi add'
alias cmap='chezmoi apply'
alias cmt='chezmoi execute-template'

# Tools
alias cat='bat --paging=never'
alias kpxc='keepassxc-cli'
alias rm='rm -i'
