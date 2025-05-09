package repository

import (
	"context"
	"spotify_recommender/internal/domain/entity"
)

type User interface {
	GetByID(ctx context.Context, id string) (*entity.User, error)
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	GetBySpotifyID(ctx context.Context, spotifyID string) (*entity.User, error)
	Save(ctx context.Context, user *entity.User) error
	Update(ctx context.Context, user *entity.User) error
	Delete(ctx context.Context, id string) error

	FindByName(ctx context.Context, name string) ([]*entity.User, error)

	UpdatePreferences(ctx context.Context, userID string, preferences entity.Preferences) error

	LogTrackInteraction(ctx context.Context, userID, trackID string, liked bool) error
}
