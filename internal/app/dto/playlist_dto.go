package dto

import (
	"spotify_recommender/internal/domain/entity"
	"spotify_recommender/internal/domain/valueObject"
	"time"
)

type PlaylistDTO struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Mood        string     `json:"mood"`
	Weather     string     `json:"weather,omitempty"`
	TimeOfDay   string     `json:"time_of_day,omitempty"`
	Tracks      []TrackDTO `json:"tracks,omitempty"`
	IsPublic    bool       `json:"is_public"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type CreatePlaylistDTO struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Mood        string `json:"mood" binding:"required"`
	IsPublic    bool   `json:"is_public"`
}

func PlaylistFromEntity(playlist *entity.Playlist, tracks []*entity.Track) PlaylistDTO {
	var trackDTOs []TrackDTO
	if tracks != nil {
		trackDTOs = TracksFromEntities(tracks)
	}

	return PlaylistDTO{
		ID:          playlist.ID,
		UserID:      playlist.UserID,
		Name:        playlist.Name,
		Description: playlist.Description,
		Mood:        string(playlist.Mood),
		Weather:     string(playlist.Weather),
		TimeOfDay:   string(playlist.TimeOfDay),
		Tracks:      trackDTOs,
		IsPublic:    playlist.IsPublic,
		CreatedAt:   playlist.CreatedAt,
		UpdatedAt:   playlist.UpdatedAt,
	}
}

func (dto CreatePlaylistDTO) ToEntity(userID string) *entity.Playlist {
	mood := valueObject.Mood(dto.Mood)
	if !valueObject.ValidMood(mood) {
		mood = valueObject.MoodHappy
	}

	playlist := entity.NewPlaylist(userID, dto.Name, dto.Description, mood)
	playlist.IsPublic = dto.IsPublic

	return playlist
}

func PlaylistsFromEntities(playlists []*entity.Playlist) []PlaylistDTO {
	result := make([]PlaylistDTO, len(playlists))
	for i, playlist := range playlists {
		result[i] = PlaylistFromEntity(playlist, nil)
	}
	return result
}
