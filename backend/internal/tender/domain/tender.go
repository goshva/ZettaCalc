// Package domain контекста Tender: аукцион подрядчиков.
//
// Инвариант из плана миграции: ставки сравнимы только в рамках одного ЦПР,
// поэтому у каждой ставки есть variant_id (NULL — статичная демонстрационная
// подборка /api/bids).
package domain

import (
	"context"
	"errors"
)

// ErrNotFound — ставка отсутствует.
var ErrNotFound = errors.New("tender: ставка не найдена")

// Bid — предложение подрядчика.
type Bid struct {
	ID        string
	Name      string
	Rating    *float64
	Price     *float64
	Days      *int
	Scope     string
	ImageKey  string
	VariantID string
	Order     int
	Payload   []byte
}

// Repository — порт чтения ставок.
type Repository interface {
	List(ctx context.Context, variantID string) ([]Bid, error)
	Save(ctx context.Context, bids []Bid) error
}

// Service — операции контекста Tender.
type Service struct {
	repo Repository
}

// NewService собирает сервис.
func NewService(repo Repository) *Service { return &Service{repo: repo} }

// Static возвращает демонстрационную подборку /api/bids.
func (s *Service) Static(ctx context.Context) ([]Bid, error) {
	return s.repo.List(ctx, "")
}
