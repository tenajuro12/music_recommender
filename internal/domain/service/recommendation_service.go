package service

import (
	"context"
	"errors"
	"math/rand"
	"spotify_recommender/internal/domain/entity"
	"spotify_recommender/internal/domain/repository"
	"spotify_recommender/internal/domain/valueObject"
	"time"
)

var ErrNoRecommendations = errors.New("no suitable recommendations found")

type RecommendationService struct {
	userRepo           repository.UserRepository
	trackRepo          repository.TrackRepository
	recommendationRepo repository.RecommendationRepository
}

func NewRecommendationService(userRepo repository.UserRepository,
	trackRepo repository.TrackRepository,
	recommendationRepo repository.RecommendationRepository) *RecommendationService {
	return &RecommendationService{
		trackRepo:          trackRepo,
		userRepo:           userRepo,
		recommendationRepo: recommendationRepo,
	}
}

func (s *RecommendationService) GetRecommendationsByContext(
	ctx context.Context,
	userID string,
	mood valueObject.Mood,
	weather valueObject.Weather,
	timeOfDay valueObject.TimeOfDay,
) (*entity.Recommendation, error) {

	cachedRec, err := s.recommendationRepo.FindByContext(ctx, userID, mood, weather, timeOfDay)
	if err == nil && cachedRec != nil && !cachedRec.IsExpired() {
		return cachedRec, nil
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	tracks, err := s.trackRepo.FindByMoodWeatherTime(ctx, mood, weather, timeOfDay, 50)
	if err != nil {
		return nil, err
	}

	if len(tracks) == 0 {
		return nil, ErrNoRecommendations
	}

	filteredTracks := s.filterTracksByUserPreferences(tracks, user.Preferences)

	if len(filteredTracks) == 0 {
		filteredTracks = tracks
	}

	//делаем перемешку
	recommendedTracks := s.selectRecommendedTracks(filteredTracks, 20)
	trackIDs := make([]string, len(recommendedTracks))

	for i, track := range recommendedTracks {
		trackIDs[i] = track.ID
	}

	recommendation := entity.NewRecommendation(
		userID,
		mood,
		weather,
		timeOfDay,
		trackIDs,
	)
	err = s.recommendationRepo.Save(ctx, recommendation)
	if err != nil {
		return nil, err
	}

	return recommendation, nil

}

func (s *RecommendationService) filterTracksByUserPreferences(
	tracks []*entity.Track,
	preferences entity.Preferences,
) []*entity.Track {
	var filtered []*entity.Track
	for _, track := range tracks {
		if track.AudioFeatures.Tempo < preferences.MinTempo ||
			track.AudioFeatures.Tempo > preferences.MaxTempo {
			continue
		}
		filtered = append(filtered, track)
	}
	return filtered
}

func (s *RecommendationService) selectRecommendedTracks(
	tracks []*entity.Track,
	count int,
) []*entity.Track {
	if len(tracks) <= count {
		return tracks
	}

	shuffled := make([]*entity.Track, len(tracks))
	copy(shuffled, tracks)

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	r.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})
	return shuffled[:count]
}

func (s *RecommendationService) GetRecommendationsByMood(
	ctx context.Context,
	userID string,
	mood valueObject.Mood,
) (*entity.Recommendation, error) {
	weather := valueObject.WeatherSunny
	timeOfDay := valueObject.GetCurrentTimeOfToday()

	return s.GetRecommendationsByContext(ctx, userID, mood, weather, timeOfDay)
}

func (s *RecommendationService) SaveUserTrackInteraction(
	ctx context.Context,
	userID string,
	trackID string,
	liked bool,
) error {
	return s.userRepo.LogTrackInteraction(ctx, userID, trackID, liked)
}
