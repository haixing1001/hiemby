package api

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"emby-server/internal/db"
	"emby-server/internal/log"
	"emby-server/internal/scanner"
	"emby-server/internal/tmdb"
)

// Admin wires admin endpoints.
type Admin struct {
	srv     *Server
	scanner *scanner.Scanner
	tmdb    *tmdb.Client
	dataDir string
}

// NewAdmin creates the admin handler set and mounts routes.
func NewAdmin(srv *Server, sc *scanner.Scanner, tc *tmdb.Client, dataDir string) *Admin {
	a := &Admin{srv: srv, scanner: sc, tmdb: tc, dataDir: dataDir}
	a.Mount()
	return a
}

// SetTMDB swaps the TMDB client (after settings change).
func (a *Admin) SetTMDB(tc *tmdb.Client) { a.tmdb = tc }

// Mount registers admin routes on the server mux.
func (a *Admin) Mount() {
	m := a.srv.mux
	admin := a.srv.requireAdmin

	// libraries
	m.HandleFunc("/api/admin/libraries", admin(a.libraries))
	m.HandleFunc("/api/admin/libraries/{id}", admin(a.libraryDetail))
	// scan
	m.HandleFunc("/api/admin/scan", admin(a.scan))
	m.HandleFunc("/api/admin/scan/status", admin(a.scanStatus))
	// users
	m.HandleFunc("/api/admin/users", admin(a.users))
	m.HandleFunc("/api/admin/users/{id}", admin(a.userDetail))
	// files
	m.HandleFunc("/api/admin/files", admin(a.files))
	// settings (tmdb key etc.)
	m.HandleFunc("/api/admin/settings", admin(a.settings))
	// tmdb scrape
	m.HandleFunc("/api/admin/scrape", admin(a.scrape))
	// realtime logs (SSE)
	m.HandleFunc("/api/admin/logs", admin(a.logsSSE))
}

// GET/POST /api/admin/libraries
func (a *Admin) libraries(w http.ResponseWriter, r *http.Request, u authUser) {
	switch r.Method {
	case http.MethodGet:
		rows, _ := a.srv.db.Query(`SELECT id,name,path,kind FROM libraries ORDER BY name`)
		type lib struct {
			id, name, path, kind string
		}
		var libs []lib
		for rows.Next() {
			var l lib
			rows.Scan(&l.id, &l.name, &l.path, &l.kind)
			libs = append(libs, l)
		}
		rows.Close()
		out := []any{}
		for _, l := range libs {
			var cnt int
			a.srv.db.QueryRow(`SELECT COUNT(*) FROM items WHERE lib=?`, l.id).Scan(&cnt)
			out = append(out, map[string]any{"Id": l.id, "Name": l.name, "Path": l.path, "Kind": l.kind, "Count": cnt})
		}
		writeJSON(w, 200, out)
	case http.MethodPost:
		var body struct {
			Name string `json:"Name"`
			Path string `json:"Path"`
			Kind string `json:"Kind"`
		}
		if err := decodeJSON(r, &body); err != nil || body.Path == "" {
			fail(w, 400, "Path required")
			return
		}
		if body.Name == "" {
			body.Name = filepath.Base(body.Path)
		}
		if body.Kind == "" {
			body.Kind = "movies"
		}
		id := db.NewID()
		if _, err := a.srv.db.Exec(`INSERT INTO libraries(id,name,path,kind) VALUES(?,?,?,?)`,
			id, body.Name, body.Path, body.Kind); err != nil {
			fail(w, 400, "library exists or db error")
			return
		}
		log.Info("library added: %s", body.Path)
		writeJSON(w, 200, map[string]string{"Id": id})
	default:
		fail(w, 405, "method not allowed")
	}
}

// DELETE /api/admin/libraries/{id}
func (a *Admin) libraryDetail(w http.ResponseWriter, r *http.Request, u authUser) {
	id := r.PathValue("id")
	if r.Method != http.MethodDelete {
		fail(w, 405, "method not allowed")
		return
	}
	a.srv.db.Exec(`DELETE FROM libraries WHERE id=?`, id)
	writeJSON(w, 200, map[string]bool{"ok": true})
}

