// Package pg — PostgreSQL-адаптер контекста Design.
package pg

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zetta/zettacalc/backend/internal/design/domain"
)

// Repository читает варианты дизайна.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository создаёт репозиторий вариантов.
func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

// VariantsPayload возвращает готовый JSON-массив вариантов.
//
// Паритет с mock API требует отдать тот же массив, что и Python, поэтому
// склеиваются сохранённые представления, а не собирается новый JSON.
func (r *Repository) VariantsPayload(ctx context.Context) ([]byte, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT payload_text FROM design_variant ORDER BY sort_order`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []byte{'['}
	first := true
	for rows.Next() {
		var payload []byte
		if err := rows.Scan(&payload); err != nil {
			return nil, err
		}
		if !first {
			out = append(out, ',')
		}
		first = false
		out = append(out, payload...)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return append(out, ']'), nil
}

// Variants возвращает все варианты по порядку.
func (r *Repository) Variants(ctx context.Context) ([]domain.Variant, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, style, palette, image_path, description, budget_min, budget_max,
		        estimated_days, sort_order
		   FROM design_variant ORDER BY sort_order`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.Variant, 0, 4)
	for rows.Next() {
		var v domain.Variant
		if err := rows.Scan(&v.ID, &v.Name, &v.Style, &v.Palette, &v.ImagePath, &v.Description,
			&v.BudgetMin, &v.BudgetMax, &v.EstimatedDays, &v.Order); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	rooms, err := r.roomsByVariant(ctx)
	if err != nil {
		return nil, err
	}
	for i := range out {
		out[i].Rooms = rooms[out[i].ID]
	}
	return out, nil
}

// VariantByID возвращает вариант по идентификатору.
func (r *Repository) VariantByID(ctx context.Context, id string) (domain.Variant, error) {
	var v domain.Variant
	err := r.pool.QueryRow(ctx,
		`SELECT id, name, style, palette, image_path, description, budget_min, budget_max,
		        estimated_days, sort_order
		   FROM design_variant WHERE id = $1`, id).
		Scan(&v.ID, &v.Name, &v.Style, &v.Palette, &v.ImagePath, &v.Description,
			&v.BudgetMin, &v.BudgetMax, &v.EstimatedDays, &v.Order)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Variant{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Variant{}, err
	}
	rooms, err := r.roomsByVariant(ctx)
	if err != nil {
		return domain.Variant{}, err
	}
	v.Rooms = rooms[v.ID]
	return v, nil
}

func (r *Repository) roomsByVariant(ctx context.Context) (map[string][]string, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT variant_id, room FROM design_variant_room ORDER BY variant_id, position`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[string][]string, 4)
	for rows.Next() {
		var id, room string
		if err := rows.Scan(&id, &room); err != nil {
			return nil, err
		}
		out[id] = append(out[id], room)
	}
	return out, rows.Err()
}

// Save сохраняет варианты дизайна вместе с их JSON-представлением.
func (r *Repository) Save(ctx context.Context, variants []domain.Variant, payloads map[string][]byte) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `DELETE FROM design_variant`); err != nil {
		return err
	}
	for _, v := range variants {
		var imageFile any
		if file := FilenameFromPath(v.ImagePath); file != "" {
			imageFile = file
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO design_variant (id, name, style, palette, image_path, image_file, description,
			                             budget_min, budget_max, estimated_days, sort_order, payload_text)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
			v.ID, v.Name, v.Style, v.Palette, v.ImagePath, imageFile, v.Description,
			v.BudgetMin, v.BudgetMax, v.EstimatedDays, v.Order, string(payloads[v.ID])); err != nil {
			return err
		}
		for i, room := range v.Rooms {
			if _, err := tx.Exec(ctx,
				`INSERT INTO design_variant_room (variant_id, position, room) VALUES ($1,$2,$3)`,
				v.ID, i, room); err != nil {
				return err
			}
		}
	}
	return tx.Commit(ctx)
}

// FilenameFromPath вытаскивает имя файла из пути /api/images/file/<name>.
func FilenameFromPath(path string) string {
	const marker = "/api/images/file/"
	if len(path) <= len(marker) || path[:len(marker)] != marker {
		return ""
	}
	return path[len(marker):]
}
