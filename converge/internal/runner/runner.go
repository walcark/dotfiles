// Package runner executes ansible-playbook and streams structured events
// from it, using the converge_json callback plugin (see callback/) instead
// of ansible-core's built-in `json` callback, which buffers everything and
// only writes it once the whole play finishes — no good for a live run log.
//
// Guardrails from the design (ansible/roles/README.md's spirit extends
// here): Apply always runs --check first, in the same Run, and refuses to
// proceed past it on any failure — "never apply without a passing --check
// immediately before" is enforced by construction, not by trusting the
// caller to have run one recently. playbook.yml is `become: false` and
// Start never passes --ask-become-pass, so a task that unexpectedly needs
// become fails loudly (no stdin to read a password from) instead of
// prompting silently.
package runner

import (
	"bufio"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/walcark/dotfiles/converge/internal/repo"
)

//go:embed callback/converge_json.py
var callbackPluginSource []byte

// callbackPluginDir writes the embedded callback plugin to a stable file
// under the OS temp dir and returns its containing directory.
//
// Earlier this located the plugin via runtime.Caller(0) relative to this
// source file — which works when running `go run` from a checkout, but
// breaks completely for a built, deployed, or copied-to-another-machine
// binary: runtime.Caller bakes in the *build* machine's source path, which
// has no reason to exist on whatever machine actually runs the binary.
// Found the hard way, copying a binary built here straight to the VM.
// Embedding the plugin's bytes at compile time and writing them out at
// startup makes the binary self-contained, as it needs to be.
func callbackPluginDir() (string, error) {
	dir := filepath.Join(os.TempDir(), "converge-callback")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("runner: callback plugin dir: %w", err)
	}
	path := filepath.Join(dir, "converge_json.py")
	if err := os.WriteFile(path, callbackPluginSource, 0o644); err != nil {
		return "", fmt.Errorf("runner: write callback plugin: %w", err)
	}
	return dir, nil
}

// Event is one line the converge_json callback plugin emits.
type Event struct {
	Event  string `json:"event"`
	Task   string `json:"task,omitempty"`
	Status string `json:"status,omitempty"`
	Host   string `json:"host,omitempty"`
	Msg    string `json:"msg,omitempty"`
	Play   string `json:"play,omitempty"`

	OK          int `json:"ok,omitempty"`
	Changed     int `json:"changed,omitempty"`
	Failed      int `json:"failed,omitempty"`
	Skipped     int `json:"skipped,omitempty"`
	Unreachable int `json:"unreachable,omitempty"`

	At time.Time `json:"-"`
}

// Stage is which part of a Run is currently active — Apply always goes
// through "check" before "apply" ever starts.
type Stage string

const (
	StageCheck  Stage = "check"
	StageApply  Stage = "apply"
	StageAbsent Stage = "absent" // uninstalling layers dropped since the last apply — see ComputeAbsentPlan
)

// State is a Run's overall outcome.
type State string

const (
	StateRunning State = "running"
	StateOK      State = "ok"
	StateFailed  State = "failed"
)

// Run is one execution, tracked in memory (see Manager) so the web UI can
// poll it while it's in progress.
type Run struct {
	ID        string
	Mode      string // "check" | "apply" — what the caller asked for
	StartedAt time.Time
	EndedAt   time.Time

	mu     sync.Mutex
	stage  Stage
	state  State
	events []Event
	err    string
}

func (r *Run) append(e Event) {
	e.At = time.Now()
	r.mu.Lock()
	r.events = append(r.events, e)
	r.mu.Unlock()
}

// Snapshot is a race-free read of a Run's current state, for rendering.
type Snapshot struct {
	ID        string
	Mode      string
	Stage     Stage
	State     State
	Err       string
	StartedAt time.Time
	EndedAt   time.Time
	Events    []Event
}

func (r *Run) Snapshot() Snapshot {
	r.mu.Lock()
	defer r.mu.Unlock()
	events := make([]Event, len(r.events))
	copy(events, r.events)
	return Snapshot{
		ID: r.ID, Mode: r.Mode, Stage: r.stage, State: r.state, Err: r.err,
		StartedAt: r.StartedAt, EndedAt: r.EndedAt, Events: events,
	}
}

// Manager runs at most one playbook execution at a time — this is a
// single-user local tool; two concurrent ansible-playbook runs against the
// same machine would just race each other.
type Manager struct {
	Repo *repo.Repo

	// OnApplySuccess, if set, runs after a mode "apply" Run finishes with
	// StateOK — see internal/webui wiring this to the ledger.
	OnApplySuccess func()

	// ComputeAbsentPlan, if set, runs right after a successful apply
	// stage, before the run is marked finished: it compares the ledger
	// (what's actually installed, as of the last successful apply)
	// against the layers active now and returns which layers need
	// uninstalling and which of their tasks to skip because another
	// still-active layer also needs that package — see internal/refcount.
	// A nil or empty `layers` return means there's nothing to do, and the
	// absent stage is skipped entirely.
	ComputeAbsentPlan func() (layers, skip []string)

	mu  sync.Mutex
	run *Run
	seq int
}

