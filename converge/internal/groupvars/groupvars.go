// Package groupvars reads ansible/group_vars/all.yml — today's single
// source of truth for which layers are active (global defaults, same on
// every machine). The design's Phase 3 moves this to a per-machine file
// rendered by chezmoi; until then, this file is what's real.
package groupvars

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Vars is the subset of group_vars/all.yml Converge reads.
type Vars struct {
	Layers map[string]bool `yaml:"layers"`
}

// Load reads ansible/group_vars/all.yml under repoRoot.
func Load(repoRoot string) (Vars, error) {
	path := filepath.Join(repoRoot, "ansible", "group_vars", "all.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		return Vars{}, err
	}
	var v Vars
	if err := yaml.Unmarshal(data, &v); err != nil {
		return Vars{}, err
	}
	return v, nil
}
