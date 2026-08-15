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
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/walcark/dotfiles/converge/internal/activelayers"
	"github.com/walcark/dotfiles/converge/internal/authoring"
	"github.com/walcark/dotfiles/converge/internal/binscan"
	"github.com/walcark/dotfiles/converge/internal/dfedit"
	"github.com/walcark/dotfiles/converge/internal/dfview"
	"github.com/walcark/dotfiles/converge/internal/envlocal"
	"github.com/walcark/dotfiles/converge/internal/ledger"
	"github.com/walcark/dotfiles/converge/internal/machinevars"
	"github.com/walcark/dotfiles/converge/internal/manifest"
	"github.com/walcark/dotfiles/converge/internal/pixiglobal"
	"github.com/walcark/dotfiles/converge/internal/refcount"
	"github.com/walcark/dotfiles/converge/internal/repo"
	"github.com/walcark/dotfiles/converge/internal/rolemerge"
	"github.com/walcark/dotfiles/converge/internal/runner"
	"github.com/walcark/dotfiles/converge/internal/sandbox"
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
	Runner   *runner.Manager
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
	for _, page := range []string{"overview", "dotfiles", "run", "layers", "dfedit", "sourceedits"} {
		tmpl, err := template.ParseFS(templateFS, "templates/layout.html", "templates/"+page+".html")
		if err != nil {
			return nil, fmt.Errorf("webui: parse templates for %s: %w", page, err)
		}
		pages[page] = tmpl
	}
	app := &App{Repo: r, Runner: runner.NewManager(r), pages: pages, pixiHome: pixiHome}
	app.Runner.OnApplySuccess = app.updateLedger
	app.Runner.ComputeAbsentPlan = app.computeAbsentPlan
	return app, nil
}

// Routes registers the app's handlers on mux.
func (a *App) Routes(mux *http.ServeMux) {
	mux.Handle("/static/", http.FileServer(http.FS(staticFS)))
	mux.HandleFunc("/", a.handleOverview)
	mux.HandleFunc("/overview", a.handleOverview)
	mux.HandleFunc("/dotfiles", a.handleDotfiles)
	mux.HandleFunc("/run", a.handleRun)
	mux.HandleFunc("/run/check", a.handleRunCheck)
	mux.HandleFunc("/run/apply", a.handleRunApply)
	mux.HandleFunc("/layers", a.handleLayers)
	mux.HandleFunc("/layers/toggle", a.handleLayerToggle)
	mux.HandleFunc("/dotfiles/env/save", a.handleEnvSave)
	mux.HandleFunc("/dotfiles/edit", a.handleDotfilesEdit)
	mux.HandleFunc("/dotfiles/edit/save", a.handleDotfilesEditSave)
	mux.HandleFunc("/source-edits", a.handleSourceEdits)
	mux.HandleFunc("/source-edits/merge", a.handleMergeRoles)
}

// updateLedger runs after every successful Apply — see runner.Manager's
// OnApplySuccess. Best-effort: a ledger write failure shouldn't be able to
// make a successful apply look like anything other than a success, so
// errors are swallowed here rather than surfaced on the run.
func (a *App) updateLedger() {
	layers, err := manifest.LoadAll(a.Repo.RootDir)
	if err != nil {
		return
	}
	active, _, err := activelayers.Load(a.Repo.RootDir, os.Getenv("HOME"))
	if err != nil {
		return
	}
	_ = ledger.Snapshot(layers, active).Save(os.Getenv("HOME"))
}