func NewManager(r *repo.Repo) *Manager { return &Manager{Repo: r} }

// Current returns the most recent run, if any.
func (m *Manager) Current() (*Run, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.run, m.run != nil
}

// StartCheck runs `ansible-playbook --check --diff` — never mutates the
// machine (subject to individual modules' check-mode support).
func (m *Manager) StartCheck() (*Run, error) { return m.start("check") }

// StartApply always runs the check stage first, in-process, before ever
// invoking a real apply — see the package doc.
func (m *Manager) StartApply() (*Run, error) { return m.start("apply") }

func (m *Manager) start(mode string) (*Run, error) {
	m.mu.Lock()
	if m.run != nil {
		snap := m.run.Snapshot()
		if snap.State == StateRunning {
			m.mu.Unlock()
			return nil, fmt.Errorf("runner: a run is already in progress")
		}
	}
	m.seq++
	run := &Run{
		ID: fmt.Sprintf("run-%d", m.seq), Mode: mode,
		StartedAt: time.Now(), stage: StageCheck, state: StateRunning,
	}
	m.run = run
	m.mu.Unlock()

	go m.execute(run, mode)
	return run, nil
}

func (m *Manager) execute(run *Run, mode string) {
	mainPlaybook := "ansible/playbook.yml"

	if err := m.runPlaybook(run, StageCheck, mainPlaybook, nil, []string{"--check", "--diff"}); err != nil {
		m.finish(run, StateFailed, err.Error())
		return
	}
	if run.Snapshot().State == StateFailed {
		return // runPlaybook itself already recorded the failure from ansible's own stats
	}
	if mode == "check" {
		m.finish(run, StateOK, "")
		return
	}
	if err := m.runPlaybook(run, StageApply, mainPlaybook, nil, nil); err != nil {
		m.finish(run, StateFailed, err.Error())
		return
	}
	if run.Snapshot().State == StateFailed {
		return
	}

	if m.ComputeAbsentPlan != nil {
		layers, skip := m.ComputeAbsentPlan()
		if len(layers) > 0 {
			vars := map[string]string{"absent_layers": jsonList(layers)}
			if len(skip) > 0 {
				vars["absent_skip"] = jsonList(skip)
			}
			if err := m.runPlaybook(run, StageAbsent, "ansible/absent.yml", vars, nil); err != nil {
				m.finish(run, StateFailed, err.Error())
				return
			}
			if run.Snapshot().State == StateFailed {
				return
			}
		}
	}

	m.finish(run, StateOK, "")
	if m.OnApplySuccess != nil {
		m.OnApplySuccess()
	}
}

// jsonList renders a Go string slice as a JSON array literal, suitable for
// `-e 'var=[...]'` — ansible parses -e values that look like JSON as
// structured data rather than a literal string.
func jsonList(items []string) string {
	b, err := json.Marshal(items)
	if err != nil {
		return "[]"
	}
	return string(b)
}

func (m *Manager) finish(run *Run, state State, errMsg string) {
	run.mu.Lock()
	run.state = state
	run.err = errMsg
	run.EndedAt = time.Now()
	run.mu.Unlock()
}

func (m *Manager) runPlaybook(run *Run, stage Stage, playbook string, extraVars map[string]string, extraArgs []string) error {
	run.mu.Lock()
	run.stage = stage
	run.mu.Unlock()

	callbackDir, err := callbackPluginDir()
	if err != nil {
		return err
	}

	args := []string{
		"-i", filepath.Join(m.Repo.RootDir, "ansible", "inventory.ini"),
		"-e", "dotfiles_url=https://github.com/walcark/dotfiles.git",
	}
	for k, v := range extraVars {
		args = append(args, "-e", k+"="+v)
	}
	args = append(args, extraArgs...)
	args = append(args, playbook)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ansible-playbook", args...)
	cmd.Dir = m.Repo.RootDir
	cmd.Env = append(cmd.Environ(),
		"ANSIBLE_STDOUT_CALLBACK=converge_json",
		"ANSIBLE_CALLBACK_PLUGINS="+callbackDir,
		"ANSIBLE_LOAD_CALLBACK_PLUGINS=1",
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("runner: stdout pipe: %w", err)
	}
	cmd.Stderr = nil // ansible's own errors show up as `failed`/`unreachable` events; stderr is mostly Python noise

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("runner: start ansible-playbook: %w", err)
	}

	failed := false
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		var e Event
		if err := json.Unmarshal(scanner.Bytes(), &e); err != nil {
			continue // not one of our JSON lines (stray plugin/module chatter) — skip rather than choke the run log on it
		}
		run.append(e)
		if e.Status == "failed" || e.Status == "unreachable" || e.Failed > 0 || e.Unreachable > 0 {
			failed = true
		}
	}

	waitErr := cmd.Wait()
	if failed {
		run.mu.Lock()
		run.state = StateFailed
		run.EndedAt = time.Now()
		run.mu.Unlock()
		return nil // the failure is already recorded as events; don't also surface the nonzero exit as a generic error
	}
	if waitErr != nil {
		return fmt.Errorf("runner: ansible-playbook: %w", waitErr)
	}
	return nil
}
