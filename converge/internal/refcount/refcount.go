// Package refcount computes what a layer disable should actually remove:
// the design's own formula — "compute the removal set as provides(layer)
// − ⋃ provides(other enabled layers)" — applied at task granularity, since
// that's what ansible/absent.yml's `absent_skip` var operates on.
package refcount

import (
	"sort"

	"github.com/walcark/dotfiles/converge/internal/ledger"
	"github.com/walcark/dotfiles/converge/internal/manifest"
)

// Plan is what to pass to ansible/absent.yml, plus what it means for
// display (the Layers screen's "Orphans" view).
type Plan struct {
	Layers []string          // layer ids to uninstall (absent_layers)
	Skip   []string          // task ids to leave installed (absent_skip)
	Kept   map[string]string // package -> the still-active layer it's kept for
	Remove map[string]string // package -> the now-inactive layer it's being removed from — what an apply should still clear
}

func (p Plan) Empty() bool { return len(p.Layers) == 0 }

// Compute diffs the ledger (what was actually installed as of the last
// successful apply) against the layers active now. A package whose ledger
// layer is no longer active is a removal candidate, UNLESS some other
// currently-active layer's task also provides it — then it's kept, and
// its owning task is added to Skip so ansible/absent.yml leaves it alone.
func Compute(layers []manifest.Layer, led ledger.Ledger, active map[string]bool) Plan {
	isActive := func(id string) bool {
		if id == "core" {
			return true // no layers.core flag — unconditional in playbook.yml
		}
		return active[id]
	}

	// package -> layer id, for every task belonging to a currently active layer.
	activeProviders := map[string]string{}
	for _, l := range layers {
		if !isActive(l.ID) {
			continue
		}
		for _, t := range l.Tasks {
			for _, pkg := range t.Provides {
				activeProviders[pkg] = l.ID
			}
		}
	}

	layersToUninstall := map[string]bool{}
	skipTasks := map[string]bool{}
	kept := map[string]string{}
	remove := map[string]string{}

	for pkg, entry := range led.Packages {
		if isActive(entry.Layer) {
			continue // still declared by its own layer — nothing to do
		}
		layersToUninstall[entry.Layer] = true
		if owner, stillProvided := activeProviders[pkg]; stillProvided {
			skipTasks[entry.Task] = true
			kept[pkg] = owner
		} else {
			remove[pkg] = entry.Layer
		}
	}

	plan := Plan{Kept: kept, Remove: remove}
	for id := range layersToUninstall {
		plan.Layers = append(plan.Layers, id)
	}
	for id := range skipTasks {
		plan.Skip = append(plan.Skip, id)
	}
	sort.Strings(plan.Layers)
	sort.Strings(plan.Skip)
	return plan
}
