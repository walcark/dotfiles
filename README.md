# My Dotfiles

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

---

## Install

```bash
sh -c "$(curl -fsLS get.chezmoi.io)" -- init --apply walcark/dotfiles
```

chezmoi bootstraps itself if not present, then runs the setup scripts in order:

| Script | Role |
|--------|------|
| `run_once_00` | Install pixi |
| `run_once_10` | Install all CLI tools via pixi global |
| `run_once_20` | Authenticate gh, configure git credential helper |
| `run_once_20` | Install KeePassXC |

After apply, a fresh shell will have everything ready.

---

## Global Structure

```
home/
├── .chezmoiscripts/       # one-time setup scripts (run in numbered order)
├── .chezmoiexternal.toml  # external git repos (zk knowledge base)
├── bin/                   # personal scripts → ~/bin/
├── dot_bashrc.tmpl        # ~/.bashrc entry point
├── dot_bashrc.d/          # modular shell config (sourced in order)
│   ├── common/            # loaded on every machine
│   │   ├── 00-xdg.sh
│   │   ├── 05-path.sh.tmpl
│   │   ├── 20-aliases.sh
│   │   ├── 25-functions.sh
│   │   ├── 40-lesspipe.sh
│   │   ├── 41-dircolors.sh
│   │   ├── 42-starship.sh
│   │   ├── 43-pixi.sh
│   │   └── 50-completions.sh
│   ├── client/            # loaded on client machines only
│   │   └── 05-exports.sh
│   └── server/            # loaded on server/HPC machines only
│       ├── 20-aliases.sh
│       └── trex/          # loaded on trex cluster only (hostname trex*)
│           └── 10-env.sh
├── dot_gitconfig
└── private_dot_config/
    ├── astro-nvim/
    ├── starship.toml
    └── wezterm.lua
```

The `dot_bashrc.tmpl` uses the chezmoi machine profile (`client` / `server`) to decide which subdirectory of `dot_bashrc.d/` to source alongside `common/`.

---

## Client / Server Differentiation

Machines are tagged with a profile in `.chezmoi.toml.tmpl` at init time:

```toml
[data]
  profile = "client"   # or "server"
```

This drives two things:

**1. Which bashrc.d subdirectory is sourced**

`common/` is always loaded. Then `client/` or `server/` is loaded on top, allowing profile-specific exports and aliases without touching shared files.

**2. Which files chezmoi deploys (`.chezmoiignore.tmpl`)**

Files or directories can be ignored per profile, so server-only configs never land on a client and vice versa.

To add machine-specific behavior: drop a file in `dot_bashrc.d/client/` or `dot_bashrc.d/server/` — it will be picked up automatically on the next shell start.

**3. Machine-specific local overrides (`~/.env.local`)**

Variables tied to my personal data layout on a given machine (project paths, dataset locations, tool installs) are **not** versioned in dotfiles — they change per machine and per reorganization.

I create `~/.env.local` directly on each machine. It is sourced last by `common/99-local.sh` on every shell start:

```bash
# ~/.env.local — not versioned, managed directly on each machine
export LIBS_PATH=$HOME/libs
export PROJ_PATH=$HOME/dev/projects
export SOS_DATA_ROOT=$HOME/data/sos

# PYTHONPATH for local project sources
for d in "$LIBS_PATH/adjeff/simulations" "$PROJ_PATH/mylab/src"; do
    [ -d "$d" ] && export PYTHONPATH="$d:$PYTHONPATH"
done
```

---

## SSH Keys

SSH keys are **not versioned** — only `~/.ssh/config` is managed by chezmoi. Keys live on each machine and are never shared.

### Philosophy

- **One key pair per machine** — identifies the machine, not the service
- **One key, multiple services** — the same key is added to GitHub, GitLab, etc.
- **Private key never leaves the machine** — agent forwarding is used to authenticate on remote servers

### Setting up a new machine

```bash
# 1. Generate a key pair
sshkey new id_<hostname>
# → comment: walcark@<hostname>
# → set a strong passphrase

# 2. Add the public key to each service
sshkey copy id_<hostname>   # copy to clipboard
# → GitHub: Settings → SSH keys → New SSH key
# → GitLab: Preferences → SSH keys → Add key

# 3. Test
ssh -T git@github.com
```

