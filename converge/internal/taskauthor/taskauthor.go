// Package taskauthor is "add a task to a layer" — the composer the design
// specs (kind selector, generated Task/Uninstall YAML that a hand edit
// overrides) and the design's own note that an empty Uninstall textarea
// means the task isn't reversible, not an omission.
package taskauthor

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/walcark/dotfiles/converge/internal/manifest"
)

// Kinds mirrors the composer's segmented control.
var Kinds = []string{"pixi", "flatpak", "git", "command", "custom"}

// FieldLabels returns the two contextual field labels for a kind — "two
// text fields whose labels change per kind" per the design. A blank label
// means that field isn't used for this kind.
func FieldLabels(kind string) (field1, field2 string) {
	switch kind {
	case "pixi":
		return "Package name (with optional =version)", ""
	case "flatpak":
		return "Flatpak app ID (e.g. org.mozilla.firefox)", "Display name"
	case "git":
		return "Repository URL", "Destination (relative to $HOME)"
	case "command":
		return "Command to run", "Description"
	default:
		return "", ""
	}
}

// Generated is what Generate pre-fills — every field is just a starting
// point; once the composer's own textareas are edited by hand, that text
// wins, same as the mockup's own "edited by hand — saved as written" rule
// (enforced by the caller, not this package: Generate only ever computes
// what a *fresh* form should show).
type Generated struct {
	TaskYAML    string
	AbsentYAML  string // empty means not reversible — shown in red by the UI, not filled in as a placeholder
	Description string
	Provides    string // comma-separated, empty if the kind has no package identity of its own
}

// Generate pre-fills a new task's fields from its kind and the two
// contextual fields — the same generic patterns core/tasks/pixi.yml,
// desktop's old kitty.yml, and every flatpak task in this repo already
// follow, so a generated task fits the existing convention instead of
// inventing a new one.
func Generate(kind, field1, field2 string) Generated {
	switch kind {
	case "pixi":
		pkg := field1
		name, _, _ := strings.Cut(pkg, "=")
		return Generated{
			TaskYAML: fmt.Sprintf(`- name: Install %s via pixi
  ansible.builtin.include_role:
    name: core
    tasks_from: pixi
  vars:
    pixi_tools: [%q]
`, pkg, pkg),
			AbsentYAML: fmt.Sprintf(`- name: Uninstall %s via pixi
  ansible.builtin.include_role:
    name: core
    tasks_from: absent_pixi
  vars:
    pixi_tools: [%q]
`, pkg, pkg),
			Description: pkg + ", installed via pixi",
			Provides:    name,
		}

	case "flatpak":
		label := field2
		if label == "" {
			label = field1
		}
		return Generated{
			TaskYAML: fmt.Sprintf(`- name: Install latest version of %s
  community.general.flatpak:
    name: %s
    state: latest
    method: user
`, label, field1),
			AbsentYAML: fmt.Sprintf(`- name: Uninstall %s
  community.general.flatpak:
    name: %s
    state: absent
    method: user
`, label, field1),
			Description: label + ", kept at latest",
			Provides:    field1,
		}

	case "git":
		return Generated{
			TaskYAML: fmt.Sprintf(`- name: Clone %s
  ansible.builtin.git:
    repo: %q
    dest: "{{ ansible_facts['user_dir'] }}/%s"
    update: false
  become: false
`, field1, field1, field2),
			AbsentYAML:  "", // not reversible by default: a clone may hold uncommitted user edits — same reasoning as core/tasks/zk.yml
			Description: "clone of " + field1,
		}

	case "command":
		desc := field2
		if desc == "" {
			desc = "Run command"
		}
		return Generated{
			TaskYAML: fmt.Sprintf(`- name: %s
  ansible.builtin.command:
    cmd: %q
`, desc, field1),
			AbsentYAML:  "", // a generic command isn't reversible without knowing what it did
			Description: desc,
		}

	default: // custom — "write my own"
		return Generated{
			TaskYAML:   "- name:\n  # write your own task\n",
			AbsentYAML: "",
		}
	}
}

