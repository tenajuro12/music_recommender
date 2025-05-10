package usecase

import (
	"context"
	"spotify_recommender/internal/app/dto"
	"spotify_recommender/internal/domain/service"
)

type SavePlaylistUseCase struct {
	playlistService service.PlaylistService
}

func NewSavePlaylistUseCase(playlistService service.PlaylistService) *SavePlaylistUseCase {
	return &SavePlaylistUseCase{
		playlistService: playlistService,
	}
}

func (uc *SavePlaylistUseCase) Execute(ctx context.Context,
	userID string,
	createDTO dto.CreatePlaylistDTO) (*dto.PlaylistDTO, error) {
	playlist := createDTO.ToEntity(userID)

	savedPlaylist, err := uc.playlistService.CreatePlaylist(ctx, userID, playlist.Name, playlist.Description, playlist.Mood)
	if err != nil {
		return nil, err
	}

	playlistDTO := dto.PlaylistFromEntity(savedPlaylist, nil)

	return &playlistDTO, nil
}

type SavePlaylistFromRecommendationUseCase struct {
	playlistService *service.PlaylistService
}

func NewSavePlaylistFromRecommendationUseCase(
	playlistService *service.PlaylistService,
) *SavePlaylistFromRecommendationUseCase {
	return &SavePlaylistFromRecommendationUseCase{
		playlistService: playlistService,
	}
}

func (uc *SavePlaylistFromRecommendationUseCase) Execute(ctx context.Context,
	userID string,
	recommendationID string,
	name string,
	description string) (*dto.PlaylistDTO, error) {
	playlist, err := uc.playlistService.CreatePlaylistFromRecommendation(ctx, userID, recommendationID, name, description)
	if err != nil {
		return nil, err
	}
	playlistDTO := dto.PlaylistFromEntity(playlist, nil)
	return &playlistDTO, err
}
