// Package pg — PostgreSQL-адаптер контекста Identity.
package pg

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zetta/zettacalc/backend/internal/identity/domain"
)

// Repository читает пользователей.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository создаёт репозиторий пользователей.
func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

// ByEmail возвращает пользователя по email.
func (r *Repository) ByEmail(ctx context.Context, email string) (domain.User, error) {
	var u domain.User
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, name, password_hash, COALESCE(role, '') FROM app_user WHERE email = $1`,
		email).Scan(&u.ID, &u.Email, &u.Name, &u.PasswordHash, &u.Role)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, domain.ErrInvalidCredentials
	}
	return u, err
}

// Save сохраняет пользователей.
func (r *Repository) Save(ctx context.Context, users []domain.User) error {
	for _, u := range users {
		if _, err := r.pool.Exec(ctx,
			`INSERT INTO app_user (id, email, name, password_hash, role)
			 VALUES ($1,$2,$3,$4,$5)
			 ON CONFLICT (email) DO UPDATE
			    SET name = EXCLUDED.name,
			        password_hash = EXCLUDED.password_hash,
			        role = EXCLUDED.role`,
			u.ID, u.Email, u.Name, u.PasswordHash, u.Role); err != nil {
			return err
		}
	}
	return nil
}
