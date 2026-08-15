// Package dfedit is the dotfiles source editor: it edits the *source* file
// in the chezmoi source tree, never the applied file in $HOME directly —
// "chezmoi edit" — then runs a targeted `chezmoi apply` to bring the
// destination in line, same as running `chezmoi edit <file>` yourself and
// exiting your editor.
package dfedit

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/walcark/dotfiles/converge/internal/repo"
)

// File is what the editor screen needs.
type File struct {
	TargetPath string // relative to $HOME, e.g. ".bashrc.d/core/05-path.sh"
	SourcePath string // absolute path in the chezmoi source tree
	IsTemplate bool
	Content    string // raw source content — the template source, for a .tmpl
}

// Read loads a managed file's source content.
func Read(r *repo.Repo, targetRelPath string) (File, error) {
	sourcePath, err := r.SourcePathFor(targetRelPath)
	if err != nil {
		return File{}, fmt.Errorf("dfedit: %w", err)
	}
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		return File{}, fmt.Errorf("dfedit: read %s: %w", sourcePath, err)
	}
	return File{
		TargetPath: targetRelPath, SourcePath: sourcePath,
		IsTemplate: strings.HasSuffix(sourcePath, ".tmpl"),
		Content:    string(data),
	}, nil
}

// Preview renders a .tmpl source through chezmoi's template engine — call
// only for IsTemplate files; for anything else the content already is
// what gets applied.
func Preview(r *repo.Repo, content string) (string, error) {
	rendered, err := r.ExecuteTemplate(content)
	if err != nil {
		return "", fmt.Errorf("dfedit: preview: %w", err)
	}
	return rendered, nil
}

// Write saves new source content and applies it to the destination.
//
// Guardrail from the design: refuse if the chezmoi source git tree is
// dirty. Not a formality — chezmoi has no merge machinery; if there's
// already an uncommitted manual edit sitting in the source tree, writing
// over it here would silently discard it the moment `chezmoi apply` runs.
func Write(r *repo.Repo, targetRelPath, newContent string) error {
	sourcePath, err := r.SourcePathFor(targetRelPath)
	if err != nil {
		return fmt.Errorf("dfedit: %w", err)
	}

	dirty, err := gitDirty(r.RootDir)
	if err != nil {
		return fmt.Errorf("dfedit: check git status: %w", err)
	}
	if dirty {
		return fmt.Errorf("dfedit: refusing to write — %s has uncommitted changes; commit or stash them first", r.RootDir)
	}

	if err := os.WriteFile(sourcePath, []byte(newContent), 0o644); err != nil {
		return fmt.Errorf("dfedit: write %s: %w", sourcePath, err)
	}

	if err := r.Apply(targetRelPath); err != nil {
		return fmt.Errorf("dfedit: chezmoi apply: %w", err)
	}
	return nil
}

func gitDirty(dir string) (bool, error) {
	out, err := exec.Command("git", "-C", dir, "status", "--porcelain").Output()
	if err != nil {
		return false, err
	}
	return len(strings.TrimSpace(string(out))) > 0, nil
}
