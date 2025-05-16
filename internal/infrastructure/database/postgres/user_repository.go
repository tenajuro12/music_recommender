package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"spotify_recommender/internal/domain/entity"
	"spotify_recommender/internal/domain/repository"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) repository.UserRepository {
	return &UserRepository{
		db: db,
	}
}

type userModel struct {
	ID           string          `db:"id"`
	Email        string          `db:"email"`
	PasswordHash string          `db:"password_hash"`
	Name         string          `db:"name"`
	SpotifyID    sql.NullString  `db:"spotify_id"`
	Preferences  json.RawMessage `db:"preferences"`
	LastLoginAt  time.Time       `db:"last_login_at"`
	CreatedAt    time.Time       `db:"created_at"`
	UpdatedAt    time.Time       `db:"updated_at"`
}

func (m *userModel) toEntity() (*entity.User, error) {
	var preferences entity.Preferences

	if len(m.Preferences) > 0 {
		err := json.Unmarshal(m.Preferences, &preferences)
		if err != nil {
			log.Error().Err(err).
				Str("user_id", m.ID).
				Msg("failed to unmarshal user preferences")

			preferences = entity.Preferences{
				FavoriteGenres: []string{},
				DislikedGenres: []string{},
				MinTempo:       0,
				MaxTempo:       250,
				PreferredMoods: []string{},
			}
		}
	}

	user := entity.NewUser(m.Email, m.PasswordHash, m.Name)
	user.ID = m.ID

	if m.SpotifyID.Valid {
		user.SpotifyID = m.SpotifyID.String
	}

	user.Preferences = preferences
	user.LastLoginAt = m.LastLoginAt
	user.CreatedAt = m.CreatedAt
	user.UpdatedAt = m.UpdatedAt

	return user, nil
}