### Passphrase management with KeePassXC

Store the passphrase in KeePassXC and let it auto-load the key into the agent on database unlock:

1. Create an entry in KeePassXC (e.g. `SSH - <hostname>`)
2. Set the passphrase as the entry password
3. Go to the entry's **SSH Agent** tab → enable, select the private key file
4. KeePassXC **Settings → SSH Agent** → enable integration

From then on: open KeePassXC → key is in the agent → `git push`, `ssh trex`... all work without typing the passphrase.

### Remote servers (agent forwarding)

`ForwardAgent yes` is set for `trex` in `~/.ssh/config`. When connected to trex, the local agent is forwarded — no key file needed on the server, no passphrase to type.

```bash
# Check which keys are loaded and their agent status
sshkey list

# Add a key manually if needed
sshkey add id_<hostname>
```

---

## Scripts & Shell Functions

### Scripts

Standalone executables in `~/bin/` — work anywhere (interactive shell, scripts, cron). Run `lsbin` to list them.

| Script | Description |
|--------|-------------|
| `lsbin` | List all executables in `~/bin` with descriptions, split between custom scripts and installed binaries |
| `tuto` | Manage zettelkasten tutorials in `~/submodules/zk/tutorials` (see Knowledge Base section) |
| `note` | Manage personal notes in `~/submodules/zk/notes` with tag filtering |
| `snip` | Manage code snippets in `~/submodules/zk/snippets`, organized by language |
| `tmpl` | Manage project template skeletons in `~/submodules/zk/templates` |
| `sshkey` | Manage SSH key pairs in `~/.ssh/` — list, generate, show, copy public key |
| `todo` | List TODO / FIXME / HACK / NOTE / XXX comments across source files |
| `extract` | Extract any archive into a folder named after it |
| `pyclean` | Remove Python cache artifacts (`__pycache__`, `*.pyc`, `.pytest_cache`) |
| `tmpclean` | Delete contents of the temporary directory |
| `backup` | Snapshot important home folders to a restic repository via KeePassXC credentials |
| `github_token` | Copy GitHub token from KeePassXC to clipboard |
| `sinter_a100` | Start an interactive SLURM session on an A100 GPU node |
| `sinter_v100` | Start an interactive SLURM session on a V100 GPU node |
| `trex-connect` | Copy server password to clipboard and SSH into trex.cnes.fr |

Scripts self-document via a `# Description: ...` comment on line 2, which `lsbin` parses.

### Shell Functions

Available in interactive shell only (sourced via `bashrc.d/`). Cannot be called from non-interactive scripts.

| Function | Usage | Description |
|----------|-------|-------------|
| `lspath` | `lspath [VAR]` | Pretty-print a colon-separated path variable (default: `PATH`) |
| | `lspath -a <entry> [VAR]` | Add entry after (append, no-op if already present) |
| | `lspath -b <entry> [VAR]` | Add entry before (prepend, deduplicates) |
| | `lspath -r <entry> [VAR]` | Remove all occurrences of entry |
| | `lspath -c [VAR]` | Remove non-existent directories and duplicates |
| `mkcd` | `mkcd <dir>` | Create directory and cd into it |
| `fdf` | `fdf [fd args]` | `fd` + `fzf` — fuzzy pick a file, outputs path to stdout |
| `fdo` | `fdo [fd args]` | `fd` + `fzf` + open — pick a file and open with `xdg-open` |
| | `fdo -o <program> [fd args]` | Open with a specific program instead |
| `z` | `z <query>` | Jump to a frecent directory matching query (zoxide) |
| `zi` | `zi` | Interactive directory jump with fzf (zoxide) |

---

## Binaries

Installed binaries (not custom scripts) live in `~/bin/` too. `lsbin` lists them separately under a **Binaries** section. They are detected by the absence of a shebang line.

