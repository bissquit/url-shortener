package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bissquit/url-shortener/internal/repository"
	"github.com/jackc/pgerrcode"
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
		"INSERT INTO urls (short_id, original_url) VALUES ($1, $2)",
		id, originalURL,
	)
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		if pgErr.ConstraintName == "idx_original_url" {
			return fmt.Errorf("%w: %s", repository.ErrURLAlreadyExists, originalURL)
		}
		// UNIQUE/PK by short_id
		return fmt.Errorf("%w: %s", repository.ErrIDAlreadyExists, id)
	}

	return err
}

func (s *PGStorage) CreateBatch(items []repository.URLItem) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, item := range items {
		if item.ID == "" {
			return fmt.Errorf("%w", repository.ErrEmptyID)
		}

		_, err = tx.Exec(ctx,
			"INSERT INTO urls (short_id, original_url) VALUES ($1, $2)",
			item.ID, item.OriginalURL,
		)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
				if pgErr.ConstraintName == "idx_original_url" {
					return fmt.Errorf("%w: %s", repository.ErrURLAlreadyExists, item.OriginalURL)
				}
				return fmt.Errorf("%w: %s", repository.ErrIDAlreadyExists, item.ID)
			}
			return err
		}
	}

	return tx.Commit(ctx)
}

func (s *PGStorage) GetURLByID(id string) (string, error) {
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

func (s *PGStorage) GetIDByURL(url string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	raw := s.pool.QueryRow(ctx,
		"SELECT short_id FROM urls WHERE original_url = $1", url)

	var id string
	err := raw.Scan(&id)

	if err == nil {
		return id, nil
	}

	if err == pgx.ErrNoRows {
		return "", repository.ErrNotFound
	}

	return "", err
}
