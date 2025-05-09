package repository

import (
	"context"
	"spotify_recommender/internal/domain/entity"
	"spotify_recommender/internal/domain/valueObject"
)

type PlaylistRepository interface {
	GetByID(ctx context.Context, id string) (*entity.Playlist, error)
	Save(ctx context.Context, playlist *entity.Playlist) error
	Update(ctx context.Context, playlist *entity.Playlist) error
	Delete(ctx context.Context, id string) error

	GetUserPlaylists(ctx context.Context, userID string) ([]*entity.Playlist, error)

	FindByName(ctx context.Context, name string) ([]*entity.Playlist, error)
	FindByMood(ctx context.Context, mood valueObject.Mood) ([]*entity.Playlist, error)

	AddTrackToPlaylist(ctx context.Context, playlistID, trackID string) error
	RemoveTrackFromPlaylist(ctx context.Context, playlistID, trackID string) error
	GetPlaylistTracks(ctx context.Context, playlistID string) ([]*entity.Track, error)

	GetPublicPlaylists(ctx context.Context, limit, offset int) ([]*entity.Playlist, error)
}
