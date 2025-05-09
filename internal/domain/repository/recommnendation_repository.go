package repository

import (
	"context"
	"spotify_recommender/internal/domain/entity"
	"spotify_recommender/internal/domain/valueObject"
)

type RecommendationRepository interface {
	GetByID(ctx context.Context, id string) (*entity.Recommendation, error)
	Save(ctx context.Context, recommendation *entity.Recommendation) error
	Delete(ctx context.Context, id string) error

	GetForUser(ctx context.Context, userID string) ([]*entity.Recommendation, error)

	FindByContext(
		ctx context.Context,
		userID string,
		mood valueObject.Mood,
		weather valueObject.Weather,
		timeOfDay valueObject.TimeOfDay,
	) (*entity.Recommendation, error)

	DeleteExpired(ctx context.Context) error
}