// Input is a submitted composer, after the user may have hand-edited
// anything Generate pre-filled.
type Input struct {
	LayerID     string
	TaskID      string
	Kind        string
	Description string
	Provides    []string
	TaskYAML    string
	AbsentYAML  string // empty = not reversible, matching the design's own rule
}

var idRe = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// Check validates before anything is written: a well-formed id, no
// collision with a task already on the layer.
func Check(layers []manifest.Layer, in Input) error {
	if !idRe.MatchString(in.TaskID) {
		return fmt.Errorf("taskauthor: task id must be lowercase letters, digits and underscores, starting with a letter")
	}
	for _, l := range layers {
		if l.ID != in.LayerID {
			continue
		}
		for _, t := range l.Tasks {
			if t.ID == in.TaskID {
				return fmt.Errorf("taskauthor: task id %q already exists on %s", in.TaskID, in.LayerID)
			}
		}
		return nil
	}
	return fmt.Errorf("taskauthor: unknown layer %q", in.LayerID)
}

// Apply writes tasks/<id>.yml (and tasks/absent_<id>.yml, if reversible),
// appends the corresponding import lines to tasks/main.yml and
// tasks/absent.yml, and adds the task to meta/layer.yml — the same
// structure Phase 0 established by hand, produced here from a form
// instead. Call only after Check has passed, on a branch (see
// internal/authoring).
func Apply(repoRoot string, layers []manifest.Layer, in Input) error {
	var layer manifest.Layer
	for _, l := range layers {
		if l.ID == in.LayerID {
			layer = l
		}
	}
	dir := filepath.Join(repoRoot, layer.RolePath, "tasks")

	taskPath := filepath.Join(dir, in.TaskID+".yml")
	if err := os.WriteFile(taskPath, []byte(ensureTrailingNewline(in.TaskYAML)), 0o644); err != nil {
		return fmt.Errorf("taskauthor: write %s: %w", taskPath, err)
	}
	if err := appendLine(filepath.Join(dir, "main.yml"), fmt.Sprintf("- import_tasks: %s.yml\n", in.TaskID)); err != nil {
		return err
	}

	reversible := "explicit"
	if in.AbsentYAML == "" {
		reversible = "none"
	} else {
		absentPath := filepath.Join(dir, "absent_"+in.TaskID+".yml")
		if err := os.WriteFile(absentPath, []byte(ensureTrailingNewline(in.AbsentYAML)), 0o644); err != nil {
			return fmt.Errorf("taskauthor: write %s: %w", absentPath, err)
		}
		absentImport := fmt.Sprintf("- import_tasks: absent_%s.yml\n  when: \"'%s' not in (absent_skip | default([]))\"\n", in.TaskID, in.TaskID)
		absentMainPath := filepath.Join(dir, "absent.yml")
		if _, err := os.Stat(absentMainPath); os.IsNotExist(err) {
			if err := os.WriteFile(absentMainPath, []byte("# Inverse of "+layer.ID+"/tasks/main.yml.\n"+absentImport), 0o644); err != nil {
				return fmt.Errorf("taskauthor: write %s: %w", absentMainPath, err)
			}
		} else if err := appendLine(absentMainPath, absentImport); err != nil {
			return err
		}
	}

	layer.Tasks = append(layer.Tasks, manifest.Task{
		ID: in.TaskID, Kind: in.Kind, Description: in.Description,
		Provides: in.Provides, Reversible: reversible,
	})
	return manifest.Write(filepath.Join(repoRoot, layer.RolePath, "meta", "layer.yml"), layer)
}

func appendLine(path, line string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("taskauthor: read %s: %w", path, err)
	}
	merged := strings.TrimRight(string(data), "\n") + "\n" + line
	if err := os.WriteFile(path, []byte(merged), 0o644); err != nil {
		return fmt.Errorf("taskauthor: write %s: %w", path, err)
	}
	return nil
}

func ensureTrailingNewline(s string) string {
	if s == "" || strings.HasSuffix(s, "\n") {
		return s
	}
	return s + "\n"
}
