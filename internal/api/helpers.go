package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

type sqlRows = sql.Rows

func decodeJSON(r *http.Request, v any) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}

type itemRow struct {
	ID, Name, Kind, Overview, Poster, Backdrop, Path, TmdbID, VCodec, ACodec string
	Year                                                                       int
	RuntimeTicks                                                               int64
	Rating                                                                     float64
	Width, Height                                                              int
	Duration                                                                   float64
}

func (s *Server) scanItems(rows *sqlRows) []any {
	out := []any{}
	for rows.Next() {
		var it itemRow
		rows.Scan(&it.ID, &it.Name, &it.Kind, &it.Overview, &it.Poster,
			&it.Year, &it.RuntimeTicks, &it.TmdbID, &it.Rating, &it.Width, &it.Height)
		out = append(out, it.toMap(s.serverID, 0, false))
	}
	return out
}

func (it itemRow) toMap(serverID string, posTicks int64, played bool) map[string]any {
	m := map[string]any{
		"Id": it.ID, "Name": it.Name, "Type": it.Kind,
		"Overview": it.Overview, "PremiereDate": "",
		"ProductionYear": it.Year, "RunTimeTicks": it.RuntimeTicks,
		"CommunityRating": it.Rating, "ServerId": serverID,
		"Path": it.Path,
		"MediaSources": []any{
			map[string]any{
				"Id": it.ID, "Name": it.Name, "Path": it.Path,
				"Protocol": "File", "SupportsDirectPlay": true,
				"SupportsDirectStream": true, "SupportsTranscoding": false,
				"MediaStreams": []any{
					map[string]any{"Type": "Video", "Codec": it.VCodec, "Width": it.Width, "Height": it.Height},
					map[string]any{"Type": "Audio", "Codec": it.ACodec},
				},
			},
		},
		"UserData": map[string]any{
			"PlaybackPositionTicks": posTicks, "Played": played,
		},
	}
	if it.Poster != "" {
		m["ImageTags"] = map[string]string{"Primary": "1"}
	}
	return m
}
