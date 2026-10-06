// Package config handles server configuration.
package config

import (
	"flag"
	"os"
)

// Config holds server settings.
type Config struct {
	Addr          string
	DataDir       string
	MediaDirs     []string
	AdminPassword string
	TMDBApiKey    string
	FFprobePath   string
}

// Load parses flags and env.
func Load() *Config {
	c := &Config{}
	flag.StringVar(&c.Addr, "addr", ":8096", "listen address")
	flag.StringVar(&c.DataDir, "data", "./data", "data directory (sqlite db, cache)")
	flag.StringVar(&c.AdminPassword, "admin-password", "", "initial admin password")
	flag.StringVar(&c.TMDBApiKey, "tmdb-key", "", "TMDB API key for scraping")
	flag.StringVar(&c.FFprobePath, "ffprobe", "ffprobe", "ffprobe binary path")
	flag.Parse()

	if c.AdminPassword == "" {
		c.AdminPassword = os.Getenv("ADMIN_PASSWORD")
	}
	if c.TMDBApiKey == "" {
		c.TMDBApiKey = os.Getenv("TMDB_API_KEY")
	}
	if v := os.Getenv("MEDIA_DIRS"); v != "" {
		// comma-separated
		for _, d := range splitDirs(v) {
			c.MediaDirs = append(c.MediaDirs, d)
		}
	}
	return c
}

func splitDirs(s string) []string {
	var out []string
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == ',' {
			if d := s[start:i]; d != "" {
				out = append(out, d)
			}
			start = i + 1
		}
	}
	return out
}
