package service

import (
	"context"
	"errors"
	"fmt"
	"spotify_recommender/internal/domain/entity"
	"spotify_recommender/internal/domain/repository"
	"spotify_recommender/internal/domain/valueObject"
)

var (
	ErrPlaylistNotFound = errors.New("playlist not found")
	ErrTrackNotFound    = errors.New("track not found")
	ErrUserNotFound     = errors.New("user not found")
)

type PlaylistService struct {
	playlistRepo       repository.PlaylistRepository
	trackRepo          repository.TrackRepository
	userRepo           repository.UserRepository
	recommendationRepo repository.RecommendationRepository
}

func NewPlaylistService(playlistRepo repository.PlaylistRepository,
	trackRepo repository.TrackRepository,
	userRepo repository.UserRepository,
	recommendationRepo repository.RecommendationRepository) *PlaylistService {
	return &PlaylistService{
		playlistRepo:       playlistRepo,
		trackRepo:          trackRepo,
		userRepo:           userRepo,
		recommendationRepo: recommendationRepo,
	}
}

func (s *PlaylistService) CreatePlaylist(ctx context.Context,
	userID, name, description string,
	mood valueObject.Mood) (*entity.Playlist, error) {
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	playlist := entity.NewPlaylist(userID, name, description, mood)

	err = s.playlistRepo.Save(ctx, playlist)
	if err != nil {
		return nil, err
	}
	return playlist, nil
}
func (s *PlaylistService) GetUserPlaylists(
	ctx context.Context,
	userID string,
) ([]*entity.Playlist, error) {
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	playlists, err := s.playlistRepo.GetUserPlaylists(ctx, userID)
	if err != nil {
		return nil, ErrPlaylistNotFound
	}
	return playlists, err
}

func (s *PlaylistService) AddTrackToPlaylist(
	ctx context.Context,
	playlistID, trackID string,
) error {
	playlist, err := s.playlistRepo.GetByID(ctx, playlistID)
	if err != nil {
		return ErrPlaylistNotFound
	}

	_, err = s.trackRepo.GetByID(ctx, trackID)
	if err != nil {
		return ErrTrackNotFound
	}

	return s.playlistRepo.AddTrackToPlaylist(ctx, playlist.ID, trackID)
}

func (s *PlaylistService) RemoveTrackFromPlaylist(
	ctx context.Context,
	playlistID, trackID string,
) error {
	_, err := s.playlistRepo.GetByID(ctx, playlistID)
	if err != nil {
		return ErrPlaylistNotFound
	}
	_, err = s.trackRepo.GetByID(ctx, trackID)
	if err != nil {
		return ErrTrackNotFound
	}
	return s.playlistRepo.RemoveTrackFromPlaylist(ctx, playlistID, trackID)
}

func (s *PlaylistService) GetPlaylistTracks(
	ctx context.Context,
	playlistID string,
) ([]*entity.Track, error) {
	_, err := s.playlistRepo.GetByID(ctx, playlistID)
	if err != nil {
		return nil, ErrPlaylistNotFound
	}
	tracks, err := s.playlistRepo.GetPlaylistTracks(ctx, playlistID)
	if err != nil {
		return nil, ErrTrackNotFound
	}

	return tracks, nil
}

func (s *PlaylistService) DeletePlaylist(
	ctx context.Context,
	playlistID string,
) error {
	_, err := s.playlistRepo.GetByID(ctx, playlistID)
	if err != nil {
		return ErrPlaylistNotFound
	}

	return s.playlistRepo.Delete(ctx, playlistID)
}

func (s *PlaylistService) CreatePlaylistFromRecommendation(
	ctx context.Context,
	userID string,
	recommendationID string,
	name, description string,
) (*entity.Playlist, error) {
	recommendation, err := s.recommendationRepo.GetByID(ctx, recommendationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recommendation by ID %s: %w", recommendationID, err)
	}

	playlist := entity.NewPlaylist(userID, name, description, recommendation.Mood)

	playlist.Tracks = recommendation.TrackIDs

	if err := s.playlistRepo.Save(ctx, playlist); err != nil {
		return nil, fmt.Errorf("failed to save playlist: %w", err)
	}

	return playlist, nil
}
