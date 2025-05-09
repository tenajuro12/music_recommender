package entity

import (
	"spotify_recommender/internal/domain/valueObject"
	"time"
)

type Recommendation struct {
	ID        string                `json:"id"`
	UserID    string                `json:"user_id"`
	Mood      valueObject.Mood      `json:"mood"`
	Weather   valueObject.Weather   `json:"weather"`
	TimeOfDay valueObject.TimeOfDay `json:"time_of_day"`
	TrackIDs  []string              `json:"track_ids"`
	CreatedAt time.Time             `json:"created_at"`
	ExpiresAt time.Time             `json:"expires_at"`
}

func NewRecommendation(
	userID string,
	mood valueObject.Mood,
	weather valueObject.Weather,
	timeOfDay valueObject.TimeOfDay,
	trackIDs []string,
) *Recommendation {
	now := time.Now()
	return &Recommendation{
		UserID:    userID,
		Mood:      mood,
		Weather:   weather,
		TimeOfDay: timeOfDay,
		TrackIDs:  trackIDs,
		CreatedAt: now,
		ExpiresAt: now.Add(24 * time.Hour),
	}
}

func (r *Recommendation) IsExpired() bool {
	return time.Now().After(r.ExpiresAt)
}
