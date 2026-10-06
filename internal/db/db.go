// Package db provides SQLite persistence.
package db

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// DB wraps sql.DB.
type DB struct {
	*sql.DB
}

// Open opens (creating if needed) the SQLite database.
func Open(dataDir string) (*DB, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}
	dsn := filepath.Join(dataDir, "emby.db")
	sqlDB, err := sql.Open("sqlite", dsn+"?cache=shared")
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(1)
	db := &DB{sqlDB}
	if err := db.migrate(); err != nil {
		return nil, err
	}
	return db, nil
}

func (d *DB) migrate() error {
	schema := `
CREATE TABLE IF NOT EXISTS users(
	id TEXT PRIMARY KEY,
	name TEXT UNIQUE NOT NULL,
	hash TEXT NOT NULL,
	admin INTEGER NOT NULL DEFAULT 0,
	max_devices INTEGER NOT NULL DEFAULT 2
);
CREATE TABLE IF NOT EXISTS tokens(
	hash TEXT PRIMARY KEY,
	user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	device TEXT NOT NULL DEFAULT '',
	expires INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS libraries(
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	path TEXT UNIQUE NOT NULL,
	kind TEXT NOT NULL DEFAULT 'movies'
);
CREATE TABLE IF NOT EXISTS items(
	id TEXT PRIMARY KEY,
	lib TEXT NOT NULL REFERENCES libraries(id) ON DELETE CASCADE,
	parent TEXT NOT NULL DEFAULT '',
	name TEXT NOT NULL,
	kind TEXT NOT NULL,
	path TEXT UNIQUE NOT NULL,
	overview TEXT NOT NULL DEFAULT '',
	poster TEXT NOT NULL DEFAULT '',
	backdrop TEXT NOT NULL DEFAULT '',
	year INTEGER NOT NULL DEFAULT 0,
	runtime_ticks INTEGER NOT NULL DEFAULT 0,
	tmdb_id TEXT NOT NULL DEFAULT '',
	rating REAL NOT NULL DEFAULT 0,
	mtime INTEGER NOT NULL DEFAULT 0,
	size INTEGER NOT NULL DEFAULT 0,
	-- media info (ffprobe)
	width INTEGER NOT NULL DEFAULT 0,
	height INTEGER NOT NULL DEFAULT 0,
	vcodec TEXT NOT NULL DEFAULT '',
	acodec TEXT NOT NULL DEFAULT '',
	duration REAL NOT NULL DEFAULT 0
);
CREATE TABLE IF NOT EXISTS subtitles(
	id TEXT PRIMARY KEY,
	item_id TEXT NOT NULL REFERENCES items(id) ON DELETE CASCADE,
	lang TEXT NOT NULL DEFAULT '',
	path TEXT NOT NULL,
	external INTEGER NOT NULL DEFAULT 1
);
CREATE TABLE IF NOT EXISTS plays(
	user_id TEXT NOT NULL,
	item TEXT NOT NULL,
	position INTEGER NOT NULL DEFAULT 0,
	played INTEGER NOT NULL DEFAULT 0,
	updated INTEGER NOT NULL DEFAULT 0,
	PRIMARY KEY(user_id, item)
);
CREATE TABLE IF NOT EXISTS settings(
	k TEXT PRIMARY KEY,
	v TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_items_lib ON items(lib);
CREATE INDEX IF NOT EXISTS idx_items_parent ON items(parent);
`
	_, err := d.Exec(schema)
	return err
}
