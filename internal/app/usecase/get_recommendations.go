package usecase

import (
	"context"
	"errors"
	"spotify_recommender/internal/app/dto"
	"spotify_recommender/internal/domain/entity"
	"spotify_recommender/internal/domain/repository"
	"spotify_recommender/internal/domain/service"
	"spotify_recommender/internal/domain/valueObject"
)

type GetRecommendations struct {
	recommendationService *service.RecommendationService
	trackRepository       repository.TrackRepository
	weatherService        WeatherService
}

type WeatherService interface {
	GetCurrentWeather(ctx context.Context, latitude, longitude float64) (valueObject.Weather, error)
}

func NewGetRecommendationsUseCase(
	recommendationService *service.RecommendationService,
	trackRepository repository.TrackRepository,
	weatherService WeatherService,
) *GetRecommendations {
	return &GetRecommendations{
		recommendationService: recommendationService,
		trackRepository:       trackRepository,
		weatherService:        weatherService,
	}
}

func (uc *GetRecommendations) Execute(ctx context.Context,
	userID string,
	request dto.RecommendationDTO,
	lat, lon float64) (*dto.RecommendationDTO, error) {
	mood := valueObject.Mood(request.Mood)

	if !valueObject.ValidMood(mood) {
		return nil, errors.New("invalid mood value")
	}

	var weather valueObject.Weather
	if request.Weather == "" {
		var err error
		weather, err = uc.weatherService.GetCurrentWeather(ctx, lat, lon)
		if err != nil {
			weather = valueObject.WeatherSunny
		}
	} else {
		weather = valueObject.Weather(request.Weather)
		if !valueObject.ValidWeather(weather) {
			return nil, errors.New("invalid weather value")
		}
	}

	var timeOfDay valueObject.TimeOfDay
	if request.TimeOfDay == "" {
		timeOfDay = valueObject.GetCurrentTimeOfToday()
	} else {
		timeOfDay = valueObject.TimeOfDay(request.TimeOfDay)
		isValid := false
		for _, t := range valueObject.AllTimesOfDay() {
			if timeOfDay == t {
				isValid = true
				break
			}
		}
		if !isValid {
			return nil, errors.New("invalid time of day value")
		}
	}

	recommendation, err := uc.recommendationService.GetRecommendationsByContext(
		ctx, userID, mood, weather, timeOfDay,
	)
	if err != nil {
		return nil, err
	}

	var tracks []*entity.Track
	for _, trackID := range recommendation.TrackIDs {
		track, err := uc.trackRepository.GetByID(ctx, trackID)
		if err != nil {
			continue
		}
		tracks = append(tracks, track)
	}

	recommendationDTO := dto.RecommendationFromEntity(recommendation, tracks)

	return &recommendationDTO, nil

}
