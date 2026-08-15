# Converge

A single Go binary that serves the Converge UI on `127.0.0.1:<random port>`
and opens the default browser — see the design handoff (`ansible/roles/README.md`
and the Phase 0 commit) for the project's overall shape.

## Status: Phase 2 (Overview, Dotfiles, Runs)

Overview, Dotfiles (source tree, binaries & scripts) and Run log are wired
up. Layers, Source edits and Machines are still disabled placeholders,
tagged with the phase that implements them.

**⚠️ `CONVERGE_DEV_SOURCE` only sandboxes reads, not runs.** It repoints
the read-only chezmoi queries (status/managed/ignored/data) at a plain
working copy for fast iteration — but Check and Apply on the Run log page
always execute `ansible-playbook` for real, against whatever machine
Converge is running on, dev mode or not. There is no dry-run-only mode
for the runner itself; `Check` (`--check --diff`) is the closest thing,
and even that isn't guaranteed side-effect-free for every module. Run
Apply on a machine you're fine converging — the VM for anything you're
not sure about.

## Running it

```sh
cd converge
go run .
```

It finds the repo by asking `chezmoi source-path`, then walking up to the
first ancestor with an `ansible/` directory — so it only works on a machine
where this repo has actually been applied via chezmoi (i.e. everywhere
`install.sh` has run).

## Known Phase 1 simplifications

- **The source tree is flat, not collapsible.** The design shows expandable
  folders; this renders every managed/ignored path as one indented row
  instead. Same data, less interaction — revisit if it's ever unwieldy at
  this repo's actual size (it isn't, yet).
- **"Source edits" on Overview counts `git status --porcelain` in the repo
  root**, not staged Converge patches (there's nothing to stage before
  Phase 6). It's a real, honest number — just not the one the design
  eventually means by that label.
- **Pixi tools installed by the dynamic Neovim language stack** (loaded at
  run time from the external `nvim-configs` repo, see `dev/tasks/nvim.yml`)
  show up untagged in Binaries & scripts: they're real installs, but no
  static manifest claims them, so there's no layer to tag them with.
- **"chezmoi diff" is still visible but disabled** — that's a dotfiles
  diff, not an ansible run; it fits better alongside the Phase 5 editor.
- **`--diff` output isn't shown.** The run is passed `--diff`, but the
  converge_json callback plugin doesn't implement `v2_on_file_diff`, so
  file-content diffs ansible would normally print are silently dropped —
  only the ok/changed/failed/skipped line survives. Worth adding if the
  run log ever needs to answer "changed to *what*", not just "changed".
- **No sandbox mode yet.** The design's `podman run --rm fedora:42` test
  gate is Phase 6; Check and Apply both run directly against the real
  machine (see the warning above).
- **Found and fixed along the way**: `dev/tasks/golang.yml`'s Go smoke
  test failed under `--check` (the `tempfile` module doesn't support check
  mode, so `go_smoke_dir` was never set and everything after it broke on
  `.path`). Fixed with `check_mode: false` on the block — it never
  persists anything, so there's nothing check mode needs to protect there.

## Package layout

| Package | Reads |
| --- | --- |
| `internal/repo` | `chezmoi status` / `managed` / `ignored` / `data`; locates the repo root |
| `internal/manifest` | `ansible/roles/*/meta/layer.yml` (Phase 0 schema) |
| `internal/groupvars` | `ansible/group_vars/all.yml` — today's actual layer on/off source of truth |
| `internal/binscan` | `~/bin`, replicating `lsbin`'s own shebang/description logic in Go |
| `internal/pixiglobal` | `<pixiHome>/manifests/pixi-global.toml` — what's actually installed |
| `internal/dfview` | turns the above into the Dotfiles tree's rows |
| `internal/runner` | runs `ansible-playbook` with a custom streaming JSON callback (`internal/runner/callback/converge_json.py`), tracks in-memory run state |
| `internal/webui` | HTTP handlers, templates, the Nocturne-based static assets |
