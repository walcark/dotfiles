// Command converge serves the Converge UI on 127.0.0.1 and opens the
// default browser. Phase 1: read-only (Overview, Dotfiles source tree,
// binaries & scripts) — nothing here writes to the machine or the repo.
package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/walcark/dotfiles/converge/internal/repo"
	"github.com/walcark/dotfiles/converge/internal/webui"
)

func main() {
	r, err := repo.Detect()
	if err != nil {
		log.Fatalf("converge: %v", err)
	}

	pixiHome := os.Getenv("PIXI_HOME")
	if pixiHome == "" {
		home, _ := os.UserHomeDir()
		pixiHome = filepath.Join(home, ".pixi")
	}

	app, err := webui.New(r, pixiHome)
	if err != nil {
		log.Fatalf("converge: %v", err)
	}

	mux := http.NewServeMux()
	app.Routes(mux)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatalf("converge: listen: %v", err)
	}
	url := fmt.Sprintf("http://%s/", ln.Addr())

	fmt.Printf("Converge — repo %s\n", r.RootDir)
	fmt.Printf("Serving on %s (Ctrl-C to stop)\n", url)
	openBrowser(url)

	if err := http.Serve(ln, mux); err != nil {
		log.Fatalf("converge: %v", err)
	}
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		return
	}
	_ = cmd.Start() // best-effort: printing the URL above is the fallback
}
