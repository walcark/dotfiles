# Role conventions

Beyond the standard ansible role layout, every role here carries two extra pieces:

```
<role>/
  meta/layer.yml     # machine-readable manifest — see below
  tasks/main.yml     # imports one file per task unit: <id>.yml
  tasks/absent.yml   # imports one file per reversible task unit: absent_<id>.yml
  defaults/main.yml
```

Neither is read by ansible or the playbook today. They exist so a future tool
(a desktop UI, `yq`, whatever) can answer "what does this layer install, task
by task, and how do I take each piece back out" without parsing task YAML.

## `meta/layer.yml`

```yaml
name: Drawing
description: Krita, kept at latest
profiles: [client, server]   # which machine profiles may enable this layer
requires: [core]             # inter-role dependencies — only ever `core`
tasks:
  - id: krita                # matches tasks/krita.yml exactly
    kind: flatpak             # pixi | flatpak | git | command | custom
    description: Krita, kept at latest
    provides: [org.kde.krita] # packages/apps/flatpaks this task owns
    reversible: derived       # derived | explicit | none — see below
```

`requires` should only ever contain `core` — it's the sole role other roles
are allowed to depend on (see the comment in `core/tasks/pixi.yml`).

The role directory name is the same string used as the key in
`group_vars/all.yml`'s `layers:` map (`desktop`, `gnome`, `gaming`, `drawing`,
`dev`). `core` has no `layers.core` flag — it's unconditional in
`playbook.yml` — so its manifest exists for completeness only.

There is no role-level `reversible:` flag: reversibility is a property of
each task, not of the role as a whole, so it lives in `tasks[].reversible`.
A UI (or `yq`) computes something like "5 of 8 tasks reversible" by counting,
rather than trusting a second, separately-maintained role-level boolean —
one source of truth, same principle as the layer flags themselves (see the
main README's note on `~/.config/dotfiles/ansible.yml`).

`tasks[].reversible` is one of:

- **`derived`** — the inverse is mechanical: a generic "undo the same thing"
  pattern (`pixi global uninstall`, `flatpak … state: absent`, delete a
  cloned directory). `tasks/absent_<id>.yml` exists.
- **`explicit`** — the inverse needed hand-written logic beyond the generic
  pattern (e.g. re-resolving a GNOME extension's uuid before it can be
  removed, or deliberately leaving part of what the task installed alone).
  `tasks/absent_<id>.yml` exists.
- **`none`** — no inverse exists, on purpose. No `tasks/absent_<id>.yml`
  file for this id; `tasks/absent.yml` has a comment explaining why.

## `tasks/<id>.yml` ↔ `tasks/absent_<id>.yml`

Every file `main.yml` imports (`import_tasks`, `include_tasks`, or
`include_role`) is one task unit, `id`-named after the file. For every id
with `reversible: derived` or `explicit`, `tasks/absent_<id>.yml` is its
hand-written inverse — same id, `absent_` prefix, nothing else. `absent.yml`
itself mirrors `main.yml`: it imports the same ids, in reverse order.

This 1:1 file naming is strict on purpose, even where it produces a one-line
file (`absent_steam.yml`, `absent_krita.yml`, …): a script can tell whether a
task is reversible just by checking whether the file exists, without parsing
`meta/layer.yml` at all. For `reversible: none`, no file is created — the
file's *absence* is the signal. This happens for tasks that hold
user-authored content ansible must never delete on its own (a private git
clone with local edits or commit history — the `zk` knowledge base) or that
underpin the bootstrap itself (`chezmoi init/apply`).

One exception to "id = filename": a role that installs pixi tools via
`include_role: {name: core, tasks_from: pixi}` has no local `tasks/pixi.yml`
to mirror — it reverses the same way it installs, via
`include_role: {name: core, tasks_from: absent_pixi}` inside its own
`absent.yml`. `core/tasks/absent_pixi.yml` is the one shared inverse, the
same way `core/tasks/pixi.yml` is already the one shared install (see
`dev/meta/layer.yml`'s `pixi` task for the comment explaining this in situ).
