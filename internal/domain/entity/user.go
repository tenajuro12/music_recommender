package entity

import "time"

type User struct {
	ID           string      `json:"id"`
	Email        string      `json:"email"`
	PasswordHash string      `json:"password_hash"`
	Name         string      `json:"name"`
	SpotifyID    string      `json:"spotify_id,omitempty"`
	Preferences  Preferences `json:"preferences"`
	LastLoginAt  time.Time   `json:"last_login_at"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}
type Preferences struct {
	FavoriteGenres []string `json:"favorite_genres"`
	DislikedGenres []string `json:"disliked_genres"`
	MinTempo       float64  `json:"min_tempo"`
	MaxTempo       float64  `json:"max_tempo"`
	PreferredMoods []string `json:"preferred_moods"`
}

func NewUser(email, passwordHash, name string) *User {
	return &User{
		Email:        email,
		PasswordHash: passwordHash,
		Name:         name,
		Preferences: Preferences{
			FavoriteGenres: []string{},
			DislikedGenres: []string{},
			MinTempo:       0,
			MaxTempo:       250,
			PreferredMoods: []string{},
		},
		LastLoginAt: time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}
