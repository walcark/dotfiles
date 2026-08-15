// Package authoring holds the git primitives Phase 6 builds on: every
// structural change (add/move/merge tasks) happens on a branch, gets
// tested, then becomes a PR — main is never touched directly and nothing
// is ever pushed without the test gates having passed first.
package authoring

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Dirty reports whether the source tree has uncommitted changes — the
// same guardrail dfedit uses, checked again here because authoring makes
// far more invasive edits.
func Dirty(repoRoot string) (bool, error) {
	out, err := exec.Command("git", "-C", repoRoot, "status", "--porcelain").Output()
	if err != nil {
		return false, err
	}
	return len(strings.TrimSpace(string(out))) > 0, nil
}

// CurrentBranch returns the checked-out branch name.
func CurrentBranch(repoRoot string) (string, error) {
	out, err := exec.Command("git", "-C", repoRoot, "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// NewBranch refuses on a dirty tree, then creates and checks out a new
// branch off the current HEAD. slug becomes part of the branch name
// (lowercased, spaces to dashes) with a timestamp suffix so repeated runs
// never collide.
func NewBranch(repoRoot, slug string) (string, error) {
	dirty, err := Dirty(repoRoot)
	if err != nil {
		return "", fmt.Errorf("authoring: check git status: %w", err)
	}
	if dirty {
		return "", fmt.Errorf("authoring: refusing to start — %s has uncommitted changes; commit or stash them first", repoRoot)
	}

	name := "converge/" + slug + "-" + time.Now().Format("20060102-150405")
	if err := run(repoRoot, "checkout", "-b", name); err != nil {
		return "", fmt.Errorf("authoring: create branch: %w", err)
	}
	return name, nil
}

// CheckoutMain switches back to main — used both on success (after
// pushing) and as a best-effort cleanup on failure, so a half-finished
// authoring attempt never leaves the working tree stuck on a stray
// branch.
func CheckoutMain(repoRoot string) error {
	return run(repoRoot, "checkout", "main")
}

// Commit stages everything in the tree and commits. Left uncommitted (not
// pushed) if the caller's test gates haven't passed — see
// internal/webui's orchestration.
func Commit(repoRoot, message string) error {
	if err := run(repoRoot, "add", "-A"); err != nil {
		return fmt.Errorf("authoring: git add: %w", err)
	}
	if err := run(repoRoot, "commit", "-m", message); err != nil {
		return fmt.Errorf("authoring: git commit: %w", err)
	}
	return nil
}

// Push pushes branch to origin, setting the upstream.
func Push(repoRoot, branch string) error {
	if err := run(repoRoot, "push", "-u", "origin", branch); err != nil {
		return fmt.Errorf("authoring: git push: %w", err)
	}
	return nil
}

// OpenPR opens a pull request for branch against main via the gh CLI and
// returns its URL.
func OpenPR(repoRoot, branch, title, body string) (string, error) {
	cmd := exec.Command("gh", "pr", "create",
		"--head", branch, "--base", "main",
		"--title", title, "--body", body)
	cmd.Dir = repoRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("authoring: gh pr create: %w (%s)", err, strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}

func run(repoRoot string, args ...string) error {
	cmd := exec.Command("git", append([]string{"-C", repoRoot}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git %s: %w (%s)", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
	}
	return nil
}
