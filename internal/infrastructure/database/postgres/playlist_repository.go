package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"spotify_recommender/internal/domain/entity"
	"spotify_recommender/internal/domain/valueObject"
	"time"
)

type PlaylistRepository struct {
	db *sqlx.DB
}

func NewPlaylistRepository(db *sqlx.DB) *PlaylistRepository {
	return &PlaylistRepository{
		db: db,
	}
}
func fromPlaylistEntity(playlist *entity.Playlist) *playlistModel {
	model := &playlistModel{
		ID:          playlist.ID,
		UserID:      playlist.UserID,
		Name:        playlist.Name,
		Description: playlist.Description,
		Mood:        string(playlist.Mood),
		IsPublic:    playlist.IsPublic,
		CreatedAt:   playlist.CreatedAt,
		UpdatedAt:   playlist.UpdatedAt,
	}

	if string(playlist.Weather) != "" {
		model.Weather = sql.NullString{
			String: string(playlist.Weather),
			Valid:  true,
		}
	}

	if string(playlist.TimeOfDay) != "" {
		model.TimeOfDay = sql.NullString{
			String: string(playlist.TimeOfDay),
			Valid:  true,
		}
	}

	return model
}

type playlistModel struct {
	ID          string         `db:"id"`
	UserID      string         `db:"user_id"`
	Name        string         `db:"name"`
	Description string         `db:"description"`
	Mood        string         `db:"mood"`
	Weather     sql.NullString `db:"weather"`
	TimeOfDay   sql.NullString `db:"time_of_day"`
	IsPublic    bool           `db:"is_public"`
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at"`
}

func (m *playlistModel) toEntity() (*entity.Playlist, error) {
	playlist := entity.NewPlaylist(m.UserID,
		m.Name,
		m.Description,
		valueObject.Mood(m.Mood))

	playlist.ID = m.ID
	if m.Weather.Valid {
		playlist.Weather = valueObject.Weather(m.Weather.String)
	}

	if m.TimeOfDay.Valid {
		playlist.TimeOfDay = valueObject.TimeOfDay(m.TimeOfDay.String)
	}
	playlist.IsPublic = m.IsPublic
	playlist.CreatedAt = m.CreatedAt
	playlist.UpdatedAt = m.UpdatedAt

	return playlist, nil
}

func (r *PlaylistRepository) GetByID(ctx context.Context, id string) (*entity.Playlist, error) {
	query := `
		SELECT * FROM playlists
		WHERE id = $1
	`

	var model playlistModel
	err := r.db.GetContext(ctx, &model, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("playlist not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get playlist by ID: %w", err)
	}

	playlist, err := model.toEntity()
	if err != nil {
		return nil, err
	}
	tracks, err := r.getPlaylistTracksIDs(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get playlist tracks: %w", err)
	}

	playlist.Tracks = tracks

	return playlist, nil
}

func (r *PlaylistRepository) getPlaylistTracksIDs(ctx context.Context, playlistID string) ([]string, error) {
	query := `SELECT track_id FROM playlist_tracks
WHERE playlist_id=$1
ORDER BY POSITION`
	var trackIDs []string
	err := r.db.SelectContext(ctx, &trackIDs, query, playlistID)
	if err != nil {
		return nil, fmt.Errorf("failed to get playlist tracks: %w", err)
	}

	return trackIDs, nil
}

