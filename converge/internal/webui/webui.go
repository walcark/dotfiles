// Package webui serves the Converge read-only UI (Phase 1): Overview and
// Dotfiles (source tree + binaries/scripts). Everything it shows is read
// straight from chezmoi, the ansible/roles/*/meta/layer.yml manifests, and
// the filesystem — nothing here writes anything.
package webui

import (
	"embed"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/walcark/dotfiles/converge/internal/binscan"
	"github.com/walcark/dotfiles/converge/internal/dfview"
	"github.com/walcark/dotfiles/converge/internal/groupvars"
	"github.com/walcark/dotfiles/converge/internal/manifest"
	"github.com/walcark/dotfiles/converge/internal/pixiglobal"
	"github.com/walcark/dotfiles/converge/internal/repo"
)

//go:embed templates/*.html
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS

// App holds everything a request handler needs. It re-reads chezmoi and the
// filesystem on every request rather than caching: this is a single-user
// local tool, and showing stale state would defeat Phase 1's entire point.
type App struct {
	Repo     *repo.Repo
	pages    map[string]*template.Template
	pixiHome string
}

var layerIcons = map[string]string{
	"core":    "cube",
	"desktop": "desktop",
	"dev":     "code",
	"drawing": "palette-fill",
	"gaming":  "game-controller",
	"gnome":   "app-window",
}

func iconFor(layerID string) string {
	if icon, ok := layerIcons[layerID]; ok {
		return icon
	}
	return "package"
}

// New builds the App and parses templates. pixiHome is used to find the
// installed pixi global environments for the Binaries tab.
//
// Each page gets its own *template.Template, combining layout.html with
// exactly its own content file: parsing every "templates/*.html" file into
// one shared set would leave only one {{define "content"}} standing (the
// last one parsed wins for every page), so pages are kept in separate sets.
func New(r *repo.Repo, pixiHome string) (*App, error) {
	pages := map[string]*template.Template{}
	for _, page := range []string{"overview", "dotfiles"} {
		tmpl, err := template.ParseFS(templateFS, "templates/layout.html", "templates/"+page+".html")
		if err != nil {
			return nil, fmt.Errorf("webui: parse templates for %s: %w", page, err)
		}
		pages[page] = tmpl
	}
	return &App{Repo: r, pages: pages, pixiHome: pixiHome}, nil
}

// Routes registers the app's handlers on mux.
func (a *App) Routes(mux *http.ServeMux) {
	mux.Handle("/static/", http.FileServer(http.FS(staticFS)))
	mux.HandleFunc("/", a.handleOverview)
	mux.HandleFunc("/overview", a.handleOverview)
	mux.HandleFunc("/dotfiles", a.handleDotfiles)
}

// — shared shell data —

type navItem struct {
	Label   string
	Icon    string
	Href    string
	Enabled bool
	Active  bool
	Badge   string
}

type machineInfo struct {
	Hostname string
	Branch   string
	Tags     []string
}

func navFor(active string) []navItem {
	items := []navItem{
		{Label: "Overview", Icon: "squares-four", Href: "/overview", Enabled: true},
		{Label: "Layers", Icon: "stack", Badge: "Phase 3"},
		{Label: "Source edits", Icon: "git-diff", Badge: "Phase 6"},
		{Label: "Dotfiles", Icon: "folder-notch", Href: "/dotfiles", Enabled: true},
		{Label: "Machines", Icon: "hard-drives", Badge: "Phase 3"},
		{Label: "Run log", Icon: "terminal-window", Badge: "Phase 2"},
	}
	for i := range items {
		if items[i].Enabled && strings.EqualFold(items[i].Label, active) {
			items[i].Active = true
		}
	}
	return items
}

func (a *App) machine() machineInfo {
	host, _ := os.Hostname()

	branch := "?"
	if out, err := exec.Command("git", "-C", a.Repo.RootDir, "rev-parse", "--abbrev-ref", "HEAD").Output(); err == nil {
		branch = strings.TrimSpace(string(out))
	}

	var tags []string
	if data, err := a.Repo.Data(); err == nil {
		if m, ok := data["machine"].(map[string]any); ok {
			if distro, ok := m["distro"].(string); ok && distro != "" {
				tags = append(tags, distro)
			}
		}
		if f, ok := data["features"].(map[string]any); ok {
			for _, key := range []string{"dev", "hpc", "admin"} {
				if on, _ := f[key].(bool); on {
					tags = append(tags, key)
				}
			}
		}
	}

	return machineInfo{Hostname: host, Branch: branch, Tags: tags}
}

// sourceEditCount counts files that differ from the last commit in the
// dotfiles repo itself (ansible/ + home/) — a real, read-only proxy for
// "source edits" until Phase 6 gives it a proper meaning (staged patches).
func (a *App) sourceEditCount() int {
	out, err := exec.Command("git", "-C", a.Repo.RootDir, "status", "--porcelain").Output()
	if err != nil {
		return 0
	}
	n := 0
	for _, line := range strings.Split(strings.TrimRight(string(out), "\n"), "\n") {
		if strings.TrimSpace(line) != "" {
			n++
		}
	}
	return n
}