// computeAbsentPlan runs after a successful apply stage, before the run
// finishes — see runner.Manager.ComputeAbsentPlan. It reads the ledger
// from *before* this run's apply stage, which is fine: updateLedger only
// runs once the whole run (apply, then this absent stage if any) is done,
// so the ledger still reflects the last successful run at this point.
func (a *App) computeAbsentPlan() (layersOut, skipOut []string) {
	layers, err := manifest.LoadAll(a.Repo.RootDir)
	if err != nil {
		return nil, nil
	}
	active, _, err := activelayers.Load(a.Repo.RootDir, os.Getenv("HOME"))
	if err != nil {
		return nil, nil
	}
	led, err := ledger.Load(os.Getenv("HOME"))
	if err != nil {
		return nil, nil
	}
	plan := refcount.Compute(layers, led, active)
	return plan.Layers, plan.Skip
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
		{Label: "Layers", Icon: "stack", Href: "/layers", Enabled: true},
		{Label: "Source edits", Icon: "git-diff", Href: "/source-edits", Enabled: true},
		{Label: "Dotfiles", Icon: "folder-notch", Href: "/dotfiles", Enabled: true},
		{Label: "Machines", Icon: "hard-drives", Badge: "Phase 3"},
		{Label: "Run log", Icon: "terminal-window", Href: "/run", Enabled: true},
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
	activeSet, _, err := activelayers.Load(a.Repo.RootDir, os.Getenv("HOME"))
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
	for k := range activeSet {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if !activeSet[k] {
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
	Path     string
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
			vmRows[i] = treeRowVM{Name: r.Name, Path: r.Path, IsDir: r.IsDir, Status: r.Status, IndentPx: 12 + r.Depth*16}
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

	case "env":
		home := os.Getenv("HOME")
		env, raw, err := envlocal.Load(home)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		expandEnv := envExpander(env)
		var pathVars []pathVarVM
		for _, pv := range env.PathVars {
			vm := pathVarVM{Name: pv.Name}
			for _, entry := range pv.Entries {
				expanded := os.Expand(entry, expandEnv)
				_, statErr := os.Stat(expanded)
				vm.Entries = append(vm.Entries, pathEntryVM{Value: entry, Exists: statErr == nil})
			}
			pathVars = append(pathVars, vm)
		}
		data["Exports"] = env.Exports
		data["PathVars"] = pathVars
		data["EnvMatchesFile"] = env.Matches(raw)
	}

	a.render(w, "dotfiles", data)
}

// envExpander resolves $VAR references inside a path variable's entries —
// both other exports in the same file (LIBS_PATH etc.) and the real
// process environment (HOME etc.), so existence checks reflect reality.
func envExpander(env envlocal.Env) func(string) string {
	values := map[string]string{}
	for _, e := range os.Environ() {
		if k, v, ok := strings.Cut(e, "="); ok {
			values[k] = v
		}
	}
	for _, e := range env.Exports {
		values[e.Key] = os.Expand(e.Value, func(k string) string { return values[k] })
	}
	return func(k string) string { return values[k] }
}

type pathEntryVM struct {
	Value  string
	Exists bool
}
type pathVarVM struct {
	Name    string
	Entries []pathEntryVM
}

func (a *App) handleEnvSave(w http.ResponseWriter, req *http.Request) {
	if err := req.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	keys := req.Form["export_key"]
	values := req.Form["export_value"]
	var env envlocal.Env
	for i, k := range keys {
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		v := ""
		if i < len(values) {
			v = values[i]
		}
		env.Exports = append(env.Exports, envlocal.Export{Key: k, Value: v})
	}

	names := req.Form["pathvar_name"]
	entries := req.Form["pathvar_entry"]
	// The form doesn't group entries per variable explicitly — reload the
	// existing structure to know how many entries belong to each name,
	// matching them back up in order.
	home := os.Getenv("HOME")
	existing, _, err := envlocal.Load(home)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	pos := 0
	for i, name := range names {
		count := 0
		if i < len(existing.PathVars) {
			count = len(existing.PathVars[i].Entries)
		}
		var vals []string
		for j := 0; j < count && pos < len(entries); j++ {
			vals = append(vals, entries[pos])
			pos++
		}
		env.PathVars = append(env.PathVars, envlocal.PathVar{Name: strings.TrimSpace(name), Entries: vals})
	}

	if err := env.Save(home); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, req, "/dotfiles?tab=env", http.StatusSeeOther)
}

// — Dotfiles source editor —

func (a *App) handleDotfilesEdit(w http.ResponseWriter, req *http.Request) {
	target := req.URL.Query().Get("path")
	if target == "" {
		http.Error(w, "webui: missing path", http.StatusBadRequest)
		return
	}
	file, err := dfedit.Read(a.Repo, target)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := map[string]any{
		"Title": file.TargetPath, "Kicker": "Dotfiles · edit",
		"Nav": navFor("Dotfiles"), "Machine": a.machine(),
		"File": file,
	}
	if req.URL.Query().Get("preview") == "1" && file.IsTemplate {
		if rendered, err := dfedit.Preview(a.Repo, file.Content); err == nil {
			data["Rendered"] = rendered
		} else {
			data["PreviewError"] = err.Error()
		}
	}
	a.render(w, "dfedit", data)
}

func (a *App) handleDotfilesEditSave(w http.ResponseWriter, req *http.Request) {
	target := req.FormValue("path")
	content := req.FormValue("content")
	if err := dfedit.Write(a.Repo, target, content); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, req, "/dotfiles/edit?path="+url.QueryEscape(target)+"&saved=1", http.StatusSeeOther)
}

// — Run log —

type logLineVM struct{ Text, Color string }

