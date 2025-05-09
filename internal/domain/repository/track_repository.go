package repository

import (
	"context"
	"spotify_recommender/internal/domain/entity"
	"spotify_recommender/internal/domain/valueObject"
)

type TrackRepository interface {
	GetByID(ctx context.Context, id string) (*entity.Track, error)
	GetBySpotifyID(ctx context.Context, spotifyID string) (*entity.Track, error)
	Save(ctx context.Context, track *entity.Track) error
	Update(ctx context.Context, track *entity.Track) error
	Delete(ctx context.Context, id string) error

	FindByArtist(ctx context.Context, artist string) (*entity.Track, error)
	FindByName(ctx context.Context, name string) ([]*entity.Track, error)

	FindByMood(ctx context.Context, mood *valueObject.Mood, limit int) (*entity.Track, error)
	FindByWeather(ctx context.Context, weather valueObject.Weather, limit int) ([]*entity.Track, error)
	FindByTimeOfDay(ctx context.Context, timeOfDay valueObject.TimeOfDay, limit int) ([]*entity.Track, error)

	FindByMoodWeatherTime(ctx context.Context,
		mood *valueObject.Mood,
		weather valueObject.Weather,
		timeOfDay valueObject.TimeOfDay, limit int,
	) (*entity.Track, error)

	GetPopularTracks(ctx context.Context, limit int) ([]*entity.Track, error)
}
