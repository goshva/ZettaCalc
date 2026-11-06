// Package domain контекста Order: корзина и заказ.
//
// В mock API корзина — заглушка на один товар, заказ не создаётся. Контекст
// заведён, чтобы write-правила (Idempotency-Key, approval-gate) жили в коде,
// а не в комментариях: полноценный заказ появится здесь же.
package domain

import (
	"context"
)

// Item — позиция корзины.
type Item struct {
	SKU      string
	Brand    string
	Title    string
	Price    float64
	Qty      int
	ImageKey string
}

// Cart — корзина пользователя.
type Cart struct {
	Items    []Item
	TotalRub int64
}

// Repository — порт чтения корзины.
type Repository interface {
	Cart(ctx context.Context) (Cart, error)
	Save(ctx context.Context, cart Cart) error
}

// Service — операции контекста Order.
type Service struct {
	repo Repository
}

// NewService собирает сервис.
func NewService(repo Repository) *Service { return &Service{repo: repo} }

// Cart возвращает текущую корзину.
func (s *Service) Cart(ctx context.Context) (Cart, error) { return s.repo.Cart(ctx) }
