package dto

import (
	"spotify_recommender/internal/domain/entity"
	"spotify_recommender/internal/domain/valueObject"
	"time"
)

type TrackDTO struct {
	ID            string           `json:"id"`
	SpotifyID     string           `json:"spotify_id"`
	Name          string           `json:"name"`
	Artist        string           `json:"artist"`
	Album         string           `json:"album"`
	ReleaseDate   time.Time        `json:"release_date"`
	Popularity    int              `json:"popularity"`
	AudioFeatures AudioFeaturesDTO `json:"audio_features"`
	PreviewURL    string           `json:"preview_url"`
	ImageURL      string           `json:"image_url"`
}

type AudioFeaturesDTO struct {
	Danceability     float64 `json:"danceability"`
	Energy           float64 `json:"energy"`
	Key              int     `json:"key"`
	Loudness         float64 `json:"loudness"`
	Mode             int     `json:"mode"`
	Speechiness      float64 `json:"speechiness"`
	Acousticness     float64 `json:"acousticness"`
	Instrumentalness float64 `json:"instrumentalness"`
	Liveness         float64 `json:"liveness"`
	Valence          float64 `json:"valence"`
	Tempo            float64 `json:"tempo"`
	Duration         int     `json:"duration_ms"`
	TimeSignature    int     `json:"time_signature"`
}

func TrackFromEntity(track *entity.Track) TrackDTO {
	return TrackDTO{
		ID:          track.ID,
		SpotifyID:   track.SpotifyID,
		Name:        track.Name,
		Artist:      track.Artist,
		Album:       track.Album,
		ReleaseDate: track.ReleaseDate,
		Popularity:  track.Popularity,
		AudioFeatures: AudioFeaturesDTO{
			Danceability:     track.AudioFeatures.Danceability,
			Energy:           track.AudioFeatures.Energy,
			Key:              track.AudioFeatures.Key,
			Loudness:         track.AudioFeatures.Loudness,
			Mode:             track.AudioFeatures.Mode,
			Speechiness:      track.AudioFeatures.Speechiness,
			Acousticness:     track.AudioFeatures.Acousticness,
			Instrumentalness: track.AudioFeatures.Instrumentalness,
			Liveness:         track.AudioFeatures.Liveness,
			Valence:          track.AudioFeatures.Valence,
			Tempo:            float64(track.AudioFeatures.Tempo),
			Duration:         track.AudioFeatures.Duration,
			TimeSignature:    track.AudioFeatures.TimeSignature,
		},
		PreviewURL: track.PreviewURL,
		ImageURL:   track.ImageURL,
	}
}

func (dto TrackDTO) ToEntity() *entity.Track {
	return entity.NewTrack(
		dto.SpotifyID,
		dto.Name,
		dto.Artist,
		dto.Album,
		dto.ReleaseDate,
		dto.Popularity,
		valueObject.AudioFeatures{
			Danceability:     dto.AudioFeatures.Danceability,
			Energy:           dto.AudioFeatures.Energy,
			Key:              dto.AudioFeatures.Key,
			Loudness:         dto.AudioFeatures.Loudness,
			Mode:             dto.AudioFeatures.Mode,
			Speechiness:      dto.AudioFeatures.Speechiness,
			Acousticness:     dto.AudioFeatures.Acousticness,
			Instrumentalness: dto.AudioFeatures.Instrumentalness,
			Liveness:         dto.AudioFeatures.Liveness,
			Valence:          dto.AudioFeatures.Valence,
			Tempo:            dto.AudioFeatures.Tempo,
			Duration:         dto.AudioFeatures.Duration,
			TimeSignature:    dto.AudioFeatures.TimeSignature,
		},
		dto.PreviewURL,
		dto.ImageURL,
	)
}

func TracksFromEntities(tracks []*entity.Track) []TrackDTO {
	result := make([]TrackDTO, len(tracks))
	for i, track := range tracks {
		result[i] = TrackFromEntity(track)
	}
	return result
}
