// emby-server: Emby-compatible media server in Go.
// Direct play only — no transcoding.
package main

import (
	"net/http"
	"os"
	"path/filepath"

	"emby-server/internal/api"
	"emby-server/internal/config"
	"emby-server/internal/db"
	"emby-server/internal/log"
	"emby-server/internal/scanner"
	"emby-server/internal/tmdb"
)

func main() {
	cfg := config.Load()

	database, err := db.Open(cfg.DataDir)
	if err != nil {
		log.Error("open db: %v", err)
		os.Exit(1)
	}

	// bootstrap admin
	if database.UserCount() == 0 {
		pw := cfg.AdminPassword
		if pw == "" {
			pw = "admin123"
		}
		id, err := database.CreateUser("admin", pw, true)
		if err != nil {
			log.Error("create admin: %v", err)
			os.Exit(1)
		}
		log.Info("created admin user (id=%s)", id)
	}

	// libraries from env MEDIA_DIRS
	for _, d := range cfg.MediaDirs {
		database.Exec(`INSERT OR IGNORE INTO libraries(id,name,path,kind) VALUES(?,?,?,?)`,
			db.NewID(), filepath.Base(d), d, "movies")
	}

	// tmdb key: flag/env first, then settings table
	tmdbKey := cfg.TMDBApiKey
	if tmdbKey == "" {
		database.QueryRow(`SELECT v FROM settings WHERE k='tmdb_key'`).Scan(&tmdbKey)
	}
	tc := tmdb.New(tmdbKey)

	serverID := db.NewID()
	srv := api.New(database, serverID)
	sc := scanner.New(database, cfg.FFprobePath)
	admin := api.NewAdmin(srv, sc, tc, cfg.DataDir)
	_ = admin

	// web UI: serve ./web/dist if built, else a stub page
	mux := http.NewServeMux()
	webDir := "./web/dist"
	if st, err := os.Stat(webDir); err == nil && st.IsDir() {
		log.Info("serving web UI from %s", webDir)
		mux.Handle("/", http.FileServer(http.Dir(webDir)))
	} else {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(`<html><body><h1>emby-server</h1><p>Web UI not built yet. API at /emby/*</p></body></html>`))
		})
	}
	// API routes take precedence for /emby/* and /api/*
	mux.Handle("/emby/", srv)
	mux.Handle("/api/", srv)

	log.Info("emby-server listening on %s", cfg.Addr)
	if err := http.ListenAndServe(cfg.Addr, mux); err != nil {
		log.Error("serve: %v", err)
		os.Exit(1)
	}
}
