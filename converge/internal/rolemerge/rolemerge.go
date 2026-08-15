// Package rolemerge implements "merge role A into role B" — Phase 6's
// flagship authoring operation, and the one the design's "done when"
// names explicitly. Everything here only ever runs against a checkout
// already on a fresh branch (see internal/authoring); nothing in this
// package touches git itself.
package rolemerge

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/walcark/dotfiles/converge/internal/manifest"
)

// Check runs the design's own guardrails, before anything is touched:
//   - core is never a merge target (core/tasks/pixi.yml's own comment
//     documents it as the only permitted cross-role dependency; merging
//     into it would make that ambiguous) — and never a source either:
//     it's unconditional in playbook.yml (no layers.core flag to remove),
//     every other role's pixi task delegates to it by name, and it owns
//     `reversible: none` tasks (chezmoi init, the zk clone) a generic
//     move would mishandle
//   - no task id collides between the two roles (a silent overwrite)
//   - no other role still references `from` via include_role/import_role
//     (deleting it out from under a dependent would break that role)
func Check(repoRoot string, layers []manifest.Layer, from, into string) error {
	if into == "core" {
		return fmt.Errorf("rolemerge: refusing core as a merge target")
	}
	if from == "core" {
		return fmt.Errorf("rolemerge: refusing core as a merge source")
	}

	var fromLayer, intoLayer *manifest.Layer
	for i := range layers {
		switch layers[i].ID {
		case from:
			fromLayer = &layers[i]
		case into:
			intoLayer = &layers[i]
		}
	}
	if fromLayer == nil {
		return fmt.Errorf("rolemerge: unknown layer %q", from)
	}
	if intoLayer == nil {
		return fmt.Errorf("rolemerge: unknown layer %q", into)
	}

	seen := map[string]bool{}
	for _, t := range intoLayer.Tasks {
		seen[t.ID] = true
	}
	for _, t := range fromLayer.Tasks {
		if seen[t.ID] {
			return fmt.Errorf("rolemerge: task id %q exists in both %s and %s", t.ID, from, into)
		}
	}

	referencedBy, err := referencingRoles(repoRoot, layers, from, into)
	if err != nil {
		return fmt.Errorf("rolemerge: %w", err)
	}
	if len(referencedBy) > 0 {
		return fmt.Errorf("rolemerge: %s is still referenced by %s — merge or update those first", from, strings.Join(referencedBy, ", "))
	}

	return nil
}

// referencingRoles greps every OTHER role's tasks/*.yml for
// `name: <from>` (an include_role/import_role call), the way
// core/tasks/pixi.yml and dev/tasks/nvim.yml already do it for core.
func referencingRoles(repoRoot string, layers []manifest.Layer, from, into string) ([]string, error) {
	pattern := regexp.MustCompile(`name:\s*` + regexp.QuoteMeta(from) + `\b`)
	var refs []string
	for _, l := range layers {
		if l.ID == from || l.ID == into {
			continue
		}
		dir := filepath.Join(repoRoot, l.RolePath, "tasks")
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			data, err := os.ReadFile(filepath.Join(dir, e.Name()))
			if err != nil {
				continue
			}
			if pattern.Match(data) {
				refs = append(refs, l.ID)
				break
			}
		}
	}
	return refs, nil
}

// Apply performs the merge: moves from's task files into into's
// directory, appends its main.yml/absent.yml/defaults content, merges
// meta/layer.yml, updates playbook.yml, group_vars/all.yml and
// absent.yml, then deletes from's role directory. Call only after Check
// has passed, on a branch — see internal/authoring.
func Apply(repoRoot string, layers []manifest.Layer, from, into string) error {
	var fromLayer, intoLayer manifest.Layer
	for _, l := range layers {
		if l.ID == from {
			fromLayer = l
		}
		if l.ID == into {
			intoLayer = l
		}
	}

	fromDir := filepath.Join(repoRoot, fromLayer.RolePath)
	intoDir := filepath.Join(repoRoot, intoLayer.RolePath)

	if err := moveTaskFiles(fromLayer, fromDir, intoDir); err != nil {
		return err
	}
	if err := appendTasksFile(fromDir, intoDir, "main.yml"); err != nil {
		return err
	}
	if err := appendTasksFile(fromDir, intoDir, "absent.yml"); err != nil {
		return err
	}
	if err := appendDefaults(fromDir, intoDir); err != nil {
		return err
	}
	if err := writeMergedManifest(repoRoot, intoLayer, fromLayer); err != nil {
		return err
	}
	if err := removePlaybookEntry(repoRoot, from); err != nil {
		return err
	}
	if err := removeGroupVarsLayer(repoRoot, from); err != nil {
		return err
	}
	if err := removeAbsentEntry(repoRoot, from); err != nil {
		return err
	}
	if err := os.RemoveAll(fromDir); err != nil {
		return fmt.Errorf("rolemerge: remove %s: %w", fromDir, err)
	}
	return nil
}

