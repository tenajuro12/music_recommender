package entity

import (
	"spotify_recommender/internal/domain/valueObject"
	"time"
)

type Track struct {
	ID            string                    `json:"id"`
	SpotifyID     string                    `json:"spotify_id"`
	Name          string                    `json:"name"`
	Artist        string                    `json:"artist"`
	Album         string                    `json:"album"`
	ReleaseDate   time.Time                 `json:"release_date"`
	Popularity    int                       `json:"popularity"`
	AudioFeatures valueObject.AudioFeatures `json:"audio_features"`
	PreviewURL    string                    `json:"preview_url"`
	ImageURL      string                    `json:"image_url"`
	CreatedAt     time.Time                 `json:"created_at"`
	UpdatedAt     time.Time                 `json:"updated_at"`
}

func NewTrack(
	spotifyID, name, artist, album string,
	releaseDate time.Time,
	popularity int,
	audioFeatures valueObject.AudioFeatures,
	previewURL, imageURL string,
) *Track {
	return &Track{
		SpotifyID:     spotifyID,
		Name:          name,
		Artist:        artist,
		Album:         album,
		ReleaseDate:   releaseDate,
		Popularity:    popularity,
		AudioFeatures: audioFeatures,
		PreviewURL:    previewURL,
		ImageURL:      imageURL,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}
