// Package tmdb is a minimal TMDB API client for metadata scraping.
package tmdb

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// Client talks to api.themoviedb.org.
type Client struct {
	key string
	hc  *http.Client
}

// New creates a client. Empty key disables scraping.
func New(key string) *Client {
	return &Client{key: key, hc: &http.Client{Timeout: 15 * time.Second}}
}

// Enabled reports whether a key is configured.
func (c *Client) Enabled() bool { return c.key != "" }

// Movie holds scraped metadata.
type Movie struct {
	TmdbID   int     `json:"id"`
	Title    string  `json:"title"`
	Overview string  `json:"overview"`
	Year     int     `json:"-"`
	Rating   float64 `json:"vote_average"`
	Poster   string  `json:"poster_path"`
	Backdrop string  `json:"backdrop_path"`
}

type searchResp struct {
	Results []Movie `json:"results"`
}

// SearchMovie finds candidates by title+year.
func (c *Client) SearchMovie(title string, year int) ([]Movie, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("tmdb key not configured")
	}
	u := "https://api.themoviedb.org/3/search/movie?api_key=" + c.key +
		"&query=" + url.QueryEscape(title) + "&language=zh-CN"
	if year > 0 {
		u += fmt.Sprintf("&year=%d", year)
	}
	resp, err := c.hc.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var sr searchResp
	if err := json.NewDecoder(resp.Body).Decode(&sr); err != nil {
		return nil, err
	}
	for i := range sr.Results {
		if len(sr.Results[i].Poster) > 0 && sr.Results[i].Poster[0] == '/' {
			sr.Results[i].Poster = "https://image.tmdb.org/t/p/w500" + sr.Results[i].Poster
		}
		if len(sr.Results[i].Backdrop) > 0 && sr.Results[i].Backdrop[0] == '/' {
			sr.Results[i].Backdrop = "https://image.tmdb.org/t/p/w1280" + sr.Results[i].Backdrop
		}
	}
	return sr.Results, nil
}

// ImageURL builds a full image URL.
func ImageURL(path, size string) string {
	if path == "" {
		return ""
	}
	if len(path) > 4 && path[:4] == "http" {
		return path
	}
	return "https://image.tmdb.org/t/p/" + size + path
}
