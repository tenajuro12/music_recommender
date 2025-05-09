package dto

import (
	"spotify_recommender/internal/domain/entity"
	"time"
)

type UserDTO struct {
	ID          string         `json:"id"`
	Email       string         `json:"email"`
	Name        string         `json:"name"`
	SpotifyID   string         `json:"spotify_id,omitempty"`
	Preferences PreferencesDTO `json:"preferences"`
	LastLoginAt time.Time      `json:"last_login_at"`
	CreatedAt   time.Time      `json:"created_at"`
}

type PreferencesDTO struct {
	FavoriteGenres []string `json:"favorite_genres"`
	DislikedGenres []string `json:"disliked_genres"`
	MinTempo       float64  `json:"min_tempo"`
	MaxTempo       float64  `json:"max_tempo"`
	PreferredMoods []string `json:"preferred_moods"`
}

func UserFromEntity(user *entity.User) UserDTO {
	return UserDTO{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		SpotifyID: user.SpotifyID,
		Preferences: PreferencesDTO{
			FavoriteGenres: user.Preferences.FavoriteGenres,
			DislikedGenres: user.Preferences.DislikedGenres,
			MinTempo:       user.Preferences.MinTempo,
			MaxTempo:       user.Preferences.MaxTempo,
			PreferredMoods: user.Preferences.PreferredMoods,
		},
		LastLoginAt: user.LastLoginAt,
		CreatedAt:   user.CreatedAt,
	}
}

func (dto UserDTO) ToEntity() *entity.User {
	user := entity.NewUser(dto.Email, "", dto.Name)
	user.ID = dto.ID
	user.SpotifyID = dto.SpotifyID
	user.Preferences = entity.Preferences{
		FavoriteGenres: dto.Preferences.FavoriteGenres,
		DislikedGenres: dto.Preferences.DislikedGenres,
		MinTempo:       dto.Preferences.MinTempo,
		MaxTempo:       dto.Preferences.MaxTempo,
		PreferredMoods: dto.Preferences.PreferredMoods,
	}
	user.LastLoginAt = dto.LastLoginAt
	user.CreatedAt = dto.CreatedAt
	return user
}

func UsersFromEntities(users []*entity.User) []UserDTO {
	result := make([]UserDTO, len(users))
	for i, user := range users {
		result[i] = UserFromEntity(user)
	}
	return result
}

type CreateUserDTO struct {
	Email       string         `json:"email" binding:"required,email"`
	Password    string         `json:"password" binding:"required,min=8"`
	Name        string         `json:"name" binding:"required"`
	Preferences PreferencesDTO `json:"preferences"`
}

type UpdatePreferencesDTO struct {
	FavoriteGenres []string `json:"favorite_genres"`
	DislikedGenres []string `json:"disliked_genres"`
	MinTempo       float64  `json:"min_tempo"`
	MaxTempo       float64  `json:"max_tempo"`
	PreferredMoods []string `json:"preferred_moods"`
}
