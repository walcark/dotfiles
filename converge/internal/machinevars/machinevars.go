// Package machinevars is how Converge changes a layer's on/off state.
//
// The single source of truth for layer flags is the per-machine file
// chezmoi renders at ~/.config/dotfiles/ansible.yml (see that file's own
// header comment and ansible/group_vars/all.yml's). That file is rendered
// from `.layers` in ~/.config/chezmoi/chezmoi.toml's `[data.layers]` table
// — itself produced once, at `chezmoi init` time, by promptBoolOnce calls
// in home/.chezmoi.toml.tmpl.
//
// promptBoolOnce's cache can't be rewritten non-interactively through the
// chezmoi CLI (its `--promptBool` flag populates the plain `promptBool`
// function's answers, not `promptBoolOnce`'s — confirmed the hard way).
// But chezmoi.toml is read once at startup and never re-rendered by
// `apply` (only `init` re-renders it), so editing `[data.layers]` in it
// directly and then running a *targeted* `chezmoi apply` on just the
// files that depend on `.layers` works, and doesn't touch any other
// pending drift on the machine — a blanket `chezmoi apply` would also
// flush unrelated local edits as a surprise side effect of toggling a
// checkbox, which this deliberately avoids.
package machinevars

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"

	"github.com/walcark/dotfiles/converge/internal/repo"
)

// targets are the destination paths that read `.layers` in their
// templates, applied explicitly after an edit — see the package doc for
// why this is targeted rather than a blanket `chezmoi apply`. Note
// .chezmoiignore.tmpl also reads `.layers.dev`/`.layers.hpc`, but it isn't
// itself an apply target — it's a source-tree control file, and any
// `chezmoi apply` run (targeted or not) already re-evaluates it to decide
// what's managed at all, no separate target needed.
var targets = []string{
	".config/dotfiles/ansible.yml", // home/private_dot_config/dotfiles/ansible.yml.tmpl
	".bashrc.d/.init.sh",           // home/dot_bashrc.d/dot_init.sh.tmpl
}

// SetLayer flips one layer's flag in chezmoi.toml's [data.layers] table
// and re-applies the files that depend on it.
func SetLayer(r *repo.Repo, layerID string, enabled bool) error {
	data, err := r.Data()
	if err != nil {
		return fmt.Errorf("machinevars: %w", err)
	}
	chezmoiInfo, _ := data["chezmoi"].(map[string]any)
	configPath, _ := chezmoiInfo["configFile"].(string)
	if configPath == "" {
		return fmt.Errorf("machinevars: could not find chezmoi's configFile")
	}

	var doc map[string]any
	if _, err := toml.DecodeFile(configPath, &doc); err != nil {
		return fmt.Errorf("machinevars: read %s: %w", configPath, err)
	}

	docData, ok := doc["data"].(map[string]any)
	if !ok {
		return fmt.Errorf("machinevars: %s has no [data] table", configPath)
	}
	layers, ok := docData["layers"].(map[string]any)
	if !ok {
		return fmt.Errorf("machinevars: %s has no [data.layers] table", configPath)
	}
	if _, known := layers[layerID]; !known {
		return fmt.Errorf("machinevars: unknown layer %q", layerID)
	}
	layers[layerID] = enabled

	f, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("machinevars: write %s: %w", configPath, err)
	}
	defer f.Close()
	if err := toml.NewEncoder(f).Encode(doc); err != nil {
		return fmt.Errorf("machinevars: encode %s: %w", configPath, err)
	}

	if err := r.Apply(targets...); err != nil {
		return fmt.Errorf("machinevars: %w", err)
	}
	return nil
}
