package api

import (
	"net/http"
	"strconv"
)

// GET /emby/System/Info/Public
func (s *Server) publicInfo(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{
		"ServerName": "emby-server",
		"Version":    "1.0.0",
		"Id":         s.serverID,
	})
}

// GET /emby/Users (public list for login screen)
func (s *Server) usersPublic(w http.ResponseWriter, r *http.Request) {
	rows, err := s.db.Query(`SELECT id,name,admin FROM users ORDER BY name`)
	if err != nil {
		fail(w, 500, "db error")
		return
	}
	defer rows.Close()
	out := []any{}
	for rows.Next() {
		var id, name string
		var admin int
		rows.Scan(&id, &name, &admin)
		out = append(out, map[string]any{
			"Id": id, "Name": name, "ServerId": s.serverID,
			"HasPassword": true, "HasConfiguredPassword": true,
			"Policy": map[string]any{"IsAdministrator": admin == 1},
		})
	}
	writeJSON(w, 200, out)
}

// POST /emby/Users/AuthenticateByName
func (s *Server) authenticate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		fail(w, 405, "POST required")
		return
	}
	var body struct {
		Username string `json:"Username"`
		Pw       string `json:"Pw"`
	}
	if err := decodeJSON(r, &body); err != nil || body.Username == "" {
		fail(w, 400, "Username required")
		return
	}
	uid, ok := s.db.VerifyUser(body.Username, body.Pw)
	if !ok {
		fail(w, 401, "Invalid username or password")
		return
	}
	maxDev := s.db.MaxDevices(uid)
	device := r.Header.Get("X-Emby-Device-Id")
	if device == "" {
		device = r.Header.Get("X-Emby-Device-Name")
	}
	tok, err := s.db.IssueToken(uid, device, maxDev)
	if err != nil {
		fail(w, 500, "token error")
		return
	}
	var name string
	var admin int
	s.db.QueryRow(`SELECT name,admin FROM users WHERE id=?`, uid).Scan(&name, &admin)
	writeJSON(w, 200, map[string]any{
		"User": map[string]any{
			"Id": name, "Name": name, "ServerId": s.serverID,
			"HasPassword": true, "HasConfiguredPassword": true,
			"Policy": map[string]any{"IsAdministrator": admin == 1},
		},
		"AccessToken": tok,
		"ServerId":    s.serverID,
	})
}

// GET /emby/System/Info
func (s *Server) systemInfo(w http.ResponseWriter, r *http.Request, u authUser) {
	writeJSON(w, 200, map[string]any{
		"ServerName": "emby-server", "Version": "1.0.0", "Id": s.serverID,
		"OperatingSystem": "Linux", "CanSelfRestart": false,
	})
}

// GET /emby/Users/{uid}/Views
func (s *Server) views(w http.ResponseWriter, r *http.Request, u authUser) {
	rows, err := s.db.Query(`SELECT id,name,kind FROM libraries ORDER BY name`)
	if err != nil {
		fail(w, 500, "db error")
		return
	}
	defer rows.Close()
	items := []any{}
	for rows.Next() {
		var id, name, kind string
		rows.Scan(&id, &name, &kind)
		items = append(items, map[string]any{
			"Id": id, "Name": name, "Type": "CollectionFolder",
			"CollectionType": kind,
		})
	}
	writeJSON(w, 200, map[string]any{"Items": items, "TotalRecordCount": len(items)})
}

// GET /emby/Users/{uid}/Items
func (s *Server) items(w http.ResponseWriter, r *http.Request, u authUser) {
	q := r.URL.Query()
	parent := q.Get("ParentId")
	limit, _ := strconv.Atoi(q.Get("Limit"))
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	start, _ := strconv.Atoi(q.Get("StartIndex"))

	var rows *sqlRows
	var err error
	if parent != "" {
		rows, err = s.db.Query(`SELECT id,name,kind,overview,poster,year,runtime_ticks,tmdb_id,rating,width,height
			FROM items WHERE parent=? ORDER BY name LIMIT ? OFFSET ?`, parent, limit, start)
	} else {
		rows, err = s.db.Query(`SELECT id,name,kind,overview,poster,year,runtime_ticks,tmdb_id,rating,width,height
			FROM items WHERE parent='' ORDER BY name LIMIT ? OFFSET ?`, limit, start)
	}
	if err != nil {
		fail(w, 500, "db error")
		return
	}
	defer rows.Close()
	items := s.scanItems(rows)
	writeJSON(w, 200, map[string]any{"Items": items, "TotalRecordCount": len(items)})
}

// GET /emby/Items/{id}
func (s *Server) itemDetail(w http.ResponseWriter, r *http.Request, u authUser) {
	id := r.PathValue("id")
	var it itemRow
	err := s.db.QueryRow(`SELECT id,name,kind,overview,poster,backdrop,year,runtime_ticks,tmdb_id,rating,
		width,height,vcodec,acodec,duration,path FROM items WHERE id=?`, id).
		Scan(&it.ID, &it.Name, &it.Kind, &it.Overview, &it.Poster, &it.Backdrop, &it.Year,
			&it.RuntimeTicks, &it.TmdbID, &it.Rating, &it.Width, &it.Height,
			&it.VCodec, &it.ACodec, &it.Duration, &it.Path)
	if err != nil {
		fail(w, 404, "not found")
		return
	}
	// playstate
	var pos int64
	var played int
	s.db.QueryRow(`SELECT position,played FROM plays WHERE user_id=? AND item=?`, u.ID, id).Scan(&pos, &played)
	writeJSON(w, 200, it.toMap(s.serverID, pos, played == 1))
}
