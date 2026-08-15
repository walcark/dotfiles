// Package sandbox runs the design's "Sandbox apply" test gate: a
// throwaway container (`podman run --rm fedora:42`, matching the design
// exactly), mounting the branch's repo tree read-only.
//
// Scope, honestly: it installs ansible-core fresh via dnf and runs
// --syntax-check on both playbooks — real, clean-room validation that the
// YAML and Jinja are well-formed, independent of whatever's already
// installed on the host. It does *not* attempt a full --check run inside
// the container: most tasks shell out to pixi/chezmoi/flatpak, none of
// which exist in a bare fedora image, so every one of them would fail for
// reasons that have nothing to do with whether the merge/edit itself is
// correct. A full simulated apply belongs to a container image that
// actually bootstraps those tools first — worth building later, not
// pretended here.
package sandbox

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const image = "fedora:42"

// Result is one container test run.
type Result struct {
	OK     bool
	Output string
}

// Run mounts repoRoot read-only into a fresh container and syntax-checks
// both playbooks.
func Run(repoRoot string) (Result, error) {
	if _, err := exec.LookPath("podman"); err != nil {
		return Result{}, fmt.Errorf("sandbox: podman not found on PATH: %w", err)
	}

	script := strings.Join([]string{
		"set -e",
		"dnf install -y -q ansible-core >/dev/null",
		// community.general provides the flatpak/flatpak_remote modules
		// several roles use — --syntax-check resolves module names, so
		// without this collection installed every one of them errors as
		// "couldn't resolve module/action", indistinguishable at a glance
		// from an actual typo. Found the same way as the :z mount fix:
		// running this against a real merge branch, not guessed upfront.
		"ansible-galaxy collection install community.general >/dev/null",
		"cd /repo",
		"ansible-playbook --syntax-check ansible/playbook.yml",
		"ansible-playbook --syntax-check ansible/absent.yml",
	}, " && ")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "podman", "run", "--rm",
		// `:z` relabels the mount for SELinux (enforcing by default on
		// Fedora): without it the mount looks empty to the container
		// ("Permission denied" on the directory itself, not the files in
		// it) — found running this against a real merge branch. `z`
		// (shared label), not `Z` (private): this is read-only and never
		// exclusive to one container.
		"-v", repoRoot+":/repo:ro,z",
		image, "bash", "-c", script)
	out, err := cmd.CombinedOutput()

	return Result{OK: err == nil, Output: string(out)}, nil
}
