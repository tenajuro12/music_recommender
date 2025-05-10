package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"log"
	"sort"
	"spotify_recommender/internal/domain/entity"
	"spotify_recommender/internal/domain/valueObject"
	"time"
)

type TrackRepository struct {
	db *sqlx.DB
}

func NewTrackRepository(db *sqlx.DB) *TrackRepository {
	return &TrackRepository{
		db: db,
	}
}

type trackModel struct {
	ID            string          `db:"id"`
	SpotifyID     string          `db:"spotify_id"`
	Name          string          `db:"name"`
	Artist        string          `db:"artist"`
	Album         string          `db:"album"`
	ReleaseDate   time.Time       `db:"release_date"`
	Popularity    int             `db:"popularity"`
	AudioFeatures json.RawMessage `db:"audio_features"`
	PreviewURL    string          `db:"preview_url"`
	ImageURL      string          `db:"image_url"`
	CreatedAt     time.Time       `db:"created_at"`
	UpdatedAt     time.Time       `db:"updated_at"`
}

func (m *trackModel) ToEntity() (*entity.Track, error) {
	var audioFeatures valueObject.AudioFeatures
	err := json.Unmarshal(m.AudioFeatures, &audioFeatures)
	if err != nil {
		return nil, err
	}
	track := entity.NewTrack(
		m.SpotifyID,
		m.Name,
		m.Artist,
		m.Album,
		m.ReleaseDate,
		m.Popularity,
		audioFeatures,
		m.PreviewURL,
		m.ImageURL,
	)

	track.ID = m.ID
	track.CreatedAt = m.CreatedAt
	track.UpdatedAt = m.UpdatedAt

	return track, nil
}

func fromTrackEntity(track *entity.Track) (*trackModel, error) {
	audioFeaturesJSON, err := json.Marshal(track.AudioFeatures)
	if err != nil {
		return nil, err
	}

	return &trackModel{
		ID:            track.ID,
		SpotifyID:     track.SpotifyID,
		Name:          track.Name,
		Artist:        track.Artist,
		Album:         track.Album,
		ReleaseDate:   track.ReleaseDate,
		Popularity:    track.Popularity,
		AudioFeatures: audioFeaturesJSON,
		PreviewURL:    track.PreviewURL,
		ImageURL:      track.ImageURL,
		CreatedAt:     track.CreatedAt,
		UpdatedAt:     track.UpdatedAt,
	}, nil
}

func (r *TrackRepository) GetByID(ctx context.Context, id string) (*entity.Track, error) {
	query := `SELECT *FROM tracks WHERE id=$1`

	var model trackModel
	err := r.db.GetContext(ctx, &model, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("track not found: %w", err)
		}
		return nil, err
	}
	return model.ToEntity()
}

func (r *TrackRepository) GetBySpotifyID(ctx context.Context, spotifyID string) (*entity.Track, error) {
	query := `SELECT *FROM tracks WHERE spotify_id  =%1`

	var model trackModel
	err := r.db.GetContext(ctx, &model, query, spotifyID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("track not found: %w", err)
		}
		return nil, err
	}
	return model.ToEntity()
}

func (r *TrackRepository) Save(ctx context.Context, track *entity.Track) error {
	if track.ID == "" {
		track.ID = uuid.New().String()
	}
	model, err := fromTrackEntity(track)
	if err != nil {
		return err
	}
	query := `INSERT INTO track (id, spotify_id, name, artist, album, release_date, popularity,
			audio_features, preview_url, image_url, created_at, updated_at)
			values (:id, :spotify_id, :name, :artist, :album, :release_date, :popularity,
			:audio_features, :preview_url, :image_url, :created_at, :updated_at)`

	_, err = r.db.NamedExecContext(ctx, query, model)
	return err
}

