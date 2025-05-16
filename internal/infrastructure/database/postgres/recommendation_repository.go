package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"spotify_recommender/internal/domain/entity"
	"spotify_recommender/internal/domain/valueObject"
	"time"
)

type RecommendationRepository struct {
	db *sqlx.DB
}

func NewRecommendationRepository(db *sqlx.DB) *RecommendationRepository {
	return &RecommendationRepository{
		db: db,
	}
}

type recommendationModel struct {
	ID        string          `db:"id"`
	UserID    string          `db:"user_id"`
	Mood      string          `db:"mood"`
	Weather   string          `db:"weather"`
	TimeOfDay string          `db:"time_of_day"`
	TrackIDs  json.RawMessage `db:"track_ids"`
	CreatedAt time.Time       `db:"created_at"`
	ExpiresAt time.Time       `db:"expires_at"`
}

func (m *recommendationModel) toEntity() (*entity.Recommendation, error) {
	var trackIDs []string

	if len(m.TrackIDs) > 0 {
		err := json.Unmarshal(m.TrackIDs, &trackIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal track IDs: %w", err)
		}
	}

	recommendation := entity.NewRecommendation(
		m.UserID,
		valueObject.Mood(m.Mood),
		valueObject.Weather(m.Weather),
		valueObject.TimeOfDay(m.TimeOfDay),
		trackIDs,
	)

	recommendation.ID = m.ID
	recommendation.CreatedAt = m.CreatedAt
	recommendation.ExpiresAt = m.ExpiresAt

	return recommendation, nil
}

func fromRecommendationEntity(rec *entity.Recommendation) (*recommendationModel, error) {
	trackIDsJSON, err := json.Marshal(rec.TrackIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal track IDs: %w", err)
	}

	return &recommendationModel{
		ID:        rec.ID,
		UserID:    rec.UserID,
		Mood:      string(rec.Mood),
		Weather:   string(rec.Weather),
		TimeOfDay: string(rec.TimeOfDay),
		TrackIDs:  trackIDsJSON,
		CreatedAt: rec.CreatedAt,
		ExpiresAt: rec.ExpiresAt,
	}, nil
}

func (r *RecommendationRepository) GetByID(ctx context.Context, id string) (*entity.Recommendation, error) {
	query := `
		SELECT * FROM recommendations
		WHERE id = $1
	`

	var model recommendationModel
	err := r.db.GetContext(ctx, &model, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("recommendation not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get recommendation by ID: %w", err)
	}

	return model.toEntity()
}

func (r *RecommendationRepository) Save(ctx context.Context, recommendation *entity.Recommendation) error {

	if recommendation.ID == "" {
		recommendation.ID = uuid.New().String()
	}

	now := time.Now()
	if recommendation.CreatedAt.IsZero() {
		recommendation.CreatedAt = now
	}
	if recommendation.ExpiresAt.IsZero() {
		recommendation.ExpiresAt = now.Add(24 * time.Hour)
	}

	model, err := fromRecommendationEntity(recommendation)
	if err != nil {
		return fmt.Errorf("failed to convert recommendation to model: %w", err)
	}

	existingRec, err := r.FindByContext(
		ctx,
		recommendation.UserID,
		recommendation.Mood,
		recommendation.Weather,
		recommendation.TimeOfDay,
	)

	if err == nil && existingRec != nil {
		existingRec.TrackIDs = recommendation.TrackIDs
		existingRec.CreatedAt = now
		existingRec.ExpiresAt = recommendation.ExpiresAt

		modelToUpdate, err := fromRecommendationEntity(existingRec)
		if err != nil {
			return fmt.Errorf("failed to convert updated recommendation to model: %w", err)
		}

		query := `
			UPDATE recommendations SET
				track_ids = :track_ids,
				created_at = :created_at,
				expires_at = :expires_at
			WHERE id = :id
		`

		_, err = r.db.NamedExecContext(ctx, query, modelToUpdate)
		if err != nil {
			return fmt.Errorf("failed to update existing recommendation: %w", err)
		}

		return nil
	}

	query := `
		INSERT INTO recommendations (
			id, user_id, mood, weather, time_of_day, track_ids, created_at, expires_at
		) VALUES (
			:id, :user_id, :mood, :weather, :time_of_day, :track_ids, :created_at, :expires_at
		)
	`

	_, err = r.db.NamedExecContext(ctx, query, model)
	if err != nil {
		return fmt.Errorf("failed to save recommendation: %w", err)
	}

	return nil
}

func (r *RecommendationRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM recommendations WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete recommendation: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("recommendation with ID %s not found", id)
	}

	return nil
}

func (r *RecommendationRepository) GetForUser(ctx context.Context, userID string) ([]*entity.Recommendation, error) {
	query := `
		SELECT * FROM recommendations
		WHERE user_id = $1
		AND expires_at > $2
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryxContext(ctx, query, userID, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to get recommendations for user: %w", err)
	}
	defer rows.Close()

	var recommendations []*entity.Recommendation

	for rows.Next() {
		var model recommendationModel

		if err := rows.StructScan(&model); err != nil {
			return nil, fmt.Errorf("failed to scan recommendation model: %w", err)
		}

		recommendation, err := model.toEntity()
		if err != nil {
			continue
		}

		recommendations = append(recommendations, recommendation)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating through recommendations: %w", err)
	}

	return recommendations, nil
}

func (r *RecommendationRepository) FindByContext(
	ctx context.Context,
	userID string,
	mood valueObject.Mood,
	weather valueObject.Weather,
	timeOfDay valueObject.TimeOfDay,
) (*entity.Recommendation, error) {
	query := `
		SELECT * FROM recommendations
		WHERE user_id = $1
		AND mood = $2
		AND weather = $3
		AND time_of_day = $4
		AND expires_at > $5
		ORDER BY created_at DESC
		LIMIT 1
	`

	var model recommendationModel
	err := r.db.GetContext(
		ctx,
		&model,
		query,
		userID,
		string(mood),
		string(weather),
		string(timeOfDay),
		time.Now(),
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find recommendation by context: %w", err)
	}

	return model.toEntity()
}

func (r *RecommendationRepository) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM recommendations WHERE expires_at <= $1`

	_, err := r.db.ExecContext(ctx, query, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete expired recommendations: %w", err)
	}

	return nil
}
