package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bissquit/url-shortener/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PGStorage struct {
	pool *pgxpool.Pool
}

func NewDBStorage(p *pgxpool.Pool) *PGStorage {
	return &PGStorage{
		pool: p,
	}
}

func (s *PGStorage) Create(id string, originalURL string) error {
	if id == "" {
		return fmt.Errorf("%w", repository.ErrEmptyID)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := s.pool.Exec(ctx,
		"INSERT INTO urls (short_id, original_url) VALUES ($1, $2)", id, originalURL)
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return fmt.Errorf("%w: %s", repository.ErrAlreadyExists, id)
	}

	return err
}

func (s *PGStorage) Get(id string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	raw := s.pool.QueryRow(ctx,
		"SELECT original_url FROM urls WHERE short_id = $1", id)

	var url string
	err := raw.Scan(&url)

	if err == nil {
		return url, nil
	}

	if err == pgx.ErrNoRows {
		return "", repository.ErrNotFound
	}

	return "", err
}