func (r *TrackRepository) Update(ctx context.Context, track *entity.Track) error {
	track.UpdatedAt = time.Now()

	model, err := fromTrackEntity(track)
	if err != nil {
		return err
	}

	query := `
		UPDATE tracks SET
			spotify_id = :spotify_id,
			name = :name,
			artist = :artist,
			album = :album,
			release_date = :release_date,
			popularity = :popularity,
			audio_features = :audio_features,
			preview_url = :preview_url,
			image_url = :image_url,
			updated_at = :updated_at
		WHERE id = :id`

	result, err := r.db.NamedExecContext(ctx, query, model)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("track with ID %s not found", track.ID)
	}

	return nil
}

func (r *TrackRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM tracks WHERE id=$1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("no such a track")
	}
	return nil
}

func (r *TrackRepository) FindByArtist(ctx context.Context, artist string) ([]*entity.Track, error) {
	query := `SELECT * FROM tracks
			WHERE artist ILIKE $1
			ORDER BY popularity DESC`

	var models []trackModel
	err := r.db.SelectContext(ctx, &models, query, "%"+artist+"%")
	if err != nil {
		return nil, err
	}
	tracks := make([]*entity.Track, 0, len(models))
	for _, model := range models {
		track, err := model.ToEntity()
		if err != nil {
			continue
		}
		tracks = append(tracks, track)
	}

	return tracks, nil
}

func (r *TrackRepository) FindByName(ctx context.Context, name string) ([]*entity.Track, error) {
	query := `
		SELECT * FROM tracks
		WHERE name ILIKE $1
		ORDER BY popularity DESC`

	var models []trackModel

	err := r.db.SelectContext(ctx, &models, query, "%"+name+"%")
	if err != nil {
		return nil, err
	}

	tracks := make([]*entity.Track, 0, len(models))
	for _, model := range models {
		track, err := model.ToEntity()
		if err != nil {
			continue
		}
		tracks = append(tracks, track)
	}
	return tracks, err
}

func (r *TrackRepository) FindByMood(ctx context.Context, mood valueObject.Mood, limit int) ([]*entity.Track, error) {
	var query string
	var args []interface{}

	switch mood {
	case valueObject.MoodHappy:
		query = `
			SELECT * FROM tracks
			WHERE audio_features->>'valence' > '0.7'
			AND audio_features->>'energy' > '0.5'
			ORDER BY popularity DESC
			LIMIT $1
		`
		args = []interface{}{limit}
	case valueObject.MoodSad:
		query = `
			SELECT * FROM tracks
			WHERE audio_features->>'valence' < '0.4'
			AND audio_features->>'energy' < '0.5'
			ORDER BY popularity DESC
			LIMIT $1
		`
		args = []interface{}{limit}
	case valueObject.MoodEnergetic:
		query = `
			SELECT * FROM tracks
			WHERE audio_features->>'energy' > '0.8'
			AND audio_features->>'tempo' > '120'
			ORDER BY popularity DESC
			LIMIT $1
		`
		args = []interface{}{limit}
	default:
		query = `
			SELECT * FROM tracks
			ORDER BY popularity DESC
			LIMIT $1
		`
		args = []interface{}{limit}
	}

	var models []trackModel
	err := r.db.SelectContext(ctx, &models, query, args...)
	if err != nil {
		return nil, err
	}

	tracks := make([]*entity.Track, 0, len(models))
	for _, model := range models {
		track, err := model.ToEntity()
		if err != nil {
			continue
		}
		tracks = append(tracks, track)
	}

	return tracks, nil
}

func (r *TrackRepository) FindByWeather(ctx context.Context, weather valueObject.Weather, limit int) ([]*entity.Track, error) {
	var query string
	var args []interface{}

	switch weather {
	case valueObject.WeatherRainy:
		query = `
			SELECT * FROM tracks
			WHERE audio_features->>'valence' < '0.5'
			AND audio_features->>'acousticness' > '0.5'
			ORDER BY popularity DESC
			LIMIT $1
		`
		args = []interface{}{limit}
	default:
		query = `
			SELECT * FROM tracks
			ORDER BY popularity DESC
			LIMIT $1
		`
		args = []interface{}{limit}
	}

	var models []*trackModel

	err := r.db.SelectContext(ctx, models, query, args...)
	if err != nil {
		return nil, err
	}

	tracks := make([]*entity.Track, 0, len(models))

	for _, model := range models {
		track, err := model.ToEntity()
		if err != nil {
			continue
		}
		tracks = append(tracks, track)
	}
	return tracks, err
}