func (r *PlaylistRepository) Save(ctx context.Context, playlist *entity.Playlist) error {
	if playlist.ID == "" {
		playlist.ID = uuid.New().String()
	}
	now := time.Now()
	if playlist.CreatedAt.IsZero() {
		playlist.CreatedAt = now
	}
	if playlist.UpdatedAt.IsZero() {
		playlist.CreatedAt = now
	}
	model := fromPlaylistEntity(playlist)
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	query := `INSERT INTO playlists (	id, user_id, name, description, mood, weather, time_of_day,
			is_public, created_at, updated_at) VALUES
			(	:id, :user_id, :name, :description, :mood, :weather, :time_of_day,
			:is_public, :created_at, :updated_at)`

	_, err = tx.NamedExecContext(ctx, query, model)
	if err != nil {
		return fmt.Errorf("failed to save playlist: %w", err)
	}

	if len(playlist.Tracks) > 0 {
		insertQuery := ` INSERT INTO playlist_tracks (playlist_id, track_id, position)
			VALUES ($1, $2, $3)`
		for i, trackID := range playlist.Tracks {
			_, err = tx.ExecContext(ctx, insertQuery, playlist.Name, trackID, i)
			if err != nil {
				return fmt.Errorf("failed to add track to playlist: %w", err)
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)

	}
	return nil
}

func (r *PlaylistRepository) Update(ctx context.Context, playlist *entity.Playlist) error {
	playlist.UpdatedAt = time.Now()
	model := fromPlaylistEntity(playlist)
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	query := `
		UPDATE playlists SET
			user_id = :user_id,
			name = :name,
			description = :description,
			mood = :mood,
			weather = :weather,
			time_of_day = :time_of_day,
			is_public = :is_public,
			updated_at = :updated_at
		WHERE id = :id`

	result, err := tx.NamedExecContext(ctx, query, model)
	if err != nil {
		return fmt.Errorf("failed to update playlist: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("playlist with ID %s not found", playlist.ID)
	}
	deleteQuery := `DELETE FROM playlist_tracks WHERE playlist_id = $1`
	_, err = tx.ExecContext(ctx, deleteQuery, playlist.ID)
	if err != nil {
		return fmt.Errorf("failed to delete playlist tracks: %w", err)
	}

	if len(playlist.Tracks) > 0 {
		insertQuery := `INSERT INTO playlist_tracks (playlist_id, track_id, position)
		VALUES ($1, $2, $3)`
		for i, trackID := range playlist.Tracks {
			_, err = tx.ExecContext(ctx, insertQuery, playlist.ID, trackID, i)
			if err != nil {
				return fmt.Errorf("failed to add track to playlist: %w", err)
			}
		}
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *PlaylistRepository) Delete(ctx context.Context, id string) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	deleteTracksQuery := `DELETE FROM playlist_tracks WHERE playlist_id = $1`
	_, err = tx.ExecContext(ctx, deleteTracksQuery, id)
	if err != nil {
		return fmt.Errorf("failed to delete playlist tracks: %w", err)
	}

	deletePlaylistQuery := `DELETE FROM playlists WHERE id = $1`
	result, err := tx.ExecContext(ctx, deletePlaylistQuery, id)
	if err != nil {
		return fmt.Errorf("failed to delete playlist: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("playlist with ID %s not found", id)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *PlaylistRepository) GetUserPlaylists(ctx context.Context, userID string) ([]*entity.Playlist, error) {
	query := `
		SELECT * FROM playlists
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.QueryxContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user playlists: %w", err)
	}
	defer rows.Close()

	var playlists []*entity.Playlist
	for rows.Next() {
		var model playlistModel
		if err := rows.StructScan(&model); err != nil {
			return nil, fmt.Errorf("failed to scan playlist model: %w", err)
		}
		playlist, err := model.toEntity()
		if err != nil {
			continue
		}
		tracks, err := r.getPlaylistTracksIDs(ctx, playlist.ID)
		if err != nil {
			continue
		}

		playlist.Tracks = tracks
		playlists = append(playlists, playlist)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating through playlists: %w", err)
	}

	return playlists, nil
}
func (r *PlaylistRepository) FindByName(ctx context.Context, name string) ([]*entity.Playlist, error) {
	query := `
		SELECT * FROM playlists
		WHERE name ILIKE $1 AND is_public = TRUE
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryxContext(ctx, query, "%"+name+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to find playlists by name: %w", err)
	}
	defer rows.Close()

	var playlists []*entity.Playlist

	for rows.Next() {
		var model playlistModel
		if err := rows.StructScan(&model); err != nil {
			return nil, fmt.Errorf("failed to scan playlist model: %w", err)
		}
		playlist, err := model.toEntity()
		if err != nil {
			continue
		}
		tracks, err := r.getPlaylistTracksIDs(ctx, playlist.ID)
		if err != nil {
			continue
		}
		playlist.Tracks = tracks
		playlists = append(playlists, playlist)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating through playlists: %w", err)
	}
	return playlists, nil

}

func (r *PlaylistRepository) FindByMood(ctx context.Context, mood valueObject.Mood) ([]*entity.Playlist, error) {
	query := `
		SELECT * FROM playlists
		WHERE mood = $1 AND is_public = TRUE
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryxContext(ctx, query, string(mood))
	if err != nil {
		return nil, fmt.Errorf("failed to find playlists by mood: %w", err)
	}
	defer rows.Close()

	var playlists []*entity.Playlist

	for rows.Next() {
		var model playlistModel

		if err := rows.StructScan(&model); err != nil {
			return nil, fmt.Errorf("failed to scan playlist model: %w", err)
		}

		playlist, err := model.toEntity()
		if err != nil {
			continue
		}

		tracks, err := r.getPlaylistTracksIDs(ctx, playlist.ID)
		if err != nil {
			continue
		}

		playlist.Tracks = tracks
		playlists = append(playlists, playlist)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating through playlists: %w", err)
	}

	return playlists, nil
}

func (r *PlaylistRepository) AddTrackToPlaylist(ctx context.Context, playlistID, trackID string) error {
	query := `SELECT COALESCE(MAX(position) + 1, 0)
		FROM playlist_tracks
		WHERE playlist_id = $1`
	var position int
	err := r.db.GetContext(ctx, &position, query, playlistID)
	if err != nil {
		return fmt.Errorf("failed to get last track position: %w", err)
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)

	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	insertQuery := `INSERT INTO playlist_tracks (playlist_id, track_id, position) VALUES ($1, $2, $3)
		ON CONFLICT (playlist_id, track_id) DO NOTHING`

	result, err := tx.ExecContext(ctx, insertQuery, playlistID, trackID, position)
	if err != nil {
		return fmt.Errorf("failed to add track to playlist: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		if err = tx.Rollback(); err != nil {
			return fmt.Errorf("failed to rollback transaction: %w", err)
		}
		return nil
	}

	updateQuery := `
		UPDATE playlists SET updated_at = $1
		WHERE id = $2
	`

	_, err = tx.ExecContext(ctx, updateQuery, time.Now(), playlistID)
	if err != nil {
		return fmt.Errorf("failed to update playlist timestamp: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *PlaylistRepository) RemoveTrackFromPlaylist(ctx context.Context, playlistID, trackID string) error {
	query := `
		SELECT position FROM playlist_tracks
		WHERE playlist_id = $1 AND track_id = $2
	`

	var position int
	err := r.db.GetContext(ctx, &position, query, playlistID, trackID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("track not found in playlist")
		}
		return fmt.Errorf("failed to get track position: %w", err)
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	deleteQuery := `
		DELETE FROM playlist_tracks
		WHERE playlist_id = $1 AND track_id = $2
	`

	_, err = tx.ExecContext(ctx, deleteQuery, playlistID, trackID)
	if err != nil {
		return fmt.Errorf("failed to remove track from playlist: %w", err)
	}

	updateQuery := `
		UPDATE playlist_tracks
		SET position = position - 1
		WHERE playlist_id = $1 AND position > $2
	`

	_, err = tx.ExecContext(ctx, updateQuery, playlistID, position)
	if err != nil {
		return fmt.Errorf("failed to update track positions: %w", err)
	}

	updateTimeQuery := `
		UPDATE playlists SET updated_at = $1
		WHERE id = $2
	`

	_, err = tx.ExecContext(ctx, updateTimeQuery, time.Now(), playlistID)
	if err != nil {
		return fmt.Errorf("failed to update playlist timestamp: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
func (r *PlaylistRepository) GetPlaylistTracks(ctx context.Context, playlistID string) ([]*entity.Track, error) {
	query := `
		SELECT t.* FROM tracks t
		JOIN playlist_tracks pt ON t.id = pt.track_id
		WHERE pt.playlist_id = $1
		ORDER BY pt.position
	`

	rows, err := r.db.QueryxContext(ctx, query, playlistID)
	if err != nil {
		return nil, fmt.Errorf("failed to get playlist tracks: %w", err)
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
			continue
		}

		tracks = append(tracks, track)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating through tracks: %w", err)
	}

	return tracks, nil
}

func (r *PlaylistRepository) GetPublicPlaylists(ctx context.Context, limit, offset int) ([]*entity.Playlist, error) {
	query := `
		SELECT * FROM playlists
		WHERE is_public = TRUE
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryxContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get public playlists: %w", err)
	}
	defer rows.Close()

	var playlists []*entity.Playlist

	for rows.Next() {
		var model playlistModel

		if err := rows.StructScan(&model); err != nil {
			return nil, fmt.Errorf("failed to scan playlist model: %w", err)
		}

		playlist, err := model.toEntity()
		if err != nil {
			continue
		}

		tracks, err := r.getPlaylistTracksIDs(ctx, playlist.ID)
		if err != nil {
			continue
		}

		playlist.Tracks = tracks
		playlists = append(playlists, playlist)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating through playlists: %w", err)
	}

	return playlists, nil
}
