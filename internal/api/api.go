// Package api implements the Emby-compatible HTTP API.
package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"emby-server/internal/db"
	"emby-server/internal/log"
)

// Server holds dependencies.
type Server struct {
	db       *db.DB
	serverID string
	mux      *http.ServeMux
}

// New creates the API server.
func New(database *db.DB, serverID string) *Server {
	s := &Server{db: database, serverID: serverID, mux: http.NewServeMux()}
	s.routes()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// CORS for web UI and clients
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Emby-Token, X-MediaBrowser-Token, X-Emby-Authorization, Authorization")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	if r.Method == http.MethodOptions {
		w.WriteHeader(200)
		return
	}
	s.mux.ServeHTTP(w, r)
}

func (s *Server) routes() {
	// Public
	s.mux.HandleFunc("/emby/System/Info/Public", s.publicInfo)
	s.mux.HandleFunc("/emby/Users", s.usersPublic)
	s.mux.HandleFunc("/emby/Users/AuthenticateByName", s.authenticate)

	// Authenticated Emby API
	s.mux.HandleFunc("/emby/System/Info", s.requireAuth(s.systemInfo))
	s.mux.HandleFunc("/emby/Users/{uid}/Views", s.requireAuth(s.views))
	s.mux.HandleFunc("/emby/Users/{uid}/Items", s.requireAuth(s.items))
	s.mux.HandleFunc("/emby/Users/{uid}/Items/{id}", s.requireAuth(s.itemDetail))
	s.mux.HandleFunc("/emby/Items/{id}", s.requireAuth(s.itemDetail))

	// Playback (direct/redirect, no transcoding)
	s.mux.HandleFunc("/emby/Videos/{id}/stream", s.requireAuth(s.streamVideo))
	s.mux.HandleFunc("/emby/Audio/{id}/stream", s.requireAuth(s.streamAudio))
	s.mux.HandleFunc("/emby/Videos/{id}/Subtitles/{idx}/Stream", s.requireAuth(s.subtitleStream))

	// Playstate
	s.mux.HandleFunc("/emby/Sessions/Playing", s.requireAuth(s.playingProgress))
	s.mux.HandleFunc("/emby/Sessions/Playing/Stopped", s.requireAuth(s.playingStopped))

	// Images
	s.mux.HandleFunc("/emby/Items/{id}/Images/{kind}", s.requireAuth(s.itemImage))
}

// --- helpers ---

type ctxKey string

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func fail(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func tokenFrom(r *http.Request) string {
	for _, h := range []string{"X-Emby-Token", "X-MediaBrowser-Token"} {
		if t := r.Header.Get(h); t != "" {
			return t
		}
	}
	if t := r.URL.Query().Get("api_key"); t != "" {
		return t
	}
	if h := r.Header.Get("X-Emby-Authorization"); h != "" {
		// MediaEmby Client="...", Token="abc", ...
		if i := strings.Index(h, `Token="`); i >= 0 {
			rest := h[i+7:]
			if j := strings.Index(rest, `"`); j >= 0 {
				return rest[:j]
			}
		}
	}
	if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	return ""
}

type authUser struct {
	ID    string
	Name  string
	Admin bool
}

func (s *Server) requireAuth(next func(http.ResponseWriter, *http.Request, authUser)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t := tokenFrom(r)
		if t == "" {
			fail(w, 401, "missing token")
			return
		}
		id, name, admin, ok := s.db.UserByToken(t)
		if !ok {
			fail(w, 401, "invalid token")
			return
		}
		next(w, r, authUser{ID: id, Name: name, Admin: admin})
	}
}

func (s *Server) requireAdmin(next func(http.ResponseWriter, *http.Request, authUser)) http.HandlerFunc {
	return s.requireAuth(func(w http.ResponseWriter, r *http.Request, u authUser) {
		if !u.Admin {
			fail(w, 403, "admin required")
			return
		}
		next(w, r, u)
	})
}

func logReq(r *http.Request, u authUser) {
	log.Info("%s %s user=%s", r.Method, r.URL.Path, u.Name)
}