func (a *App) handleRun(w http.ResponseWriter, req *http.Request) {
	data := map[string]any{
		"Title": "Run log", "Kicker": "ansible-playbook",
		"Nav": navFor("Run log"), "Machine": a.machine(),
	}

	run, ok := a.Runner.Current()
	if !ok {
		data["HasRun"] = false
		a.render(w, "run", data)
		return
	}
	snap := run.Snapshot()

	stats := struct{ OK, Changed, Failed, Skipped int }{}
	var lines []logLineVM
	for _, e := range snap.Events {
		switch e.Event {
		case "play_start":
			lines = append(lines, logLineVM{Text: "PLAY [" + e.Play + "]", Color: "rgba(233,233,237,.55)"})
		case "task_start":
			lines = append(lines, logLineVM{Text: "» " + e.Task, Color: "#9184d9"})
		case "result":
			text := e.Status + ": " + e.Task
			if e.Msg != "" {
				text += " — " + e.Msg
			}
			color := "rgba(233,233,237,.55)"
			switch e.Status {
			case "ok":
				color = "#7fb98a"
			case "changed":
				color = "#b5abfc"
			case "failed", "unreachable":
				color = "#e0a9a9"
			}
			lines = append(lines, logLineVM{Text: text, Color: color})
		case "stats":
			stats.OK, stats.Changed, stats.Failed, stats.Skipped = e.OK, e.Changed, e.Failed, e.Skipped
			lines = append(lines, logLineVM{
				Text:  fmt.Sprintf("PLAY RECAP: ok=%d changed=%d failed=%d skipped=%d unreachable=%d", e.OK, e.Changed, e.Failed, e.Skipped, e.Unreachable),
				Color: "rgba(233,233,237,.7)",
			})
		}
	}

	headIcon, headColor, headTitle := "ph-circle-notch", "var(--color-accent)", "Running…"
	switch snap.State {
	case runner.StateOK:
		headIcon, headColor, headTitle = "ph-check-circle", "#7fb98a", "Finished"
	case runner.StateFailed:
		headIcon, headColor, headTitle = "ph-x-circle", "#e0a9a9", "Failed"
	}

	data["HasRun"] = true
	data["Running"] = snap.State == runner.StateRunning
	data["ID"] = snap.ID
	data["Mode"] = snap.Mode
	data["Stage"] = snap.Stage
	data["Err"] = snap.Err
	data["HeadIcon"] = headIcon
	data["HeadColor"] = template.CSS(headColor)
	data["HeadTitle"] = headTitle
	data["Stats"] = stats
	data["Lines"] = lines

	a.render(w, "run", data)
}

func (a *App) handleRunCheck(w http.ResponseWriter, req *http.Request) {
	if _, err := a.Runner.StartCheck(); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	http.Redirect(w, req, "/run", http.StatusSeeOther)
}

func (a *App) handleRunApply(w http.ResponseWriter, req *http.Request) {
	if _, err := a.Runner.StartApply(); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	http.Redirect(w, req, "/run", http.StatusSeeOther)
}

// — Layers —

type layerCardVM struct {
	ID           string
	Name         string
	Description  string
	Icon         string
	Active       bool
	Toggleable   bool // false for core: unconditional, no flag to flip
	PackageCount int
	TaskCount    int
	Reversible   int
	TasksTotal   int
}