// moveTaskFiles relocates only the files that actually exist locally in
// from/tasks/ — a task delegated entirely to another role (like dev's
// "pixi" task, which has no local tasks/pixi.yml, only an include_role
// call) has nothing to move; its import line is carried over as text by
// appendTasksFile instead, unaffected by the move since it names a role
// other than `from`.
func moveTaskFiles(from manifest.Layer, fromDir, intoDir string) error {
	for _, t := range from.Tasks {
		for _, name := range []string{t.ID + ".yml", "absent_" + t.ID + ".yml"} {
			src := filepath.Join(fromDir, "tasks", name)
			if _, err := os.Stat(src); err != nil {
				continue
			}
			dst := filepath.Join(intoDir, "tasks", name)
			data, err := os.ReadFile(src)
			if err != nil {
				return fmt.Errorf("rolemerge: read %s: %w", src, err)
			}
			if err := os.WriteFile(dst, data, 0o644); err != nil {
				return fmt.Errorf("rolemerge: write %s: %w", dst, err)
			}
		}
	}
	return nil
}

// appendTasksFile appends from/tasks/<name>'s task list onto
// into/tasks/<name>'s — both are YAML lists of tasks (main.yml,
// absent.yml), so concatenating the non-comment content is a valid merge
// of the two lists. Skips entirely if from has no such file (e.g. a role
// with reversible: none tasks might have no absent.yml at all).
func appendTasksFile(fromDir, intoDir, name string) error {
	fromPath := filepath.Join(fromDir, "tasks", name)
	fromData, err := os.ReadFile(fromPath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("rolemerge: read %s: %w", fromPath, err)
	}

	intoPath := filepath.Join(intoDir, "tasks", name)
	intoData, err := os.ReadFile(intoPath)
	if err != nil {
		return fmt.Errorf("rolemerge: read %s: %w", intoPath, err)
	}

	merged := strings.TrimRight(string(intoData), "\n") + "\n\n# — merged from " + filepath.Base(fromDir) + " —\n" + string(fromData)
	if err := os.WriteFile(intoPath, []byte(merged), 0o644); err != nil {
		return fmt.Errorf("rolemerge: write %s: %w", intoPath, err)
	}
	return nil
}

// appendDefaults merges from/defaults/main.yml into into/defaults/main.yml
// (creating it if into didn't have one), so vars like dev_pixi_tools that
// moved task files reference by name keep resolving.
func appendDefaults(fromDir, intoDir string) error {
	fromPath := filepath.Join(fromDir, "defaults", "main.yml")
	fromData, err := os.ReadFile(fromPath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("rolemerge: read %s: %w", fromPath, err)
	}

	intoPath := filepath.Join(intoDir, "defaults", "main.yml")
	intoData, err := os.ReadFile(intoPath)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(intoPath), 0o755); err != nil {
			return err
		}
		return os.WriteFile(intoPath, fromData, 0o644)
	}
	if err != nil {
		return fmt.Errorf("rolemerge: read %s: %w", intoPath, err)
	}

	merged := strings.TrimRight(string(intoData), "\n") + "\n\n" + string(fromData)
	return os.WriteFile(intoPath, []byte(merged), 0o644)
}

func writeMergedManifest(repoRoot string, into, from manifest.Layer) error {
	merged := into
	merged.Tasks = append(append([]manifest.Task{}, into.Tasks...), from.Tasks...)
	return manifest.Write(filepath.Join(repoRoot, into.RolePath, "meta", "layer.yml"), merged)
}

func removePlaybookEntry(repoRoot, from string) error {
	path := filepath.Join(repoRoot, "ansible", "playbook.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("rolemerge: read %s: %w", path, err)
	}
	re := regexp.MustCompile(`\n[ \t]*- role: ` + regexp.QuoteMeta(from) + `\n[ \t]*when: layers\.` + regexp.QuoteMeta(from) + `\n`)
	out := re.ReplaceAll(data, []byte("\n"))
	return os.WriteFile(path, out, 0o644)
}

func removeGroupVarsLayer(repoRoot, from string) error {
	path := filepath.Join(repoRoot, "ansible", "group_vars", "all.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("rolemerge: read %s: %w", path, err)
	}
	re := regexp.MustCompile(`(?m)^[ \t]*` + regexp.QuoteMeta(from) + `:\s*(true|false)\s*\n`)
	out := re.ReplaceAll(data, nil)
	return os.WriteFile(path, out, 0o644)
}

func removeAbsentEntry(repoRoot, from string) error {
	path := filepath.Join(repoRoot, "ansible", "absent.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("rolemerge: read %s: %w", path, err)
	}
	// Matches the "- name: Uninstall <from>\n  ansible.builtin.include_role:\n
	// ... \n  when: ...\n" block as written by every entry in absent.yml.
	re := regexp.MustCompile(`(?ms)^    - name: Uninstall ` + regexp.QuoteMeta(from) + `\n(?:.+\n)*?      when:.*\n\n?`)
	out := re.ReplaceAll(data, nil)
	return os.WriteFile(path, out, 0o644)
}