func fromUserEntity(user *entity.User) (*userModel, error) {
	preferencesJSON, err := json.Marshal(user.Preferences)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal user preferences: %w", err)
	}

	model := &userModel{
		ID:           user.ID,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		Name:         user.Name,
		Preferences:  preferencesJSON,
		LastLoginAt:  user.LastLoginAt,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}

	if user.SpotifyID != "" {
		model.SpotifyID = sql.NullString{
			String: user.SpotifyID,
			Valid:  true,
		}
	}

	return model, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*entity.User, error) {
	query := `
		SELECT * FROM users
		WHERE id = $1
	`

	var model userModel
	err := r.db.GetContext(ctx, &model, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return model.toEntity()
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `
		SELECT * FROM users
		WHERE email = $1
	`

	var model userModel
	err := r.db.GetContext(ctx, &model, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return model.toEntity()
}

func (r *UserRepository) GetBySpotifyID(ctx context.Context, spotifyID string) (*entity.User, error) {
	query := `
		SELECT * FROM users
		WHERE spotify_id = $1
	`

	var model userModel
	err := r.db.GetContext(ctx, &model, query, spotifyID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get user by Spotify ID: %w", err)
	}

	return model.toEntity()
}

func (r *UserRepository) Save(ctx context.Context, user *entity.User) error {
	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	now := time.Now()
	if user.CreatedAt.IsZero() {
		user.CreatedAt = now
	}
	if user.UpdatedAt.IsZero() {
		user.UpdatedAt = now
	}
	if user.LastLoginAt.IsZero() {
		user.LastLoginAt = now
	}

	existingUser, err := r.GetByEmail(ctx, user.Email)
	if err == nil && existingUser != nil {
		return fmt.Errorf("user with email %s already exists", user.Email)
	}

	model, err := fromUserEntity(user)
	if err != nil {
		return fmt.Errorf("failed to convert user to model: %w", err)
	}

	query := `
		INSERT INTO users (
			id, email, password_hash, name, spotify_id, preferences,
			last_login_at, created_at, updated_at
		) VALUES (
			:id, :email, :password_hash, :name, :spotify_id, :preferences,
			:last_login_at, :created_at, :updated_at
		)
	`

	_, err = r.db.NamedExecContext(ctx, query, model)
	if err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}

	return nil
}

func (r *UserRepository) Update(ctx context.Context, user *entity.User) error {
	user.UpdatedAt = time.Now()

	model, err := fromUserEntity(user)
	if err != nil {
		return fmt.Errorf("failed to convert user to model: %w", err)
	}

	query := `
		UPDATE users SET
			email = :email,
			password_hash = :password_hash,
			name = :name,
			spotify_id = :spotify_id,
			preferences = :preferences,
			last_login_at = :last_login_at,
			updated_at = :updated_at
		WHERE id = :id
	`

	result, err := r.db.NamedExecContext(ctx, query, model)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user with ID %s not found", user.ID)
	}

	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
	_, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	query := `DELETE FROM users WHERE id = $1`

	_, err = r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

func (r *UserRepository) FindByName(ctx context.Context, name string) ([]*entity.User, error) {
	query := `
		SELECT * FROM users
		WHERE name ILIKE $1
		ORDER BY name
	`

	rows, err := r.db.QueryxContext(ctx, query, "%"+name+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to find users by name: %w", err)
	}
	defer rows.Close()

	var users []*entity.User

	for rows.Next() {
		var model userModel

		if err := rows.StructScan(&model); err != nil {
			return nil, fmt.Errorf("failed to scan user model: %w", err)
		}

		user, err := model.toEntity()
		if err != nil {
			log.Error().Err(err).
				Str("user_id", model.ID).
				Msg("failed to convert user model to entity")
			continue
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating through users: %w", err)
	}

	return users, nil
}

func (r *UserRepository) UpdatePreferences(ctx context.Context, userID string, preferences entity.Preferences) error {
	user, err := r.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	user.Preferences = preferences

	return r.Update(ctx, user)
}

func (r *UserRepository) LogTrackInteraction(ctx context.Context, userID, trackID string, liked bool) error {
	_, err := r.GetByID(ctx, userID)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO user_track_interactions (
			user_id, track_id, liked, created_at
		) VALUES (
			$1, $2, $3, $4
		)
	`

	_, err = r.db.ExecContext(ctx, query, userID, trackID, liked, time.Now())
	if err != nil {
		return fmt.Errorf("failed to log track interaction: %w", err)
	}

	return nil
}

func (r *UserRepository) GetUserLikedTracks(ctx context.Context, userID string, limit, offset int) ([]*entity.Track, int, error) {
	query := `
		SELECT t.*, COUNT(*) OVER() AS total_count
		FROM tracks t
		JOIN user_track_interactions uti ON t.id = uti.track_id
		WHERE uti.user_id = $1 AND uti.liked = TRUE
		GROUP BY t.id
		ORDER BY MAX(uti.created_at) DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryxContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get user liked tracks: %w", err)
	}
	defer rows.Close()

	var tracks []*entity.Track
	var totalCount int

	for rows.Next() {
		var model struct {
			trackModel
			TotalCount int `db:"total_count"`
		}

		if err := rows.StructScan(&model); err != nil {
			return nil, 0, fmt.Errorf("failed to scan track model: %w", err)
		}

		track, err := model.trackModel.ToEntity()
		if err != nil {
			log.Error().Err(err).
				Str("track_id", model.ID).
				Msg("failed to convert track model to entity")
			continue
		}

		tracks = append(tracks, track)
		totalCount = model.TotalCount
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating through tracks: %w", err)
	}

	return tracks, totalCount, nil
}

func (r *UserRepository) GetUserListeningHistory(ctx context.Context, userID string, limit, offset int) ([]*entity.Track, int, error) {
	query := `
		SELECT t.*, COUNT(*) OVER() AS total_count
		FROM tracks t
		JOIN user_track_interactions uti ON t.id = uti.track_id
		WHERE uti.user_id = $1
		GROUP BY t.id
		ORDER BY MAX(uti.created_at) DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryxContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get user listening history: %w", err)
	}
	defer rows.Close()

	var tracks []*entity.Track
	var totalCount int

	for rows.Next() {
		var model struct {
			trackModel
			TotalCount int `db:"total_count"`
		}

		if err := rows.StructScan(&model); err != nil {
			return nil, 0, fmt.Errorf("failed to scan track model: %w", err)
		}

		track, err := model.trackModel.ToEntity()
		if err != nil {
			log.Error().Err(err).
				Str("track_id", model.ID).
				Msg("failed to convert track model to entity")
			continue
		}

		tracks = append(tracks, track)
		totalCount = model.TotalCount
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating through tracks: %w", err)
	}

	return tracks, totalCount, nil
}

func (r *UserRepository) GetUserRecommendationStats(ctx context.Context, userID string) (map[string]int, error) {
	query := `
		SELECT 
			COUNT(*) FILTER (WHERE liked = TRUE) AS liked_count,
			COUNT(*) FILTER (WHERE liked = FALSE) AS disliked_count,
			COUNT(*) AS total_count
		FROM user_track_interactions
		WHERE user_id = $1
	`

	var stats struct {
		LikedCount    int `db:"liked_count"`
		DislikedCount int `db:"disliked_count"`
		TotalCount    int `db:"total_count"`
	}

	err := r.db.GetContext(ctx, &stats, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user recommendation stats: %w", err)
	}

	return map[string]int{
		"liked":    stats.LikedCount,
		"disliked": stats.DislikedCount,
		"total":    stats.TotalCount,
	}, nil
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID string) error {
	query := `
		UPDATE users
		SET last_login_at = $1, updated_at = $1
		WHERE id = $2
	`

	now := time.Now()

	result, err := r.db.ExecContext(ctx, query, now, userID)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user with ID %s not found", userID)
	}

	return nil
}

func (r *UserRepository) ConnectSpotifyAccount(ctx context.Context, userID, spotifyID string) error {
	query := `
		UPDATE users
		SET spotify_id = $1, updated_at = $2
		WHERE id = $3
	`

	now := time.Now()

	result, err := r.db.ExecContext(ctx, query, spotifyID, now, userID)
	if err != nil {
		return fmt.Errorf("failed to connect Spotify account: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user with ID %s not found", userID)
	}

	return nil
}

func (r *UserRepository) DisconnectSpotifyAccount(ctx context.Context, userID string) error {
	query := `
		UPDATE users
		SET spotify_id = NULL, updated_at = $1
		WHERE id = $2
	`

	now := time.Now()

	result, err := r.db.ExecContext(ctx, query, now, userID)
	if err != nil {
		return fmt.Errorf("failed to disconnect Spotify account: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user with ID %s not found", userID)
	}

	return nil
}
