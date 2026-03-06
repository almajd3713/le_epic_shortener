package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"shortener.reeler.com/backend/internal/models"
)

type URLRepository struct {
	pool *pgxpool.Pool
}

func NewURLRepository(pool *pgxpool.Pool) *URLRepository {
	return &URLRepository{pool: pool}
}

func (r *URLRepository) Create(shortCode, longURL string, expiresAt *time.Time) (*models.URL, error) {
	row := r.pool.QueryRow(
		context.Background(),
		`INSERT INTO urls (short_code, long_url, expires_at)
		 VALUES ($1, $2, $3)
		 RETURNING id, short_code, long_url, created_at, expires_at, is_active`,
		shortCode, longURL, expiresAt,
	)
	return scanURL(row)
}

func (r *URLRepository) GetByCode(code string) (*models.URL, error) {
	row := r.pool.QueryRow(
		context.Background(),
		`SELECT id, short_code, long_url, created_at, expires_at, is_active
		 FROM urls
		 WHERE short_code = $1
		   AND is_active = TRUE
		   AND (expires_at IS NULL OR expires_at > NOW())`,
		code,
	)
	u, err := scanURL(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("URL not found or expired")
	}
	return u, err
}

// Deactivates a short code (e.g. once expired)
func (r *URLRepository) Deactivate(code string) error {
	tag, err := r.pool.Exec(
		context.Background(),
		`UPDATE urls SET is_active = FALSE
		 WHERE short_code = $1 AND is_active = TRUE`,
		code,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("URL not found or already deactivated")
	}
	return nil
}

// scanURL reads a single urls row into a models.URL.
func scanURL(row pgx.Row) (*models.URL, error) {
	var u models.URL
	err := row.Scan(
		&u.ID,
		&u.ShortCode,
		&u.LongURL,
		&u.CreatedAt,
		&u.ExpiresAt,
		&u.IsActive,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}
