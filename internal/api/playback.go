package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"emby-server/internal/log"
)

// Direct play / redirect. No transcoding.
func (s *Server) streamVideo(w http.ResponseWriter, r *http.Request, u authUser) {
	s.serveMedia(w, r, u)
}

func (s *Server) streamAudio(w http.ResponseWriter, r *http.Request, u authUser) {
	s.serveMedia(w, r, u)
}

func (s *Server) serveMedia(w http.ResponseWriter, r *http.Request, u authUser) {
	id := r.PathValue("id")
	var path, streamURL string
	err := s.db.QueryRow(`SELECT path,stream_url FROM items WHERE id=?`, id).Scan(&path, &streamURL)
	if err != nil {
		fail(w, 404, "not found")
		return
	}
	// .strm: redirect straight to the target URL (direct play, no proxy)
	if streamURL != "" {
		log.Info("strm redirect %s -> %s", u.Name, streamURL)
		http.Redirect(w, r, streamURL, http.StatusFound)
		return
	}
	// ?redirect=1 → 302 to a direct file URL (for reverse-proxy setups)
	if r.URL.Query().Get("redirect") == "1" {
		base := r.URL.Query().Get("base")
		http.Redirect(w, r, base+"/emby/Videos/"+id+"/stream?api_key="+tokenFrom(r), http.StatusFound)
		return
	}
	f, err := os.Open(path)
	if err != nil {
		fail(w, 404, "file missing")
		return
	}
	defer f.Close()
	fi, _ := f.Stat()
	ext := strings.ToLower(filepath.Ext(path))
	ct := "video/mp4"
	switch ext {
	case ".mkv":
		ct = "video/x-matroska"
	case ".avi":
		ct = "video/x-msvideo"
	case ".mp3":
		ct = "audio/mpeg"
	case ".flac":
		ct = "audio/flac"
	}
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Accept-Ranges", "bytes")
	log.Info("stream %s -> %s", u.Name, filepath.Base(path))
	http.ServeContent(w, r, fi.Name(), fi.ModTime(), f)
}

// GET /emby/Videos/{id}/Subtitles/{idx}/Stream
func (s *Server) subtitleStream(w http.ResponseWriter, r *http.Request, u authUser) {
	id := r.PathValue("id")
	idx, _ := strconv.Atoi(r.PathValue("idx"))
	var path string
	err := s.db.QueryRow(`SELECT path FROM subtitles WHERE item_id=? ORDER BY lang LIMIT 1 OFFSET ?`, id, idx).Scan(&path)
	if err != nil {
		fail(w, 404, "subtitle not found")
		return
	}
	// Convert to WebVTT for client compatibility
	data, err := os.ReadFile(path)
	if err != nil {
		fail(w, 404, "subtitle missing")
		return
	}
	vtt := toWebVTT(string(data), filepath.Ext(path))
	w.Header().Set("Content-Type", "text/vtt; charset=utf-8")
	w.Write([]byte(vtt))
}

// POST /emby/Sessions/Playing
func (s *Server) playingProgress(w http.ResponseWriter, r *http.Request, u authUser) {
	var body struct {
		ItemId       string `json:"ItemId"`
		PositionTicks int64 `json:"PositionTicks"`
	}
	decodeJSON(r, &body)
	if body.ItemId != "" {
		s.db.Exec(`INSERT INTO plays(user_id,item,position,updated) VALUES(?,?,?,?)
			ON CONFLICT(user_id,item) DO UPDATE SET position=excluded.position,updated=excluded.updated`,
			u.ID, body.ItemId, body.PositionTicks, time.Now().Unix())
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}

// POST /emby/Sessions/Playing/Stopped
func (s *Server) playingStopped(w http.ResponseWriter, r *http.Request, u authUser) {
	var body struct {
		ItemId        string `json:"ItemId"`
		PositionTicks int64  `json:"PositionTicks"`
		Failed        bool   `json:"Failed"`
	}
	decodeJSON(r, &body)
	if body.ItemId != "" {
		played := 0
		// mark played if >90% watched
		var dur int64
		s.db.QueryRow(`SELECT runtime_ticks FROM items WHERE id=?`, body.ItemId).Scan(&dur)
		if dur > 0 && float64(body.PositionTicks)/float64(dur) > 0.9 {
			played = 1
		}
		s.db.Exec(`INSERT INTO plays(user_id,item,position,played,updated) VALUES(?,?,?,?,?)
			ON CONFLICT(user_id,item) DO UPDATE SET position=excluded.position,played=excluded.played,updated=excluded.updated`,
			u.ID, body.ItemId, body.PositionTicks, played, time.Now().Unix())
	}
	writeJSON(w, 200, map[string]bool{"ok": true})
}

// GET /emby/Items/{id}/Images/{kind}
func (s *Server) itemImage(w http.ResponseWriter, r *http.Request, u authUser) {
	id := r.PathValue("id")
	kind := r.PathValue("kind")
	var poster, backdrop string
	err := s.db.QueryRow(`SELECT poster,backdrop FROM items WHERE id=?`, id).Scan(&poster, &backdrop)
	if err != nil {
		fail(w, 404, "not found")
		return
	}
	p := poster
	if kind == "Backdrop" {
		p = backdrop
	}
	if p == "" || !fileExists(p) {
		fail(w, 404, "no image")
		return
	}
	http.ServeFile(w, r, p)
}

func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
