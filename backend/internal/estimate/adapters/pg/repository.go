// Package pg — PostgreSQL-адаптер контекста Estimate.
package pg

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zetta/zettacalc/backend/internal/estimate/domain"
)

// Repository читает сметы.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository создаёт репозиторий смет.
func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

// PayloadByID отдаёт смету теми же байтами, что и mock API.
func (r *Repository) PayloadByID(ctx context.Context, id string) ([]byte, error) {
	var payload []byte
	err := r.pool.QueryRow(ctx, `SELECT payload_text FROM estimate WHERE id = $1`, id).Scan(&payload)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return payload, err
}

// ByID возвращает смету со строками и допущениями.
func (r *Repository) ByID(ctx context.Context, id string) (domain.Estimate, error) {
	var e domain.Estimate
	err := r.pool.QueryRow(ctx,
		`SELECT id, status, total_rub, range_min, range_max, days,
		        COALESCE(variant_id, ''), COALESCE(variant_name, ''), COALESCE(project_id, '')
		   FROM estimate WHERE id = $1`, id).
		Scan(&e.ID, &e.Status, &e.TotalRub, &e.RangeMin, &e.RangeMax, &e.Days,
			&e.VariantID, &e.VariantName, &e.ProjectID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Estimate{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Estimate{}, err
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, title, qty, unit, price, COALESCE(sku, ''), COALESCE(brand, '')
		   FROM estimate_line WHERE estimate_id = $1 ORDER BY position`, id)
	if err != nil {
		return domain.Estimate{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var l domain.Line
		if err := rows.Scan(&l.ID, &l.Title, &l.Qty, &l.Unit, &l.Price, &l.SKU, &l.Brand); err != nil {
			return domain.Estimate{}, err
		}
		e.Lines = append(e.Lines, l)
	}
	if err := rows.Err(); err != nil {
		return domain.Estimate{}, err
	}

	assumptionRows, err := r.pool.Query(ctx,
		`SELECT text FROM estimate_assumption WHERE estimate_id = $1 ORDER BY position`, id)
	if err != nil {
		return domain.Estimate{}, err
	}
	defer assumptionRows.Close()
	for assumptionRows.Next() {
		var text string
		if err := assumptionRows.Scan(&text); err != nil {
			return domain.Estimate{}, err
		}
		e.Assumptions = append(e.Assumptions, text)
	}
	return e, assumptionRows.Err()
}

// Save сохраняет сметы вместе с их JSON-представлением.
func (r *Repository) Save(ctx context.Context, estimates []domain.Estimate, payloads map[string][]byte) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `DELETE FROM estimate`); err != nil {
		return err
	}
	for _, e := range estimates {
		var variantID any
		if e.VariantID != "" {
			variantID = e.VariantID
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO estimate (id, status, total_rub, range_min, range_max, days,
			                       variant_id, variant_name, project_id, payload_text)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
			e.ID, e.Status, e.TotalRub, e.RangeMin, e.RangeMax, e.Days,
			variantID, nullStr(e.VariantName), nullStr(e.ProjectID), string(payloads[e.ID])); err != nil {
			return err
		}
		for i, l := range e.Lines {
			if _, err := tx.Exec(ctx,
				`INSERT INTO estimate_line (estimate_id, id, position, title, qty, unit, price, sku, brand)
				 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
				e.ID, l.ID, i, l.Title, l.Qty, l.Unit, l.Price,
				nullStr(l.SKU), nullStr(l.Brand)); err != nil {
				return err
			}
		}
		for i, a := range e.Assumptions {
			if _, err := tx.Exec(ctx,
				`INSERT INTO estimate_assumption (estimate_id, position, text) VALUES ($1,$2,$3)`,
				e.ID, i, a); err != nil {
				return err
			}
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
