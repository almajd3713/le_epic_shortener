package repository

import (
	"context"
	"log/slog"
	"time"
	
	"github.com/jackc/pgx/v5/pgxpool"
	"shortener.reeler.com/backend/internal/models"
)

type IClickRepository interface {
	Create(code string) (*models.Click, error)
	GetByCode(code string) ([]*models.Click, error)
	GetCountByCode(code string) (int, error)
}

type ClickRepository struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func NewClickRepository(pool *pgxpool.Pool) *ClickRepository {
	return &ClickRepository{pool: pool}
}

func (r *ClickRepository) Create(code string) (*models.Click, error) {
	var click models.Click
	err := r.pool.QueryRow(
		context.Background(),
		`INSERT INTO clicks (short_code, clicked_at)
		 VALUES ($1, $2)
		 RETURNING id, short_code, clicked_at`,
		code, time.Now(),
	).Scan(&click.ID, &click.ShortCode, &click.ClickedAt)
	if err != nil {
		r.logger.Error("failed to insert click", "error", err)
		return nil, err
	}
	return &click, nil
}

func (r *ClickRepository) GetByCode(code string) ([]*models.Click, error) {
	rows, err := r.pool.Query(
		context.Background(),
		`SELECT id, short_code, clicked_at
		 FROM clicks
		 WHERE short_code = $1
		 ORDER BY clicked_at DESC`,
		code,
	)
	if err != nil {
		r.logger.Error("failed to query clicks by code", "error", err)
		return nil, err
	}
	defer rows.Close()

	var clicks []*models.Click
	for rows.Next() {
		var click models.Click
		err := rows.Scan(&click.ID, &click.ShortCode, &click.ClickedAt)
		if err != nil {
			r.logger.Error("failed to scan click row", "error", err)
			return nil, err
		}
		clicks = append(clicks, &click)
	}
	return clicks, nil
}

func (r *ClickRepository) GetCountByCode(code string) (int, error) {
	row := r.pool.QueryRow(
		context.Background(),
		`SELECT COUNT(*)
		 FROM clicks
		 WHERE short_code = $1`,
		code,
	)
	var count int
	err := row.Scan(&count)
	if err != nil {
		r.logger.Error("failed to count clicks by code", "error", err)
		return 0, err
	}
	return count, nil
}