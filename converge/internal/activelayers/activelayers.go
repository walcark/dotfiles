// Package activelayers resolves which layers are actually active on this
// machine — the single source of truth the design settled on: the
// chezmoi-rendered ~/.config/dotfiles/ansible.yml, falling back to
// ansible/group_vars/all.yml only for a machine bootstrapped before that
// file carried a `layers:` map. See machinevars for how a layer's state
// changes.
package activelayers

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/walcark/dotfiles/converge/internal/groupvars"
)

type rendered struct {
	Layers map[string]bool `yaml:"layers"`
}

// Load returns the active/inactive state of every optional layer (core is
// always active — it carries no `layers.core` flag, see
// ansible/roles/core/meta/layer.yml) and whether the value came from the
// per-machine file or the group_vars fallback.
func Load(repoRoot, homeDir string) (layers map[string]bool, fromMachine bool, err error) {
	path := filepath.Join(homeDir, ".config", "dotfiles", "ansible.yml")
	if data, readErr := os.ReadFile(path); readErr == nil {
		var r rendered
		if yaml.Unmarshal(data, &r) == nil && len(r.Layers) > 0 {
			return r.Layers, true, nil
		}
	}

	gv, err := groupvars.Load(repoRoot)
	if err != nil {
		return nil, false, err
	}
	return gv.Layers, false, nil
}
