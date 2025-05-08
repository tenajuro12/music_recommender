package entity

import "time"

type Track struct {
	ID            string                    `json:"id"`
	SpotifyID     string                    `json:"spotify_id"`
	Name          string                    `json:"name"`
	Artist        string                    `json:"artist"`
	Album         string                    `json:"album"`
	ReleaseDate   time.Time                 `json:"release_date"`
	Popularity    int                       `json:"popularity"`
	AudioFeatures valueobject.AudioFeatures `json:"audio_features"`
	PreviewURL    string                    `json:"preview_url"`
	ImageURL      string                    `json:"image_url"`
	CreatedAt     time.Time                 `json:"created_at"`
	UpdatedAt     time.Time                 `json:"updated_at"`
}

func newTrack(spotifyID, name, album string, releaseDate time.Time, popularity int,
	audioFeatures valueobject.AudioFeatures, previewURL, imageURL string) *Track {
	return &Track{
		SpotifyID:     spotifyID,
		Name:          name,
		Album:         album,
		ReleaseDate:   releaseDate,
		Popularity:    popularity,
		AudioFeatures: audioFeatures,
		PreviewURL:    previewURL,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}