All CLI tools are managed by [pixi](https://pixi.sh) global installs — no system package manager needed:

`starship` · `ripgrep` · `fd-find` · `fzf` · `zoxide` · `jq` · `bat` · `tree` · `gh` · `nvim` · `ruff` · `git-delta` · `typst` · `typst-lsp` · `plantuml` · `stylua` · `lazygit` · `restic` · `python 3.11`

---

## Knowledge Base

All knowledge lives in `~/submodules/zk/`, a private git repo synced via `.chezmoiexternal.toml` (pulled daily on `chezmoi apply`). Every write operation auto-commits and pushes.

```
zk/
├── tutorials/    # step-by-step procedures
├── snippets/     # reusable code, organized by language
├── templates/    # full project skeletons
├── notes/        # free-form notes
└── capture/      # inbox
```

### tuto

```bash
tuto new <name>              # create, prompt for tags, open in AstroNvim
tuto list                    # list all with titles and tags
tuto list --tags python,pypi # filter by tags (comma-separated, AND logic)
tuto show <name>             # display (bat if available, else less)
tuto edit <name>             # edit, auto-commit & push on save
tuto del <name>              # delete with confirmation
```

Format: `# Title`, `Date:`, `Tags:` header, then `## Tutorial` and `## Key commands`.

### snip

Snippets are stored as `snippets/<lang>/<name>.<ext>` and organized by language.

```bash
snip new <lang> <name>       # create and open in AstroNvim
snip list                    # list all snippets grouped by language
snip list --lang python      # filter by language
snip show <lang> <name>      # display with syntax highlighting
snip copy <lang> <name>      # copy to clipboard (wl-copy / xclip)
```

### tmpl

Templates are full directory skeletons stored under `templates/<name>/`.

```bash
tmpl new <name>              # save current directory as a template
tmpl list                    # list templates with file count
tmpl use <name>              # apply template into current directory
tmpl use <name> <dest>       # apply template into a new directory
```

Available templates: `python-pixi-lib` (src layout, pixi, ruff, mypy, pytest, CI/CD workflows).

---

## How to Extend

The philosophy: **one concern per file, numbered for load order, profile-scoped for machine differences.**

### Adding a shell alias or export

- Universal → `dot_bashrc.d/common/20-aliases.sh`
- Client-only → `dot_bashrc.d/client/20-aliases.sh` (create if missing)
- Server-only → `dot_bashrc.d/server/20-aliases.sh`

### Adding a shell function

Drop it in `dot_bashrc.d/common/25-functions.sh`.

### Adding a new CLI tool

1. Add it to `run_once_10-pixi-global-install.sh`
2. If it needs init (like starship or pixi), add a `4X-toolname.sh` in `common/`
3. Run `pixi global install <tool>` manually to get it immediately without waiting for a new machine

### Adding a new script

1. Create `home/bin/executable_<name>`
2. Add `# Description: ...` on line 2 — `lsbin` will pick it up automatically
3. Run `chezmoi add ~/bin/<name>` if the script already exists locally

### Adding a new tutorial, snippet or template

```bash
tuto new <name>        # guided: prompts for tags, opens editor, auto-pushes
snip new <lang> <name> # opens editor on a blank file, auto-pushes
tmpl new <name>        # snapshots current directory, auto-pushes
```

### Adding a new machine profile

1. Add a new subdirectory under `dot_bashrc.d/` (e.g. `hpc/`)
2. Update `dot_bashrc.tmpl` to source it when the profile matches
3. Add the profile value to `.chezmoi.toml.tmpl`

---

## Roadmap

- [ ] **Other shells** — Fish (native autocompletion, abbreviations), Zsh
- [ ] **Tmux** — add config alongside WezTerm/Kitty for terminal multiplexing
- [ ] **Neovim integration** — Telescope picker for `snip` and `:TmplUse` command, so snippets and templates are accessible directly from the editor without leaving neovim
- [ ] **Improve SLURM** — richer `sinter` options (memory, CPU count, partition), `server/05-exports.sh` with default SLURM env vars
- [ ] **lazygit & lazydocker configs** — add to `private_dot_config/`
- [ ] **Capture & TODO tool** — CLI tool for quick idea capture and task tracking inspired by org-mode (inbox capture, priorities, deadlines, agenda view), backed by `zk/capture/`
- [ ] **KeePassXC secret management** — centralize all secret retrieval (tokens, passwords, API keys) through `keepassxc-cli` to avoid any plaintext secrets in dotfiles or environment variables

---

## Chezmoi Cheatsheet

```bash
cmap             # apply source → home
cme <file>       # edit a tracked dotfile
cmad <file>      # start tracking a new file
cmcd             # cd into the chezmoi source directory
cm diff          # preview pending changes
cmt              # execute a template inline
```
