// Package pg — PostgreSQL-адаптер контекста Tender.
package pg

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zetta/zettacalc/backend/internal/tender/domain"
)

// Repository читает ставки подрядчиков.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository создаёт репозиторий ставок.
func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

// List возвращает ставки; при variantID = "" — только статичные.
func (r *Repository) List(ctx context.Context, variantID string) ([]domain.Bid, error) {
	const q = `SELECT id, name, rating, price, days, COALESCE(scope, ''), COALESCE(image_key, ''),
	                  COALESCE(variant_id, ''), sort_order, payload_text
	             FROM contractor_bid
	            WHERE ($1 = '' AND variant_id IS NULL) OR ($1 <> '' AND variant_id = $1)
	            ORDER BY sort_order`
	rows, err := r.pool.Query(ctx, q, variantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.Bid, 0, 4)
	for rows.Next() {
		var b domain.Bid
		if err := rows.Scan(&b.ID, &b.Name, &b.Rating, &b.Price, &b.Days, &b.Scope,
			&b.ImageKey, &b.VariantID, &b.Order, &b.Payload); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

// Save сохраняет ставки.
func (r *Repository) Save(ctx context.Context, bids []domain.Bid) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `DELETE FROM contractor_bid`); err != nil {
		return err
	}
	for _, b := range bids {
		var variant any
		if b.VariantID != "" {
			variant = b.VariantID
		}
		var imageKey any
		if b.ImageKey != "" {
			imageKey = b.ImageKey
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO contractor_bid (id, name, rating, price, days, scope, image_key, variant_id,
			                             sort_order, payload, payload_text)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10::jsonb,$11)`,
			b.ID, b.Name, b.Rating, b.Price, b.Days, nullStr(b.Scope), imageKey, variant,
			b.Order, string(b.Payload), string(b.Payload)); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}
