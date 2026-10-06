// Package subs handles external subtitle detection and WebVTT conversion.
package subs

import (
	"os"
	"path/filepath"
	"strings"
)

// IsSubtitle reports whether ext is a subtitle file.
func IsSubtitle(ext string) bool {
	switch strings.ToLower(ext) {
	case ".srt", ".ass", ".ssa", ".vtt", ".sub":
		return true
	}
	return false
}

// FindExternal returns subtitle files next to a media file.
func FindExternal(mediaPath string) []string {
	dir := filepath.Dir(mediaPath)
	base := strings.TrimSuffix(filepath.Base(mediaPath), filepath.Ext(mediaPath))
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		name := e.Name()
		ext := filepath.Ext(name)
		if !IsSubtitle(ext) {
			continue
		}
		stem := strings.TrimSuffix(name, ext)
		if stem == base || strings.HasPrefix(stem, base+".") || strings.HasPrefix(stem, base+"_") {
			out = append(out, filepath.Join(dir, name))
		}
	}
	return out
}

// LangFromName extracts a language hint from filename like movie.zh.srt.
func LangFromName(p string) string {
	stem := strings.TrimSuffix(filepath.Base(p), filepath.Ext(p))
	parts := strings.Split(stem, ".")
	if len(parts) >= 2 {
		return parts[len(parts)-1]
	}
	return ""
}
