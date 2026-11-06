// Package domain контекста Design: четыре варианта дизайна и производные сметы.
//
// В прототипе это заглушка вместо загрузки фото: клиент выбирает вариант,
// для варианта детерминированно считаются смета, ЦПР и ставки подрядчиков.
// Расчёт только детерминированный, без LLM (AGENTS.md, принцип 1).
package domain

import (
	"context"
	"errors"
)

// ErrNotFound — вариант не найден.
var ErrNotFound = errors.New("design: вариант не найден")

// Variant — вариант дизайна.
type Variant struct {
	ID            string
	Name          string
	Style         string
	Palette       string
	ImagePath     string
	Description   string
	Rooms         []string
	BudgetMin     int64
	BudgetMax     int64
	EstimatedDays int
	Order         int
}

// Repository — порт чтения вариантов.
type Repository interface {
	Variants(ctx context.Context) ([]Variant, error)
	VariantsPayload(ctx context.Context) ([]byte, error)
	VariantByID(ctx context.Context, id string) (Variant, error)
	Save(ctx context.Context, variants []Variant, payloads map[string][]byte) error
}

// Service — операции контекста Design.
type Service struct {
	repo Repository
}

// NewService собирает сервис.
func NewService(repo Repository) *Service { return &Service{repo: repo} }

// Variants возвращает все варианты.
func (s *Service) Variants(ctx context.Context) ([]Variant, error) { return s.repo.Variants(ctx) }

// VariantsJSON отдаёт массив вариантов как в mock API.
func (s *Service) VariantsJSON(ctx context.Context) ([]byte, error) {
	return s.repo.VariantsPayload(ctx)
}

// ByID возвращает вариант по идентификатору.
func (s *Service) ByID(ctx context.Context, id string) (Variant, error) {
	return s.repo.VariantByID(ctx, id)
}
