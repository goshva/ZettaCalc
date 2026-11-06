// Package pg — PostgreSQL-адаптер контекста Techcard.
package pg

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zetta/zettacalc/backend/internal/techcard/domain"
)

// Repository читает техкарты.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository создаёт репозиторий техкарт.
func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

// List возвращает техкарты, при необходимости только по поверхности.
func (r *Repository) List(ctx context.Context, surface string) ([]domain.Techcard, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, code, title, surface, material, version, min_area_m2, max_unevenness,
		        sort_order, payload_text
		   FROM techcard
		  WHERE ($1 = '' OR surface = $1)
		  ORDER BY sort_order`, surface)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.Techcard, 0, 4)
	for rows.Next() {
		var c domain.Techcard
		if err := rows.Scan(&c.ID, &c.Code, &c.Title, &c.Surface, &c.Material, &c.Version,
			&c.MinAreaM2, &c.MaxUnevenness, &c.Order, &c.Payload); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// Save сохраняет техкарты.
func (r *Repository) Save(ctx context.Context, cards []domain.Techcard) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `DELETE FROM techcard`); err != nil {
		return err
	}
	for _, c := range cards {
		if _, err := tx.Exec(ctx,
			`INSERT INTO techcard (id, code, title, surface, material, version, min_area_m2,
			                       max_unevenness, sort_order, payload, payload_text)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10::jsonb,$11)`,
			c.ID, c.Code, c.Title, c.Surface, c.Material, c.Version, c.MinAreaM2,
			c.MaxUnevenness, c.Order, string(c.Payload), string(c.Payload)); err != nil {
			return err
		}
		for i, op := range c.Operations {
			if _, err := tx.Exec(ctx,
				`INSERT INTO techcard_operation (techcard_id, position, name) VALUES ($1,$2,$3)`,
				c.ID, i, op); err != nil {
				return err
			}
		}
		for name, value := range c.Consumption {
			if _, err := tx.Exec(ctx,
				`INSERT INTO techcard_consumption (techcard_id, name, value) VALUES ($1,$2,$3)
				 ON CONFLICT DO NOTHING`, c.ID, name, value); err != nil {
				return err
			}
		}
	}
	return tx.Commit(ctx)
}
