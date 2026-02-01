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
	"github.com/lib/pq"
)

type PGStorage struct {
	pool *pgxpool.Pool
}

func NewDBStorage(p *pgxpool.Pool) *PGStorage {
	return &PGStorage{
		pool: p,
	}
}

func (s *PGStorage) Create(id string, originalURL, userID string) error {
	if id == "" {
		return fmt.Errorf("%w", repository.ErrEmptyID)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := s.pool.Exec(ctx,
		"INSERT INTO urls (short_id, original_url, user_id) VALUES ($1, $2, $3)",
		id, originalURL, userID,
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

func (s *PGStorage) CreateBatch(items []repository.URLItem, userID string) error {
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
			"INSERT INTO urls (short_id, original_url, user_id) VALUES ($1, $2, $3)",
			item.ID, item.OriginalURL, userID,
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

	row := s.pool.QueryRow(ctx,
		"SELECT original_url, is_deleted FROM urls WHERE short_id = $1", id)

	var originalURL string
	var deleted bool
	err := row.Scan(&originalURL, &deleted)
	if err == pgx.ErrNoRows {
		return "", repository.ErrNotFound
	}
	if err != nil {
		return "", err
	}
	if deleted {
		return "", repository.ErrDeleted
	}

	return originalURL, nil
}

func (s *PGStorage) GetIDByURL(url string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	row := s.pool.QueryRow(ctx,
		"SELECT short_id, is_deleted FROM urls WHERE original_url = $1", url)

	var id string
	var deleted bool
	err := row.Scan(&id, &deleted)
	if err == pgx.ErrNoRows {
		return "", repository.ErrNotFound
	}
	if err != nil {
		return "", err
	}
	if deleted {
		return "", repository.ErrDeleted
	}

	return id, nil
}

func (s *PGStorage) GetURLsByUserID(userID string) ([]repository.UserURL, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := s.pool.Query(ctx,
		"SELECT short_id, original_url, is_deleted FROM urls WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []repository.UserURL
	for rows.Next() {
		var shortID, originalURL string
		var deleted bool
		if err = rows.Scan(&shortID, &originalURL, &deleted); err != nil {
			return nil, err
		}
		if !deleted {
			items = append(items, repository.UserURL{
				ShortID:     shortID,
				OriginalURL: originalURL,
			})
		}
	}

	return items, rows.Err()
}

func (s *PGStorage) DeleteBatch(userID string, ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := s.pool.Exec(ctx,
		"UPDATE urls SET is_deleted = TRUE WHERE user_id = $1 AND short_id = ANY($2)",
		userID, pq.Array(ids))
	return err
}
