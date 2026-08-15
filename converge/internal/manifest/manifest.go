// Package manifest reads and writes the ansible/roles/*/meta/layer.yml
// manifests added in Phase 0 of the Converge project (see
// ansible/roles/README.md for the schema). Write exists only for Phase 6's
// authoring flows (internal/rolemerge); everything through Phase 5 only
// ever read these files.
package manifest

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"gopkg.in/yaml.v3"
)

// Task is one entry in a layer's `tasks:` list — the unit that maps to
// ansible/roles/<layer>/tasks/<id>.yml (and, when reversible, tasks/absent_<id>.yml).
type Task struct {
	ID          string   `yaml:"id"`
	Kind        string   `yaml:"kind"`
	Description string   `yaml:"description"`
	Provides    []string `yaml:"provides"`
	Reversible  string   `yaml:"reversible"` // derived | explicit | none
}

// Layer is one role's meta/layer.yml, plus the fields the loader fills in
// from where it found the file rather than from the YAML itself.
type Layer struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Profiles    []string `yaml:"profiles"`
	Requires    []string `yaml:"requires"`
	Tasks       []Task   `yaml:"tasks"`

	ID       string `yaml:"-"` // role directory name, e.g. "desktop" — also the group_vars/all.yml layers key
	RolePath string `yaml:"-"` // path to the role dir, relative to the repo root
}

// PackageCount is the number of distinct packages/apps this layer's tasks
// declare across all of them (sum of len(Task.Provides)).
func (l Layer) PackageCount() int {
	n := 0
	for _, t := range l.Tasks {
		n += len(t.Provides)
	}
	return n
}

// ReversibleCount returns (reversible, total) task counts, matching the
// "N of M tasks reversible" line in the design.
func (l Layer) ReversibleCount() (reversible, total int) {
	total = len(l.Tasks)
	for _, t := range l.Tasks {
		if t.Reversible != "none" {
			reversible++
		}
	}
	return reversible, total
}

// LoadAll reads every ansible/roles/*/meta/layer.yml under repoRoot,
// sorted by role directory name for stable output.
func LoadAll(repoRoot string) ([]Layer, error) {
	pattern := filepath.Join(repoRoot, "ansible", "roles", "*", "meta", "layer.yml")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("manifest: glob %s: %w", pattern, err)
	}

	layers := make([]Layer, 0, len(matches))
	for _, path := range matches {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("manifest: read %s: %w", path, err)
		}
		var l Layer
		if err := yaml.Unmarshal(data, &l); err != nil {
			return nil, fmt.Errorf("manifest: parse %s: %w", path, err)
		}
		roleDir := filepath.Dir(filepath.Dir(path)) // .../ansible/roles/<id>/meta/layer.yml -> .../<id>
		l.ID = filepath.Base(roleDir)
		l.RolePath, err = filepath.Rel(repoRoot, roleDir)
		if err != nil {
			l.RolePath = roleDir
		}
		layers = append(layers, l)
	}

	sort.Slice(layers, func(i, j int) bool { return layers[i].ID < layers[j].ID })
	return layers, nil
}

// Write marshals a Layer back to a meta/layer.yml at path, with the same
// header comment every hand-written one in this repo carries.
func Write(path string, l Layer) error {
	data, err := yaml.Marshal(l)
	if err != nil {
		return fmt.Errorf("manifest: marshal: %w", err)
	}
	header := "# Machine-readable manifest for the Converge UI (see ansible/roles/README.md).\n"
	if err := os.WriteFile(path, append([]byte(header), data...), 0o644); err != nil {
		return fmt.Errorf("manifest: write %s: %w", path, err)
	}
	return nil
}
