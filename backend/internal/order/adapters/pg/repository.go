// Package pg — PostgreSQL-адаптер контекста Order.
package pg

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zetta/zettacalc/backend/internal/order/domain"
)

// Repository читает корзину.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository создаёт репозиторий корзины.
func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

// Cart возвращает позиции корзины.
func (r *Repository) Cart(ctx context.Context) (domain.Cart, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT sku, COALESCE(brand, ''), title, price, qty, COALESCE(image_key, '')
		   FROM cart_item ORDER BY sort_order`)
	if err != nil {
		return domain.Cart{}, err
	}
	defer rows.Close()

	cart := domain.Cart{Items: []domain.Item{}}
	for rows.Next() {
		var it domain.Item
		if err := rows.Scan(&it.SKU, &it.Brand, &it.Title, &it.Price, &it.Qty, &it.ImageKey); err != nil {
			return domain.Cart{}, err
		}
		cart.Items = append(cart.Items, it)
		cart.TotalRub += int64(it.Price * float64(it.Qty))
	}
	return cart, rows.Err()
}

// Save сохраняет корзину.
func (r *Repository) Save(ctx context.Context, cart domain.Cart) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `DELETE FROM cart_item`); err != nil {
		return err
	}
	for i, it := range cart.Items {
		var imageKey any
		if it.ImageKey != "" {
			imageKey = it.ImageKey
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO cart_item (id, sku, brand, title, price, qty, image_key, sort_order)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
			"cart_"+itoa(i), it.SKU, it.Brand, it.Title, it.Price, it.Qty, imageKey, i); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

var _ = errors.New
var _ = pgx.ErrNoRows
