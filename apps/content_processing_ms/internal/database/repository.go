package database

import (
	"context"
	"fmt"

	"github.com/ecbDeveloper/netflix-architecture/apps/content_processing_ms/internal/config"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ContentType string

const (
	TypeMovie   ContentType = "movie"
	TypeEpisode ContentType = "episode"
)

type PendingContent struct {
	ID   uuid.UUID
	Type ContentType
}

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(ctx context.Context, cfg *config.Config) (*Repository, error) {
	pool, err := pgxpool.New(ctx, cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &Repository{pool: pool}, nil
}

func (r *Repository) Close() {
	r.pool.Close()
}

func (r *Repository) GetNextPendingContent(ctx context.Context) (*PendingContent, error) {
	movieQuery := `
		SELECT content_id 
		FROM movies 
		WHERE status = 'PENDING' 
		   OR (status = 'PROCESSING' AND updated_at < NOW() - INTERVAL '2 hours')
		LIMIT 1 FOR UPDATE SKIP LOCKED;
	`
	var movieID uuid.UUID
	err := r.pool.QueryRow(ctx, movieQuery).Scan(&movieID)
	if err == nil {
		return &PendingContent{ID: movieID, Type: TypeMovie}, nil
	}

	episodeQuery := `
		SELECT id 
		FROM episodes 
		WHERE status = 'PENDING' 
		   OR (status = 'PROCESSING' AND updated_at < NOW() - INTERVAL '2 hours')
		LIMIT 1 FOR UPDATE SKIP LOCKED;
	`
	var episodeID uuid.UUID
	err = r.pool.QueryRow(ctx, episodeQuery).Scan(&episodeID)
	if err == nil {
		return &PendingContent{ID: episodeID, Type: TypeEpisode}, nil
	}

	return nil, nil
}

func (r *Repository) UpdateContent(ctx context.Context, id uuid.UUID, contentType ContentType, status string, url *string, durationSeconds int) error {
	var query string

	switch contentType {
	case TypeMovie:
		query = `
			UPDATE movies 
			SET status = $1, content_url = $2, duration_seconds = $3, updated_at = NOW() 
			WHERE content_id = $4;
		`
	case TypeEpisode:
		query = `
			UPDATE episodes 
			SET status = $1, content_url = $2, duration_seconds = $3, updated_at = NOW() 
			WHERE id = $4;
		`
	default:
		return fmt.Errorf("invalid content type")
	}

	_, err := r.pool.Exec(ctx, query, status, url, durationSeconds, id)
	return err
}

func (r *Repository) GetContentStatus(ctx context.Context, id uuid.UUID, contentType ContentType) (string, error) {
	var query string
	var status string

	switch contentType {
	case TypeMovie:
		query = "SELECT status FROM movies WHERE content_id = $1"
	case TypeEpisode:
		query = "SELECT status FROM episodes WHERE id = $1"
	default:
		return "", fmt.Errorf("invalid content type")
	}

	err := r.pool.QueryRow(ctx, query, id).Scan(&status)
	if err != nil {
		return "", err
	}
	return status, nil
}
