package entity

import "time"

type Recommendation struct {
	ID        string                `json:"id"`
	UserID    string                `json:"user_id"`
	Mood      valueobject.Mood      `json:"mood"`
	Weather   valueobject.Weather   `json:"weather"`
	TimeOfDay valueobject.TimeOfDay `json:"time_of_day"`
	TrackIDs  []string              `json:"track_ids"`
	CreatedAt time.Time             `json:"created_at"`
	ExpiresAt time.Time             `json:"expires_at"`
}

func NewRecommendation(
	userID string,
	mood valueobject.Mood,
	weather valueobject.Weather,
	timeOfDay valueobject.TimeOfDay,
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
