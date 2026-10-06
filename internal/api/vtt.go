package api

import (
	"regexp"
	"strings"
)

// toWebVTT converts SRT/ASS-ish content to WebVTT.
func toWebVTT(data, ext string) string {
	switch ext {
	case ".vtt":
		return data
	case ".srt":
		return srtToVTT(data)
	default:
		return srtToVTT(stripASS(data))
	}
}

var srtTime = regexp.MustCompile(`(\d{2}:\d{2}:\d{2}),(\d{3})`)

func srtToVTT(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = srtTime.ReplaceAllString(s, `$1.$2`)
	if !strings.HasPrefix(s, "WEBVTT") {
		s = "WEBVTT\n\n" + s
	}
	return s
}

var assTag = regexp.MustCompile(`\{[^}]*\}`)

func stripASS(s string) string {
	lines := strings.Split(s, "\n")
	var out []string
	for _, ln := range lines {
		ln = strings.TrimSpace(ln)
		if strings.HasPrefix(ln, "Dialogue:") {
			parts := strings.SplitN(ln, ",", 10)
			if len(parts) == 10 {
				text := assTag.ReplaceAllString(parts[9], "")
				text = strings.ReplaceAll(text, `\N`, "\n")
				start := strings.Replace(parts[1], ".", ",", -1)
				end := strings.Replace(parts[2], ".", ",", -1)
				out = append(out, start+" --> "+end+"\n"+text+"\n")
			}
		}
	}
	return strings.Join(out, "\n")
}