func (a *App) render(w http.ResponseWriter, page string, data map[string]any) {
	tmpl, ok := a.pages[page]
	if !ok {
		http.Error(w, "webui: unknown page "+page, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// — Overview —

// Color fields are template.CSS rather than string: html/template's
// contextual auto-escaper can't statically prove a plain string dropped
// into a `style="color:{{.}}"` attribute is safe CSS and replaces it with
// the literal text "ZgotmplZ" — template.CSS marks a value this code
// builds (not user input) as pre-vetted, so the color renders as intended.
type statTile struct {
	Label, Value, Unit string
	Color              template.CSS
}
type driftRow struct {
	Path, Note, State, TagClass, Icon string
	Color                             template.CSS
}
type activeLayerRow struct {
	Name         string
	Icon         string
	PackageCount int
}

func (a *App) handleOverview(w http.ResponseWriter, req *http.Request) {
	layers, err := manifest.LoadAll(a.Repo.RootDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	gv, err := groupvars.Load(a.Repo.RootDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	status, err := a.Repo.Status()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	managed, err := a.Repo.Managed()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	byID := make(map[string]manifest.Layer, len(layers))
	for _, l := range layers {
		byID[l.ID] = l
	}

	var active []activeLayerRow
	activeCount := 0
	packageCount := 0
	// core has no `layers.core` flag — it's unconditional in playbook.yml.
	if l, ok := byID["core"]; ok {
		active = append(active, activeLayerRow{Name: l.Name, Icon: iconFor("core"), PackageCount: l.PackageCount()})
		activeCount++
		packageCount += l.PackageCount()
	}
	var keys []string
	for k := range gv.Layers {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if !gv.Layers[k] {
			continue
		}
		l, ok := byID[k]
		if !ok {
			continue
		}
		active = append(active, activeLayerRow{Name: l.Name, Icon: iconFor(k), PackageCount: l.PackageCount()})
		activeCount++
		packageCount += l.PackageCount()
	}

	var drift []driftRow
	for _, s := range status {
		if s.DriftCode == ' ' {
			continue // no local drift — an apply-pending-only entry belongs on the Dotfiles tree, not here
		}
		icon, color, tagClass := "pencil-simple", template.CSS("#b5abfc"), "tag-accent-2"
		state := "modified"
		switch s.DriftCode {
		case 'A':
			icon, state = "plus-circle", "added"
		case 'D':
			icon, color, state = "minus-circle", template.CSS("#e0a9a9"), "deleted"
		}
		if state == "deleted" {
			tagClass = "tag-outline"
		}
		drift = append(drift, driftRow{
			Path: s.Path, Note: s.Note(), State: state,
			TagClass: tagClass, Icon: icon, Color: color,
		})
	}

	const textColor = template.CSS("var(--color-text)")
	stats := []statTile{
		{Label: "Active layers", Value: fmt.Sprint(activeCount), Unit: "layers", Color: textColor},
		{Label: "Packages", Value: fmt.Sprint(packageCount), Unit: "declared", Color: textColor},
		{Label: "Managed files", Value: fmt.Sprint(len(managed)), Unit: "files", Color: textColor},
		{Label: "Source edits", Value: fmt.Sprint(a.sourceEditCount()), Unit: "vs last commit", Color: textColor},
	}

	a.render(w, "overview", map[string]any{
		"Title": "Overview", "Kicker": "State",
		"Nav": navFor("Overview"), "Machine": a.machine(),
		"Stats": stats, "Drift": drift, "ActiveLayers": active,
	})
}

// — Dotfiles —

type treeRowVM struct {
	Name     string
	IsDir    bool
	Status   string
	IndentPx int
}
type scriptVM struct{ Name, Description string }
type pixiVM struct {
	Name   string
	Layers []string
}

func (a *App) handleDotfiles(w http.ResponseWriter, req *http.Request) {
	tab := req.URL.Query().Get("tab")
	if tab == "" {
		tab = "tree"
	}

	data := map[string]any{
		"Title": "Dotfiles", "Kicker": "home/",
		"Nav": navFor("Dotfiles"), "Machine": a.machine(),
		"Tab": tab,
	}

	switch tab {
	case "tree":
		managed, err := a.Repo.Managed()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		ignored, err := a.Repo.Ignored()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		status, err := a.Repo.Status()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		destDir := os.Getenv("HOME")
		if d, err := a.Repo.Data(); err == nil {
			if c, ok := d["chezmoi"].(map[string]any); ok {
				if dd, ok := c["destDir"].(string); ok && dd != "" {
					destDir = dd
				}
			}
		}
		rows := dfview.Rows(managed, ignored, status, destDir)
		vmRows := make([]treeRowVM, len(rows))
		for i, r := range rows {
			vmRows[i] = treeRowVM{Name: r.Name, IsDir: r.IsDir, Status: r.Status, IndentPx: 12 + r.Depth*16}
		}
		data["Rows"] = vmRows
		data["Counts"] = dfview.Counts(rows)

	case "bin":
		binDir := filepath.Join(os.Getenv("HOME"), "bin")
		entries, err := binscan.Scan(binDir)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var scripts []scriptVM
		for _, e := range binscan.Scripts(entries) {
			scripts = append(scripts, scriptVM{Name: e.Name, Description: e.Description})
		}
		data["Scripts"] = scripts

		layers, err := manifest.LoadAll(a.Repo.RootDir)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		owners := make(map[string][]string) // pixi global env name -> layer IDs that declare it
		for _, l := range layers {
			for _, t := range l.Tasks {
				if t.Kind != "pixi" {
					continue
				}
				for _, pkg := range t.Provides {
					name, _, _ := strings.Cut(pkg, "=")
					owners[name] = append(owners[name], l.ID)
				}
			}
		}
		globals, err := pixiglobal.List(a.pixiHome)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		sort.Strings(globals)
		var pixis []pixiVM
		for _, g := range globals {
			pixis = append(pixis, pixiVM{Name: g, Layers: owners[g]})
		}
		data["PixiGlobals"] = pixis
	}

	a.render(w, "dotfiles", data)
}
