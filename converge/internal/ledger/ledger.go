// Package ledger records what Converge last confirmed was actually
// installed on this machine — ~/.local/state/converge/ledger.json, not
// versioned (it's a fact about this machine, not the repo). Without it, a
// package dropped from a layer's source stays installed and becomes
// invisible; the ledger is what a later orphan-detection screen (Phase 4)
// would diff against.
//
// Written after every successful Apply run, from the layers active at
// that moment — see internal/webui wiring this to runner.Manager.
package ledger

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/walcark/dotfiles/converge/internal/manifest"
)

// Package is one entry the ledger knows about — traced back to the layer
// and task that declared it, so a later "why is this installed" question
// has an answer.
type Package struct {
	Layer string `json:"layer"`
	Task  string `json:"task"`
	Kind  string `json:"kind"`
}

// Ledger is the whole file.
type Ledger struct {
	UpdatedAt time.Time          `json:"updatedAt"`
	Packages  map[string]Package `json:"packages"`
}

func path(homeDir string) string {
	return filepath.Join(homeDir, ".local", "state", "converge", "ledger.json")
}

// Load reads the ledger, or returns an empty one if it doesn't exist yet
// (first run).
func Load(homeDir string) (Ledger, error) {
	data, err := os.ReadFile(path(homeDir))
	if os.IsNotExist(err) {
		return Ledger{Packages: map[string]Package{}}, nil
	}
	if err != nil {
		return Ledger{}, err
	}
	var l Ledger
	if err := json.Unmarshal(data, &l); err != nil {
		return Ledger{}, err
	}
	if l.Packages == nil {
		l.Packages = map[string]Package{}
	}
	return l, nil
}

// Save writes the ledger, creating ~/.local/state/converge if needed.
func (l Ledger) Save(homeDir string) error {
	p := path(homeDir)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o644)
}

// Snapshot builds the ledger that should hold true right after a
// successful apply: every package declared by an active layer's tasks.
// It does not (yet) reference the previous ledger — Phase 4's orphan
// screen is what diffs old vs. new to find packages an apply should have
// removed but couldn't (see ansible/roles/README.md's `reversible: none`
// tasks for why some never will).
func Snapshot(layers []manifest.Layer, active map[string]bool) Ledger {
	l := Ledger{UpdatedAt: time.Now(), Packages: map[string]Package{}}
	for _, layer := range layers {
		if layer.ID != "core" && !active[layer.ID] {
			continue
		}
		for _, task := range layer.Tasks {
			for _, pkg := range task.Provides {
				l.Packages[pkg] = Package{Layer: layer.ID, Task: task.ID, Kind: task.Kind}
			}
		}
	}
	return l
}
