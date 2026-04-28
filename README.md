# dotfiles

<p align="center">
  <a href="https://www.chezmoi.io">
    <img src="https://img.shields.io/badge/managed%20with-chezmoi-blue?logo=chezmoi">
  </a>
  <a href="https://pixi.sh">
    <img src="https://img.shields.io/endpoint?url=https://raw.githubusercontent.com/prefix-dev/pixi/main/assets/badge/v0.json">
  </a>
  <img src="https://img.shields.io/badge/platform-linux-orange?logo=linux&logoColor=white">
  <img src="https://img.shields.io/github/last-commit/walcark/dotfiles">
</p>

Personal dotfiles managed with [chezmoi](https://www.chezmoi.io/). Linux only.

## Quick install

```bash
sh -c "$(curl -fsLS get.chezmoi.io)" -- init --apply <your-repo-url>
```

chezmoi will bootstrap itself if not already present, then apply all configs.

## What's inside

### Shell

- **Bash** with a modular `~/.bashrc.d/` — each file handles one concern (XDG paths, aliases, prompt, completions, tools)
- **Starship** prompt — minimal, fast, pixi-aware
- Common aliases for git (`gs`, `ga`, `glog`, …) and chezmoi (`cm`, `cme`, `cmap`, …)

### Terminal

- **WezTerm** — Tokyo Night Storm theme, JetBrains Mono Nerd Font, vim-style pane splits (`Alt+hjkl` to navigate, `Ctrl+Alt+hjkl` to split)

### Editor

Two Neovim configs, switchable via `$NVIM_APPNAME`:

| Alias | Config | Description |
|-------|--------|-------------|
| `vi`  | `astro-nvim` | AstroNvim — full-featured daily driver |
| `vik` | `nvim-kickstart` | Kickstart — lightweight fallback |

### Package management

[Pixi](https://pixi.sh) handles all CLI tools as global installs — no system package manager needed:

`starship` · `ripgrep` · `fd-find` · `fzf` · `jq` · `bat` · `tree` · `gh` · `nvim` · `ruff` · `typst` · `plantuml` · `stylua` · `lazygit` · `python 3.11`

### Other tools

- **KeePassXC** — installed from AppImage into `~/.local/share/keepassxc/`
- **Kitty** — theme config kept around as a secondary terminal option
- **lazygit** — TUI git client with custom keybindings

## Repo layout

```
.
├── home/                        # chezmoi source root
│   ├── .chezmoiscripts/         # run_once setup scripts
│   │   ├── 00 — install pixi
│   │   ├── 10 — pixi global installs
│   │   └── 20 — install KeePassXC
│   ├── bin/                     # personal scripts (sinter, trex, github token)
│   ├── dot_bashrc.tmpl          # ~/.bashrc
│   ├── dot_bashrc.d/            # modular shell config
│   ├── dot_gitconfig            # git aliases and user config
│   └── private_dot_config/
│       ├── astro-nvim/          # AstroNvim config
│       ├── nvim/                # NvChad / kickstart config
│       ├── kitty/               # Kitty terminal
│       ├── lazygit/             # lazygit config
│       ├── pixi/                # pixi global manifest
│       ├── starship.toml        # Starship prompt
│       └── wezterm.lua          # WezTerm config
├── install/                     # standalone install helpers
│   └── common/                  # pixi backup installer, starship, uv
└── install.sh                   # bootstrap entrypoint
```

## Chezmoi cheatsheet

```bash
cm diff          # preview pending changes
cmap             # apply source → home
cme <file>       # edit a dotfile
cmad <file>      # track a new file
cmcd             # cd into the source directory
cmt              # execute a template inline
```