func (r *TrackRepository) FindByTimeOfDay(ctx context.Context, timeOfDay valueobject.TimeOfDay, limit int) ([]*entity.Track, error) {

	var query string
	var args []interface{}

	switch timeOfDay {
	case valueObject.TimeOfDayMorning:
		query = `
			SELECT * FROM tracks
			WHERE audio_features->>'valence' > '0.5'
			AND audio_features->>'energy' > '0.5'
			AND audio_features->>'energy' < '0.8'
			ORDER BY popularity DESC
			LIMIT $1
		`
		args = []interface{}{limit}
	default:
		query = `
			SELECT * FROM tracks
			ORDER BY popularity DESC
			LIMIT $1
		`
		args = []interface{}{limit}
	}

	var models []trackModel
	err := r.db.SelectContext(ctx, &models, query, args...)
	if err != nil {
		return nil, err
	}

	tracks := make([]*entity.Track, 0, len(models))
	for _, model := range models {
		track, err := model.ToEntity()
		if err != nil {
			continue
		}
		tracks = append(tracks, track)
	}

	return tracks, nil
}

func (r *TrackRepository) FindByMoodWeatherTime(
	ctx context.Context,
	mood valueObject.Mood,
	weather valueObject.Weather,
	timeOfDay valueObject.TimeOfDay,
	limit int,
) ([]*entity.Track, error) {

	moodTracks, err := r.FindByMood(ctx, mood, limit*2)
	if err != nil {
		return nil, err
	}

	weatherTracks, err := r.FindByWeather(ctx, weather, limit*2)
	if err != nil {
		return nil, err
	}

	timeTracks, err := r.FindByTimeOfDay(ctx, timeOfDay, limit*2)
	if err != nil {
		return nil, err
	}

	trackScores := make(map[string]int)
	trackMap := make(map[string]*entity.Track)

	for _, track := range moodTracks {
		trackScores[track.ID] += 3
		trackMap[track.ID] = track
	}

	for _, track := range weatherTracks {
		trackScores[track.ID] += 2
		trackMap[track.ID] = track
	}

	for _, track := range timeTracks {
		trackScores[track.ID] += 1
		trackMap[track.ID] = track
	}

	type trackScore struct {
		track *entity.Track
		score int
	}

	var scoredTracks []trackScore
	for id, score := range trackScores {
		scoredTracks = append(scoredTracks, trackScore{
			track: trackMap[id],
			score: score,
		})
	}

	sort.Slice(scoredTracks, func(i, j int) bool {
		if scoredTracks[i].score == scoredTracks[j].score {
			return scoredTracks[i].track.Popularity > scoredTracks[j].track.Popularity
		}
		return scoredTracks[i].score > scoredTracks[j].score
	})

	result := make([]*entity.Track, 0, limit)
	for i := 0; i < len(scoredTracks) && i < limit; i++ {
		result = append(result, scoredTracks[i].track)
	}

	return result, nil
}

func (r *TrackRepository) GetPopularTracks(ctx context.Context, limit int) ([]*entity.Track, error) {
	query := `
        SELECT * FROM tracks
        ORDER BY popularity DESC
        LIMIT $1
    `

	rows, err := r.db.QueryxContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get popular tracks: %w", err)
	}
	defer rows.Close()

	var tracks []*entity.Track

	for rows.Next() {
		var model trackModel

		if err := rows.StructScan(&model); err != nil {
			return nil, fmt.Errorf("failed to scan track model: %w", err)
		}

		track, err := model.ToEntity()
		if err != nil {
			return nil, err
		}

		tracks = append(tracks, track)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating through tracks: %w", err)
	}

	return tracks, nil
}
