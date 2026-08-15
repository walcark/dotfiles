# Converge

A single Go binary that serves the Converge UI on `127.0.0.1:<random port>`
and opens the default browser — see the design handoff (`ansible/roles/README.md`
and the Phase 0 commit) for the project's overall shape.

## Status: Phase 1 (read-only)

Only Overview and Dotfiles (source tree, binaries & scripts) are wired up.
Everything shown is read live from `chezmoi` (`status`, `managed`, `ignored`,
`data`), the `ansible/roles/*/meta/layer.yml` manifests, and the filesystem —
nothing here writes to the machine or the repo. The remaining sidebar items
(Layers, Source edits, Machines, Run log) render as disabled, tagged with
the phase that implements them, rather than showing fake data.

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
- **The "Converge" button and "chezmoi diff" are visible but disabled** —
  running anything is Phase 2.

## Package layout

| Package | Reads |
| --- | --- |
| `internal/repo` | `chezmoi status` / `managed` / `ignored` / `data`; locates the repo root |
| `internal/manifest` | `ansible/roles/*/meta/layer.yml` (Phase 0 schema) |
| `internal/groupvars` | `ansible/group_vars/all.yml` — today's actual layer on/off source of truth |
| `internal/binscan` | `~/bin`, replicating `lsbin`'s own shebang/description logic in Go |
| `internal/pixiglobal` | `<pixiHome>/manifests/pixi-global.toml` — what's actually installed |
| `internal/dfview` | turns the above into the Dotfiles tree's rows |
| `internal/webui` | HTTP handlers, templates, the Nocturne-based static assets |
