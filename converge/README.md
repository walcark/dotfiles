# Converge

A single Go binary that serves the Converge UI on `127.0.0.1:<random port>`
and opens the default browser — see the design handoff (`ansible/roles/README.md`
and the Phase 0 commit) for the project's overall shape.

## Status: Phase 4 (+ Overview, Dotfiles, Runs, Layers)

Overview, Dotfiles, Run log and Layers are wired up. Source edits and
Machines are still disabled placeholders, tagged with the phase that
implements them.

**Layers now has a real per-machine source of truth.** Repo change
alongside the app code: `home/.chezmoi.toml.tmpl` prompts for all 7
layers (`layers.desktop`, `.gnome`, `.gaming`, `.drawing`, `.hpc`, `.dev`,
`.homelab` — same set and defaults as `group_vars/all.yml`), replacing the
old, narrower, disconnected `features.*` prompts (`dev`/`hpc`/`admin`,
the latter unused anywhere) that didn't even feed the ansible run at all.
`home/private_dot_config/dotfiles/ansible.yml.tmpl` now renders a real
`layers:` map from that data — which `include_vars` gives precedence over
`group_vars/all.yml` for every role's `when:` in `ansible/playbook.yml`,
exactly the "app writes the per-machine file" decision from the design
review.

Toggling a layer (`internal/machinevars`) edits `[data.layers]` in
`~/.config/chezmoi/chezmoi.toml` directly and re-applies just the two
files that depend on it — not a blanket `chezmoi apply`, which would also
flush any unrelated pending drift as a surprise side effect of clicking a
checkbox. `chezmoi init --promptBool` looked like the obvious tool for
this and does not work: it only populates the plain `promptBool` template
function's answers, not `promptBoolOnce`'s cache, which is what
`.chezmoi.toml.tmpl` actually calls (confirmed the hard way — the flag
silently no-ops). chezmoi.toml is read once at startup and never
re-rendered by `apply` (only `init` re-renders it), so a direct edit
followed by a targeted apply works without touching that cache at all.

**Ledger** (`internal/ledger`, `~/.local/state/converge/ledger.json`,
not versioned): after every successful run, snapshots every package every
currently-active layer's tasks declare.

## Phase 4 — declarative uninstall, reference counting, orphans

`ansible/absent.yml` is the new entry point: `-e absent_layers=[...]`
picks which roles' `tasks/absent.yml` to run, `-e absent_skip=[...]`
(every role's `absent.yml` now takes this var) leaves specific tasks
installed. Every role's `tasks/absent.yml` import now carries a
`when: "'<id>' not in (absent_skip | default([]))"` guard.

Apply now has a third possible stage after check and apply:
`internal/refcount` diffs the ledger (what's really installed, as of the
last successful run) against the layers active now. Any package whose
ledger layer went inactive is a removal candidate — unless another
currently-active layer's task also provides it, in which case its task id
goes into `absent_skip` and it's left alone. If there's anything left to
remove, the run's third stage invokes `ansible/absent.yml` with the
computed plan, in the same run log, before the ledger updates.

**Orphans** show up right on the Layers screen (not a separate route):
same `refcount.Compute`, called again at render time against the
*current* ledger — so a package stays listed as an orphan for exactly as
long as it takes the next Apply to clear it (or forever, if its task is
`reversible: none`, shown as "stays forever" rather than implying an
Apply will ever fix it).

Found and fixed via real end-to-end VM testing (disable Gaming, Apply,
confirm only Steam/Lutris/Discord are gone — Phase 4's own "done when"):
`ansible/roles/{dev,drawing,gaming,gnome}/meta/main.yml` each declared
`dependencies: [core]` — ansible's own role-dependency mechanism, not the
Phase 0 `meta/layer.yml` schema, and already redundant with
`ansible/playbook.yml` explicitly listing `core` first. Invoking a role
via `include_role` (as `ansible/absent.yml` does) honors that dependency
regardless of `tasks_from`, so uninstalling Gaming was **also** silently
re-running core's full *install* (chezmoi init, pixi installs, the
FlatHub remote, the zk clone) as a side effect. Deleted all four files —
harmless for the normal install flow, actively wrong for this one.

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
- **Found and fixed along the way, testing on the VM**: the runner used to
  locate its ansible callback plugin via `runtime.Caller(0)` relative to
  this source file — works with `go run` from a checkout, breaks
  completely for a binary built here and copied to another machine
  (`runtime.Caller` bakes in the *build* machine's source path). The
  plugin is `//go:embed`ded and written to a temp file at startup now, so
  the binary is actually self-contained.

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
| `internal/activelayers` | resolves the real on/off state of every layer (per-machine file, falling back to group_vars) |
| `internal/machinevars` | how a layer's on/off state actually changes — edits `chezmoi.toml`, re-applies the dependent files |
| `internal/ledger` | reads/writes `~/.local/state/converge/ledger.json` |
| `internal/refcount` | diffs the ledger against active layers — what to actually remove, what to keep because another active layer still needs it |
| `internal/webui` | HTTP handlers, templates, the Nocturne-based static assets |
