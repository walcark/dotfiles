// Package dfview turns chezmoi's flat, machine-oriented output (managed
// paths, ignored paths, status codes) into the rows the Dotfiles screen
// renders. The tree is shown flat (one row per managed path, indented by
// depth) rather than as a collapsible tree — a deliberate Phase 1
// simplification; see converge/README.md.
package dfview

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/walcark/dotfiles/converge/internal/repo"
)

// Row is one line in the Dotfiles source tree.
type Row struct {
	Path   string // full relative target path, e.g. ".bashrc.d/core/20-aliases.sh"
	Name   string // basename
	Depth  int    // number of path separators, for indent
	IsDir  bool
	Status string // applied | modified | pending | ignored
}

// Rows builds the flat, sorted, indented row list for the source tree.
// destDir is chezmoi's destination directory (normally $HOME), used only
// to tell files from directories that already exist on disk.
func Rows(managed, ignored []string, drift []repo.DriftState, destDir string) []Row {
	driftByPath := make(map[string]repo.DriftState, len(drift))
	for _, d := range drift {
		driftByPath[d.Path] = d
	}

	all := make(map[string]bool, len(managed)+len(ignored))
	for _, p := range managed {
		all[p] = false
	}
	for _, p := range ignored {
		all[p] = true
	}

	rows := make([]Row, 0, len(all))
	for p, isIgnored := range all {
		row := Row{
			Path:  p,
			Name:  filepath.Base(p),
			Depth: strings.Count(p, string(filepath.Separator)),
		}
		if info, err := os.Stat(filepath.Join(destDir, p)); err == nil {
			row.IsDir = info.IsDir()
		}
		switch {
		case isIgnored:
			row.Status = "ignored"
		default:
			if d, ok := driftByPath[p]; ok {
				row.Status = d.Label()
			} else {
				row.Status = "applied"
			}
		}
		rows = append(rows, row)
	}

	sort.Slice(rows, func(i, j int) bool { return rows[i].Path < rows[j].Path })
	return rows
}

// Counts tallies rows by status, for the Dotfiles screen's legend.
func Counts(rows []Row) map[string]int {
	c := map[string]int{"applied": 0, "modified": 0, "pending": 0, "ignored": 0}
	for _, r := range rows {
		c[r.Status]++
	}
	return c
}
