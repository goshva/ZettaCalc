// Package pg — PostgreSQL-адаптер контекста Booking.
package pg

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zetta/zettacalc/backend/internal/booking/domain"
)

// Repository хранит записи в PostgreSQL.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository создаёт репозиторий записей.
func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

// Save сохраняет запись.
func (r *Repository) Save(ctx context.Context, a domain.Appointment) error {
	tasks, _ := json.Marshal(a.Request.Tasks)
	ideas, _ := json.Marshal(a.Request.Ideas)
	files, _ := json.Marshal(a.Request.Files)
	validation, _ := json.Marshal(map[string]any{
		"ok":       a.Validation.OK,
		"errors":   a.Validation.Errors,
		"warnings": a.Validation.Warnings,
		"summary":  a.Validation.Summary,
	})

	var preferredDate any
	if a.Request.PreferredDate != "" {
		preferredDate = a.Request.PreferredDate
	}

	_, err := r.pool.Exec(ctx,
		`INSERT INTO appointment (id, status, created_at, name, phone, tasks, preferred_date,
		                          preferred_time, ideas, files, wish, consent, place, validation)
		 VALUES ($1,$2,$3,$4,$5,$6::jsonb,$7,$8,$9::jsonb,$10::jsonb,$11,$12,$13,$14::jsonb)
		 ON CONFLICT (id) DO NOTHING`,
		a.ID, a.Status, a.CreatedAt, a.Request.Name, a.Request.Phone, string(tasks), preferredDate,
		a.Request.PreferredTime, string(ideas), string(files), a.Request.Wish, a.Request.Consent,
		domain.PlaceAddress, string(validation))
	return err
}

// ByID возвращает запись по идентификатору.
func (r *Repository) ByID(ctx context.Context, id string) (domain.Appointment, error) {
	var (
		a             domain.Appointment
		tasksRaw      []byte
		ideasRaw      []byte
		filesRaw      []byte
		validationRaw []byte
		preferredDate *time.Time
	)
	err := r.pool.QueryRow(ctx,
		`SELECT id, status, created_at, name, phone, tasks, preferred_date, preferred_time,
		        ideas, files, wish, consent, validation
		   FROM appointment WHERE id = $1`, id).
		Scan(&a.ID, &a.Status, &a.CreatedAt, &a.Request.Name, &a.Request.Phone, &tasksRaw,
			&preferredDate, &a.Request.PreferredTime, &ideasRaw, &filesRaw, &a.Request.Wish,
			&a.Request.Consent, &validationRaw)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Appointment{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Appointment{}, err
	}
	if preferredDate != nil {
		a.Request.PreferredDate = preferredDate.Format("2006-01-02")
	}
	_ = json.Unmarshal(tasksRaw, &a.Request.Tasks)
	_ = json.Unmarshal(ideasRaw, &a.Request.Ideas)
	_ = json.Unmarshal(filesRaw, &a.Request.Files)
	_ = json.Unmarshal(validationRaw, &a.Validation)
	return a, nil
}

// Count возвращает число записей.
func (r *Repository) Count(ctx context.Context) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `SELECT count(*) FROM appointment`).Scan(&n)
	return n, err
}