func (a *App) handleLayers(w http.ResponseWriter, req *http.Request) {
	layers, err := manifest.LoadAll(a.Repo.RootDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	active, fromMachine, err := activelayers.Load(a.Repo.RootDir, os.Getenv("HOME"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var cards []layerCardVM
	for _, l := range layers {
		rev, total := l.ReversibleCount()
		cards = append(cards, layerCardVM{
			ID: l.ID, Name: l.Name, Description: l.Description, Icon: iconFor(l.ID),
			Active:       l.ID == "core" || active[l.ID],
			Toggleable:   l.ID != "core",
			PackageCount: l.PackageCount(), TaskCount: len(l.Tasks),
			Reversible: rev, TasksTotal: total,
		})
	}

	var orphans []orphanVM
	if led, err := ledger.Load(os.Getenv("HOME")); err == nil {
		plan := refcount.Compute(layers, led, active)
		for pkg, layerID := range plan.Remove {
			task := led.Packages[pkg].Task
			reversible := "unknown"
			for _, l := range layers {
				if l.ID != layerID {
					continue
				}
				for _, t := range l.Tasks {
					if t.ID == task {
						reversible = t.Reversible
					}
				}
			}
			orphans = append(orphans, orphanVM{Package: pkg, Layer: layerID, Reversible: reversible})
		}
		sort.Slice(orphans, func(i, j int) bool { return orphans[i].Package < orphans[j].Package })
	}

	a.render(w, "layers", map[string]any{
		"Title": "Layers", "Kicker": "ansible/roles",
		"Nav": navFor("Layers"), "Machine": a.machine(),
		"Cards": cards, "FromMachine": fromMachine, "Orphans": orphans,
	})
}

// orphanVM is a package the ledger says was installed by a layer that's no
// longer active, and reference counting confirms no other active layer
// still needs it — so Apply should still remove it (`reversible` is
// derived/explicit), or it's a `reversible: none` task that never will,
// which is worth knowing about even though nothing here can fix it.
type orphanVM struct {
	Package    string
	Layer      string
	Reversible string
}

func (a *App) handleLayerToggle(w http.ResponseWriter, req *http.Request) {
	id := req.FormValue("id")
	enable := req.FormValue("enable") == "true"
	if id == "" || id == "core" {
		http.Error(w, "webui: refusing to toggle core or an empty id", http.StatusBadRequest)
		return
	}
	if err := machinevars.SetLayer(a.Repo, id, enable); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, req, "/layers", http.StatusSeeOther)
}

// — Source edits (Phase 6 authoring) —

type layerOptionVM struct{ ID, Name string }

func (a *App) handleSourceEdits(w http.ResponseWriter, req *http.Request) {
	layers, err := manifest.LoadAll(a.Repo.RootDir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var options []layerOptionVM
	for _, l := range layers {
		options = append(options, layerOptionVM{ID: l.ID, Name: l.Name})
	}
	branch, _ := authoring.CurrentBranch(a.Repo.RootDir)
	dirty, _ := authoring.Dirty(a.Repo.RootDir)

	a.render(w, "sourceedits", map[string]any{
		"Title": "Source edits", "Kicker": "authoring",
		"Nav": navFor("Source edits"), "Machine": a.machine(),
		"Layers": options, "Branch": branch, "Dirty": dirty,
	})
}

func (a *App) handleMergeRoles(w http.ResponseWriter, req *http.Request) {
	from := req.FormValue("from")
	into := req.FormValue("into")

	result := map[string]any{
		"Title": "Merge result", "Kicker": "authoring",
		"Nav": navFor("Source edits"), "Machine": a.machine(),
		"From": from, "Into": into,
	}
	fail := func(stage, errMsg string) {
		result["Stage"] = stage
		result["Error"] = errMsg
		a.render(w, "sourceedits", result)
	}

	layers, err := manifest.LoadAll(a.Repo.RootDir)
	if err != nil {
		fail("load manifests", err.Error())
		return
	}
	if err := rolemerge.Check(a.Repo.RootDir, layers, from, into); err != nil {
		fail("guardrail check", err.Error())
		return
	}

	branch, err := authoring.NewBranch(a.Repo.RootDir, "merge-"+from+"-into-"+into)
	if err != nil {
		fail("create branch", err.Error())
		return
	}
	result["Branch"] = branch

	if err := rolemerge.Apply(a.Repo.RootDir, layers, from, into); err != nil {
		result["LeftOnBranch"] = true
		fail("apply merge", err.Error())
		return
	}

	sandboxResult, err := sandbox.Run(a.Repo.RootDir)
	result["SandboxOutput"] = sandboxResult.Output
	if err != nil {
		result["LeftOnBranch"] = true
		fail("sandbox test", err.Error())
		return
	}
	if !sandboxResult.OK {
		result["LeftOnBranch"] = true
		fail("sandbox test", "syntax check failed in the container — see output below")
		return
	}

	message := fmt.Sprintf("refactor: merge %s into %s\n\nGenerated by Converge's role-merge authoring flow.", from, into)
	if err := authoring.Commit(a.Repo.RootDir, message); err != nil {
		result["LeftOnBranch"] = true
		fail("commit", err.Error())
		return
	}
	if err := authoring.Push(a.Repo.RootDir, branch); err != nil {
		result["LeftOnBranch"] = true
		fail("push", err.Error())
		return
	}
	prURL, err := authoring.OpenPR(a.Repo.RootDir, branch,
		fmt.Sprintf("Merge %s into %s", from, into),
		fmt.Sprintf("Merges the `%s` role into `%s`, generated and sandbox-tested by Converge's authoring flow. Nothing has been applied to any machine — review the diff before merging.", from, into))
	if err != nil {
		result["LeftOnBranch"] = true
		fail("open PR", err.Error())
		return
	}
	result["PRURL"] = prURL

	if err := authoring.CheckoutMain(a.Repo.RootDir); err != nil {
		result["CheckoutMainError"] = err.Error()
	}

	a.render(w, "sourceedits", result)
}
