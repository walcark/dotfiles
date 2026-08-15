// Package pixiglobal reads which pixi global environments are actually
// installed, straight from pixi's own manifest — not from what any layer
// merely declares it should install, so it reflects the machine as it is.
package pixiglobal

import (
	"bufio"
	"os"
	"regexp"
)

var envHeader = regexp.MustCompile(`^\[envs\.([^\]]+)\]$`)

// List returns the names of the installed pixi global environments (the
// key each one was installed under — e.g. "starship", "python" — not the
// list of binaries each one exposes), read from
// <pixiHome>/manifests/pixi-global.toml.
func List(pixiHome string) ([]string, error) {
	f, err := os.Open(pixiHome + "/manifests/pixi-global.toml")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var names []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if m := envHeader.FindStringSubmatch(scanner.Text()); m != nil {
			names = append(names, m[1])
		}
	}
	return names, scanner.Err()
}