// POST /api/admin/scan?mode=full|incremental
func (a *Admin) scan(w http.ResponseWriter, r *http.Request, u authUser) {
	if r.Method != http.MethodPost {
		fail(w, 405, "POST required")
		return
	}
	if a.scanner.Running() {
		fail(w, 409, "scan already running")
		return
	}
	mode := r.URL.Query().Get("mode")
	if mode == "incremental" {
		go a.scanner.Refresh()
	} else {
		go a.scanner.ScanAll()
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func (a *Admin) scanStatus(w http.ResponseWriter, r *http.Request, u authUser) {
	writeJSON(w, 200, map[string]bool{"running": a.scanner.Running()})
}

// GET/POST /api/admin/users
func (a *Admin) users(w http.ResponseWriter, r *http.Request, u authUser) {
	switch r.Method {
	case http.MethodGet:
		rows, _ := a.srv.db.Query(`SELECT id,name,admin,max_devices FROM users ORDER BY name`)
		type usr struct {
			id, name string
			admin, maxDev int
		}
		var usrs []usr
		for rows.Next() {
			var u2 usr
			rows.Scan(&u2.id, &u2.name, &u2.admin, &u2.maxDev)
			usrs = append(usrs, u2)
		}
		rows.Close()
		out := []any{}
		for _, u2 := range usrs {
			var devs int
			a.srv.db.QueryRow(`SELECT COUNT(*) FROM tokens WHERE user_id=?`, u2.id).Scan(&devs)
			out = append(out, map[string]any{
				"Id": u2.id, "Name": u2.name, "IsAdmin": u2.admin == 1,
				"MaxDevices": u2.maxDev, "Devices": devs,
			})
		}
		writeJSON(w, 200, out)
	case http.MethodPost:
		var body struct {
			Name       string `json:"Name"`
			Password   string `json:"Password"`
			IsAdmin    bool   `json:"IsAdmin"`
			MaxDevices int    `json:"MaxDevices"`
		}
		if err := decodeJSON(r, &body); err != nil || body.Name == "" {
			fail(w, 400, "Name required")
			return
		}
		id, err := a.srv.db.CreateUser(body.Name, body.Password, body.IsAdmin)
		if err != nil {
			fail(w, 400, err.Error())
			return
		}
		if body.MaxDevices > 0 {
			a.srv.db.Exec(`UPDATE users SET max_devices=? WHERE id=?`, body.MaxDevices, id)
		}
		writeJSON(w, 200, map[string]string{"Id": id})
	}
}

// PUT/DELETE /api/admin/users/{id}
func (a *Admin) userDetail(w http.ResponseWriter, r *http.Request, u authUser) {
	id := r.PathValue("id")
	switch r.Method {
	case http.MethodPut:
		var body struct {
			Password   string `json:"Password"`
			MaxDevices *int   `json:"MaxDevices"`
		}
		decodeJSON(r, &body)
		if body.Password != "" {
			if err := a.srv.db.SetPassword(id, body.Password); err != nil {
				fail(w, 400, err.Error())
				return
			}
		}
		if body.MaxDevices != nil {
			a.srv.db.Exec(`UPDATE users SET max_devices=? WHERE id=?`, *body.MaxDevices, id)
		}
		writeJSON(w, 200, map[string]bool{"ok": true})
	case http.MethodDelete:
		if id == u.ID {
			fail(w, 400, "cannot delete self")
			return
		}
		a.srv.db.Exec(`DELETE FROM users WHERE id=?`, id)
		writeJSON(w, 200, map[string]bool{"ok": true})
	default:
		fail(w, 405, "method not allowed")
	}
}

// GET /api/admin/files?path=... — file manager
func (a *Admin) files(w http.ResponseWriter, r *http.Request, u authUser) {
	if r.Method == http.MethodDelete {
		p := r.URL.Query().Get("path")
		if p == "" {
			fail(w, 400, "path required")
			return
		}
		if err := os.Remove(p); err != nil {
			fail(w, 400, err.Error())
			return
		}
		writeJSON(w, 200, map[string]bool{"ok": true})
		return
	}
	dir := r.URL.Query().Get("path")
	if dir == "" {
		dir = "/"
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		fail(w, 400, err.Error())
		return
	}
	out := []any{}
	for _, e := range entries {
		fi, _ := e.Info()
		out = append(out, map[string]any{
			"Name": e.Name(), "Path": filepath.Join(dir, e.Name()),
			"IsDir": e.IsDir(), "Size": fi.Size(),
			"ModTime": fi.ModTime().Unix(),
		})
	}
	writeJSON(w, 200, map[string]any{"Path": dir, "Items": out})
}

// GET/POST /api/admin/settings
func (a *Admin) settings(w http.ResponseWriter, r *http.Request, u authUser) {
	switch r.Method {
	case http.MethodGet:
		rows, _ := a.srv.db.Query(`SELECT k,v FROM settings`)
		defer rows.Close()
		out := map[string]string{}
		for rows.Next() {
			var k, v string
			rows.Scan(&k, &v)
			if k == "tmdb_key" && v != "" {
				v = "***"
			}
			out[k] = v
		}
		writeJSON(w, 200, out)
	case http.MethodPost:
		var body map[string]string
		if err := decodeJSON(r, &body); err != nil {
			fail(w, 400, "bad json")
			return
		}
		for k, v := range body {
			if v == "***" {
				continue
			}
			a.srv.db.Exec(`INSERT INTO settings(k,v) VALUES(?,?)
				ON CONFLICT(k) DO UPDATE SET v=excluded.v`, k, v)
			if k == "tmdb_key" {
				a.tmdb = tmdb.New(v)
			}
		}
		writeJSON(w, 200, map[string]bool{"ok": true})
	}
}

// POST /api/admin/scrape — scrape metadata for items missing it
func (a *Admin) scrape(w http.ResponseWriter, r *http.Request, u authUser) {
	if r.Method != http.MethodPost {
		fail(w, 405, "POST required")
		return
	}
	if !a.tmdb.Enabled() {
		fail(w, 400, "tmdb key not configured")
		return
	}
	go a.runScrape()
	writeJSON(w, 200, map[string]bool{"ok": true})
}

func (a *Admin) runScrape() {
	rows, err := a.srv.db.Query(`SELECT id,name,year FROM items WHERE (tmdb_id='' OR overview='') AND kind='Movie' LIMIT 200`)
	if err != nil {
		return
	}
	defer rows.Close()
	type job struct{ id, name string; year int }
	var jobs []job
	for rows.Next() {
		var j job
		rows.Scan(&j.id, &j.name, &j.year)
		jobs = append(jobs, j)
	}
	for _, j := range jobs {
		res, err := a.tmdb.SearchMovie(j.name, j.year)
		if err != nil || len(res) == 0 {
			continue
		}
		m := res[0]
		yr := j.year
		a.srv.db.Exec(`UPDATE items SET tmdb_id=?,overview=?,rating=? WHERE id=?`,
			strconv.Itoa(m.TmdbID), m.Overview, m.Rating, j.id)
		log.Info("scraped %s -> tmdb %d (year %d)", j.name, m.TmdbID, yr)
	}
	log.Info("scrape complete (%d items)", len(jobs))
}

// GET /api/admin/logs — Server-Sent Events realtime log stream
func (a *Admin) logsSSE(w http.ResponseWriter, r *http.Request, u authUser) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	ch, unsub := log.Subscribe()
	defer unsub()
	fl, ok := w.(http.Flusher)
	if !ok {
		return
	}
	notify := r.Context().Done()
	for {
		select {
		case <-notify:
			return
		case e := <-ch:
			io.WriteString(w, "data: "+e.Time+" ["+e.Level+"] "+e.Message+"\n\n")
			fl.Flush()
		}
	}
}
