// Package nfo parses Kodi-style NFO metadata files.
package nfo

import (
	"encoding/xml"
	"os"
	"strings"
)

// Movie is Kodi movie.nfo structure (subset).
type Movie struct {
	XMLName       xml.Name `xml:"movie"`
	Title         string   `xml:"title"`
	OriginalTitle string   `xml:"originaltitle"`
	Plot          string   `xml:"plot"`
	Year          int      `xml:"year"`
	Rating        float64  `xml:"rating"`
	TmdbID        string   `xml:"tmdbid"`
	Poster        string   `xml:"thumb"`
}

// ParseMovie reads a movie NFO file.
func ParseMovie(path string) (*Movie, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m Movie
	if err := xml.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	m.Title = strings.TrimSpace(m.Title)
	m.Plot = strings.TrimSpace(m.Plot)
	return &m, nil
}

// ParseTVShow reads tvshow.nfo (same subset).
func ParseTVShow(path string) (*Movie, error) { return ParseMovie(path) }
