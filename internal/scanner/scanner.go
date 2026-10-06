// Package scanner walks media directories: full scan, incremental refresh,
// NFO parsing, local posters, and ffprobe media info extraction.
package scanner

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"emby-server/internal/db"
	"emby-server/internal/log"
	"emby-server/internal/nfo"
	"emby-server/internal/subs"
)

var videoExts = map[string]bool{
	".mp4": true, ".mkv": true, ".avi": true, ".mov": true,
	".wmv": true, ".flv": true, ".ts": true, ".m2ts": true,
}

// Scanner holds scan state.
type Scanner struct {
	db          *db.DB
	ffprobePath string
	mu          sync.Mutex
	running     bool
}

// New creates a scanner.
func New(d *db.DB, ffprobePath string) *Scanner {
	return &Scanner{db: d, ffprobePath: ffprobePath}
}

// Running reports whether a scan is in progress.
func (s *Scanner) Running() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// ScanAll performs a full scan of all libraries.
func (s *Scanner) ScanAll() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
	}()

	rows, err := s.db.Query(`SELECT id,name,path FROM libraries`)
	if err != nil {
		log.Error("scan: %v", err)
		return
	}
	type lib struct{ id, name, path string }
	var libs []lib
	for rows.Next() {
		var l lib
		rows.Scan(&l.id, &l.name, &l.path)
		libs = append(libs, l)
	}
	rows.Close()
	for _, l := range libs {
		log.Info("scan library %s (%s)", l.name, l.path)
		s.scanDir(l.id, l.path, true)
	}
	log.Info("full scan complete")
}

// Refresh performs incremental refresh: new/changed files only.
func (s *Scanner) Refresh() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		s.running = false
		s.mu.Unlock()
	}()

	rows, err := s.db.Query(`SELECT id,path FROM libraries`)
	if err != nil {
		return
	}
	type lib struct{ id, path string }
	var libs []lib
	for rows.Next() {
		var l lib
		rows.Scan(&l.id, &l.path)
		libs = append(libs, l)
	}
	rows.Close()
	for _, l := range libs {
		s.scanDir(l.id, l.path, false)
	}
	log.Info("incremental refresh complete")
}

func (s *Scanner) scanDir(libID, root string, full bool) {
	filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(p))
		if !videoExts[ext] {
			return nil
		}
		fi, _ := d.Info()
		s.indexFile(libID, p, fi.ModTime().Unix(), fi.Size(), full)
		return nil
	})
	// remove missing
	rows, _ := s.db.Query(`SELECT id,path FROM items WHERE lib=?`, libID)
	var gone []string
	for rows.Next() {
		var id, p string
		rows.Scan(&id, &p)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			gone = append(gone, id)
		}
	}
	rows.Close()
	for _, id := range gone {
		s.db.Exec(`DELETE FROM items WHERE id=?`, id)
		s.db.Exec(`DELETE FROM subtitles WHERE item_id=?`, id)
		log.Info("removed missing item %s", id)
	}
}

func (s *Scanner) indexFile(libID, path string, mtime, size int64, full bool) {
	var id string
	var oldMtime int64
	err := s.db.QueryRow(`SELECT id,mtime FROM items WHERE path=?`, path).Scan(&id, &oldMtime)
	if err == nil && !full && oldMtime >= mtime {
		return // unchanged, incremental skip
	}
	if err != nil {
		id = db.NewID()
	}

	dir := filepath.Dir(path)
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	name := prettify(base)

	it := itemData{ID: id, Lib: libID, Name: name, Kind: "Movie", Path: path,
		Mtime: mtime, Size: size}

	// NFO
	if m, err := nfo.ParseMovie(filepath.Join(dir, base+".nfo")); err == nil {
		if m.Title != "" {
			it.Name = m.Title
		}
		it.Overview = m.Plot
		it.Year = m.Year
		it.Rating = m.Rating
		it.TmdbID = m.TmdbID
	}
	// local poster
	for _, cand := range []string{
		filepath.Join(dir, base+"-poster.jpg"),
		filepath.Join(dir, base+".jpg"),
		filepath.Join(dir, "poster.jpg"),
		filepath.Join(dir, "folder.jpg"),
	} {
		if _, err := os.Stat(cand); err == nil {
			it.Poster = cand
			break
		}
	}
	// local backdrop
	if _, err := os.Stat(filepath.Join(dir, "backdrop.jpg")); err == nil {
		it.Backdrop = filepath.Join(dir, "backdrop.jpg")
	}

	// ffprobe media info
	if mi := s.probe(path); mi != nil {
		it.Width, it.Height = mi.Width, mi.Height
		it.VCodec, it.ACodec = mi.VCodec, mi.ACodec
		it.Duration = mi.Duration
		if mi.Duration > 0 {
			it.RuntimeTicks = int64(mi.Duration * 1e7)
		}
	}

	_, err = s.db.Exec(`INSERT INTO items(id,lib,parent,name,kind,path,overview,poster,backdrop,year,
		runtime_ticks,tmdb_id,rating,mtime,size,width,height,vcodec,acodec,duration)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
		ON CONFLICT(path) DO UPDATE SET name=excluded.name,kind=excluded.kind,overview=excluded.overview,
		poster=excluded.poster,backdrop=excluded.backdrop,year=excluded.year,runtime_ticks=excluded.runtime_ticks,
		tmdb_id=excluded.tmdb_id,rating=excluded.rating,mtime=excluded.mtime,size=excluded.size,
		width=excluded.width,height=excluded.height,vcodec=excluded.vcodec,acodec=excluded.acodec,
		duration=excluded.duration`,
		it.ID, it.Lib, "", it.Name, it.Kind, it.Path, it.Overview, it.Poster, it.Backdrop,
		it.Year, it.RuntimeTicks, it.TmdbID, it.Rating, it.Mtime, it.Size,
		it.Width, it.Height, it.VCodec, it.ACodec, it.Duration)
	if err != nil {
		log.Error("index %s: %v", path, err)
		return
	}

	// external subtitles
	for _, sp := range subs.FindExternal(path) {
		s.db.Exec(`INSERT OR IGNORE INTO subtitles(id,item_id,lang,path,external) VALUES(?,?,?,?,1)`,
			db.NewID(), id, subs.LangFromName(sp), sp)
	}
	log.Info("indexed %s", name)
}

type itemData struct {
	ID, Lib, Name, Kind, Path, Overview, Poster, Backdrop, TmdbID, VCodec, ACodec string
	Year                                                                          int
	Rating                                                                        float64
	Mtime, Size, RuntimeTicks                                                     int64
	Width, Height                                                                 int
	Duration                                                                      float64
}

func prettify(base string) string {
	// Movie.Name.2024.1080p -> Movie Name 2024
	s := strings.ReplaceAll(base, ".", " ")
	s = strings.ReplaceAll(s, "_", " ")
	return strings.TrimSpace(s)
}

// --- ffprobe ---

type mediaInfo struct {
	Width, Height  int
	VCodec, ACodec string
	Duration       float64
}

func (s *Scanner) probe(path string) *mediaInfo {
	if s.ffprobePath == "" {
		return nil
	}
	out, err := exec.Command(s.ffprobePath,
		"-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", path).Output()
	if err != nil {
		return nil
	}
	return parseProbe(out)
}
