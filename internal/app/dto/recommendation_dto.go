package dto

import (
	"spotify_recommender/internal/domain/entity"
	"spotify_recommender/internal/domain/valueObject"
	"time"
)

type RecommendationDTO struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	Mood      string     `json:"mood"`
	Weather   string     `json:"weather"`
	TimeOfDay string     `json:"time_of_day"`
	Tracks    []TrackDTO `json:"tracks"`
	CreatedAt time.Time  `json:"created_at"`
}

type RecommendationRequestDTO struct {
	Mood      string `json:"mood" binding:"required"`
	Weather   string `json:"weather"`
	TimeOfDay string `json:"time_of_day"`
}

func RecommendationFromEntity(rec *entity.Recommendation, tracks []*entity.Track) RecommendationDTO {
	trackDTOs := TracksFromEntities(tracks)

	return RecommendationDTO{
		ID:        rec.ID,
		UserID:    rec.UserID,
		Mood:      string(rec.Mood),
		Weather:   string(rec.Weather),
		TimeOfDay: string(rec.TimeOfDay),
		Tracks:    trackDTOs,
		CreatedAt: rec.CreatedAt,
	}
}

func (dto RecommendationRequestDTO) ToEntity(userID string) *entity.Recommendation {
	mood := valueObject.Mood(dto.Mood)
	weather := valueObject.Weather(dto.Weather)
	timeOfDay := valueObject.TimeOfDay(dto.TimeOfDay)

	if !valueObject.ValidMood(mood) {
		mood = valueObject.MoodHappy
	}

	if !valueObject.ValidWeather(weather) {
		weather = valueObject.WeatherSunny
	}

	if dto.TimeOfDay == "" {
		timeOfDay = valueObject.GetCurrentTimeOfToday()
	}

	return entity.NewRecommendation(
		userID,
		mood,
		weather,
		timeOfDay,
		[]string{},
	)
}
