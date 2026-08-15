// Package repo locates the chezmoi-managed dotfiles tree and the dotfiles
// git repository around it, and wraps the small set of chezmoi subcommands
// Converge reads from — never anything that mutates the destination
// directory. See the design note in ansible/roles/README.md: this package
// is what Phase 1 ("read-only app") is built on.
package repo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// Repo holds everything Converge needs to find: the chezmoi binary, the
// chezmoi source directory (the "home" subtree chezmoi actually manages,
// after .chezmoiroot), and the repo root (the git checkout containing both
// ansible/ and that source directory).
type Repo struct {
	ChezmoiBin string
	SourceDir  string // e.g. ~/.local/share/chezmoi/home
	RootDir    string // e.g. ~/.local/share/chezmoi — contains ansible/, install.sh

	// devSource, when set, is passed as `--source` on every chezmoi call —
	// see Detect's CONVERGE_DEV_SOURCE handling. It points chezmoi straight
	// at a plain working copy (no init, no clone, no commit even required)
	// so day-to-day development doesn't need a push to be visible.
	devSource string
}

// Detect finds the repo Converge should read.
//
// If CONVERGE_DEV_SOURCE is set (to a working copy of this repo, e.g.
// ~/dev/current/dotfiles), it's used directly — chezmoi reads it in place
// via --source, compared against the real destination directory ($HOME),
// with no init/clone/commit/push involved. This is the fast loop for
// developing Converge itself or the ansible/roles manifests.
//
// Otherwise, it asks the system chezmoi where its source directory is (the
// machine's actual applied dotfiles) and walks up from there to find the
// repo root — the first ancestor with an ansible/ directory next to it,
// which is what .chezmoiroot in the source tree makes SourceDir a
// subdirectory of.
func Detect() (*Repo, error) {
	bin, err := exec.LookPath("chezmoi")
	if err != nil {
		return nil, fmt.Errorf("repo: chezmoi not found on PATH: %w", err)
	}

	if dev := os.Getenv("CONVERGE_DEV_SOURCE"); dev != "" {
		root, err := filepath.Abs(dev)
		if err != nil {
			return nil, fmt.Errorf("repo: CONVERGE_DEV_SOURCE %q: %w", dev, err)
		}
		if info, err := os.Stat(filepath.Join(root, "ansible")); err != nil || !info.IsDir() {
			return nil, fmt.Errorf("repo: CONVERGE_DEV_SOURCE %q has no ansible/ directory", root)
		}
		sourceDir := filepath.Join(root, "home") // matches this repo's .chezmoiroot ("home")
		return &Repo{ChezmoiBin: bin, SourceDir: sourceDir, RootDir: root, devSource: sourceDir}, nil
	}

	out, err := exec.Command(bin, "source-path").Output()
	if err != nil {
		return nil, fmt.Errorf("repo: chezmoi source-path: %w", err)
	}
	sourceDir := strings.TrimSpace(string(out))

	root := sourceDir
	for {
		parent := filepath.Dir(root)
		if parent == root {
			return nil, fmt.Errorf("repo: no ansible/ directory found above %s", sourceDir)
		}
		root = parent
		if info, err := os.Stat(filepath.Join(root, "ansible")); err == nil && info.IsDir() {
			break
		}
	}

	return &Repo{ChezmoiBin: bin, SourceDir: sourceDir, RootDir: root}, nil
}

func (r *Repo) run(args ...string) ([]byte, error) {
	if r.devSource != "" {
		args = append([]string{"--source", r.devSource}, args...)
	}
	cmd := exec.Command(r.ChezmoiBin, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("repo: chezmoi %s: %w (%s)", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return stdout.Bytes(), nil
}

// DriftState is what changed on the machine since the last chezmoi apply,
// versus what the next apply would still change. Together they drive the
// tree's status dot and the Overview "Local drift" list.
type DriftState struct {
	Path      string // relative to the destination directory (usually $HOME)
	DriftCode byte   // 'chezmoi status' first column
	ApplyCode byte   // 'chezmoi status' second column
}

// Label matches the four states the design's Dotfiles tree distinguishes:
// applied, modified (local drift), pending (apply would change it), and —
// separately, since chezmoi status never lists ignored paths — ignored.
func (d DriftState) Label() string {
	switch {
	case d.DriftCode != ' ':
		return "modified"
	case d.ApplyCode != ' ':
		return "pending"
	default:
		return "applied"
	}
}

// Note is a short human line for the Overview "Local drift" table.
func (d DriftState) Note() string {
	code := d.DriftCode
	verb := "modified"
	switch code {
	case 'A':
		verb = "added"
	case 'D':
		verb = "deleted"
	case 'M':
		verb = "modified"
	case ' ':
		switch d.ApplyCode {
		case 'A':
			return "not yet applied"
		case 'D':
			return "will be removed on apply"
		case 'M':
			return "source changed, apply pending"
		case 'R':
			return "script will run on apply"
		}
	}
	return verb + " since last apply"
}

// Status runs `chezmoi status` and parses its two-column, git-style output.
func (r *Repo) Status() ([]DriftState, error) {
	out, err := r.run("status", "--path-style", "relative")
	if err != nil {
		return nil, err
	}
	var entries []DriftState
	for _, line := range strings.Split(strings.TrimRight(string(out), "\n"), "\n") {
		if line == "" {
			continue
		}
		if len(line) < 4 {
			continue
		}
		entries = append(entries, DriftState{
			DriftCode: line[0],
			ApplyCode: line[1],
			Path:      line[3:],
		})
	}
	return entries, nil
}

// Managed lists every target path chezmoi manages, relative to the
// destination directory, alphabetically — the basis of the source tree.
func (r *Repo) Managed() ([]string, error) {
	out, err := r.run("managed", "--path-style", "relative")
	if err != nil {
		return nil, err
	}
	return splitLines(out), nil
}

// Ignored lists paths chezmoi ignores for this machine (computed from
// .chezmoiignore.tmpl — never stored, always recomputed).
func (r *Repo) Ignored() ([]string, error) {
	out, err := r.run("ignored")
	if err != nil {
		return nil, err
	}
	return splitLines(out), nil
}

// Data returns the computed chezmoi template data (machine.distro,
// features.*, pixi.home, ...) as a generic JSON tree.
func (r *Repo) Data() (map[string]any, error) {
	out, err := r.run("data", "--format", "json")
	if err != nil {
		return nil, err
	}
	var data map[string]any
	if err := json.Unmarshal(out, &data); err != nil {
		return nil, fmt.Errorf("repo: parse chezmoi data: %w", err)
	}
	return data, nil
}

func splitLines(out []byte) []string {
	lines := strings.Split(strings.TrimRight(string(out), "\n"), "\n")
	result := lines[:0]
	for _, l := range lines {
		if l != "" {
			result = append(result, l)
		}
	}
	sort.Strings(result)
	return result
}
