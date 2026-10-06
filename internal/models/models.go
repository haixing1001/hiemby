// Package models defines core data structures.
package models

// User is an Emby user.
type User struct {
	ID                  string     `json:"Id"`
	Name                string     `json:"Name"`
	ServerID            string     `json:"ServerId"`
	HasPassword         bool       `json:"HasPassword"`
	HasConfiguredPassword bool     `json:"HasConfiguredPassword"`
	Policy              UserPolicy `json:"Policy"`
}

// UserPolicy holds permissions.
type UserPolicy struct {
	IsAdministrator bool `json:"IsAdministrator"`
	IsDisabled      bool `json:"IsDisabled"`
}

// Library is a media library.
type Library struct {
	ID   string `json:"Id"`
	Name string `json:"Name"`
	Path string `json:"Path"`
	Kind string `json:"Kind"` // movies, tvshows, music
}

// Item is a media item.
type Item struct {
	ID          string  `json:"Id"`
	Lib         string  `json:"-"`
	Parent      string  `json:"-"`
	Name        string  `json:"Name"`
	Kind        string  `json:"Kind"` // Movie, Episode, Season, Series, Folder
	Path        string  `json:"Path"`
	Overview    string  `json:"Overview"`
	Poster      string  `json:"Poster"`
	Backdrop    string  `json:"Backdrop"`
	Year        int     `json:"Year"`
	RuntimeTicks int64  `json:"RunTimeTicks"`
	TmdbID      string  `json:"TmdbId"`
	Rating      float64 `json:"Rating"`
	Width       int     `json:"Width"`
	Height      int     `json:"Height"`
	VCodec      string  `json:"VCodec"`
	ACodec      string  `json:"ACodec"`
	Duration    float64 `json:"Duration"`
}

// Subtitle is an external subtitle.
type Subtitle struct {
	ID     string `json:"Id"`
	ItemID string `json:"ItemId"`
	Lang   string `json:"Lang"`
	Path   string `json:"Path"`
}

// PlayState tracks playback position.
type PlayState struct {
	UserID   string `json:"UserId"`
	ItemID   string `json:"ItemId"`
	Position int64  `json:"PositionTicks"`
	Played   bool   `json:"Played"`
}
