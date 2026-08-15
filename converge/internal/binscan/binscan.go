// Package binscan replicates home/bin/executable_lsbin's own logic in Go,
// rather than shelling out to it and parsing its ANSI-formatted text: same
// rule (a shebang line makes it a script, its absence makes it a binary),
// same source of the description (the `# Description: ...` comment on one
// of the first 10 lines), structured instead of scraped.
package binscan

import (
	"bufio"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Entry is one executable found under ~/bin.
type Entry struct {
	Name        string // basename
	Path        string // full path
	Subdir      string // relative subdirectory under ~/bin, "" at the top level
	IsScript    bool   // has a #! shebang
	Description string // from "# Description: ..." on one of the first 10 lines; "" if absent
}

// Scan walks binDir (normally $HOME/bin) the same way lsbin does: files
// only, skips a lib/ subdirectory (helper code, not a user-facing entry),
// and classifies each by shebang presence.
func Scan(binDir string) ([]Entry, error) {
	var entries []Entry

	err := filepath.WalkDir(binDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "lib" {
				return filepath.SkipDir
			}
			return nil
		}
		info, err := d.Info()
		if err != nil || info.Mode()&0o111 == 0 {
			return nil // not executable
		}

		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()

		scanner := bufio.NewScanner(f)
		var firstLine string
		var lines []string
		for i := 0; i < 10 && scanner.Scan(); i++ {
			line := scanner.Text()
			if i == 0 {
				firstLine = line
			}
			lines = append(lines, line)
		}

		rel, _ := filepath.Rel(binDir, filepath.Dir(path))
		if rel == "." {
			rel = ""
		}

		e := Entry{
			Name:     d.Name(),
			Path:     path,
			Subdir:   rel,
			IsScript: strings.HasPrefix(firstLine, "#!"),
		}
		for _, line := range lines {
			if desc, ok := strings.CutPrefix(line, "# Description:"); ok {
				e.Description = strings.TrimSpace(desc)
				break
			}
		}
		entries = append(entries, e)
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(entries, func(i, j int) bool { return entries[i].Path < entries[j].Path })
	return entries, nil
}

// Scripts filters Entry list to those with a shebang.
func Scripts(entries []Entry) []Entry { return filterBy(entries, true) }

// Binaries filters Entry list to those without a shebang.
func Binaries(entries []Entry) []Entry { return filterBy(entries, false) }

func filterBy(entries []Entry, script bool) []Entry {
	var out []Entry
	for _, e := range entries {
		if e.IsScript == script {
			out = append(out, e)
		}
	}
	return out
}
