// Package domain контекста Estimate: смета варианта дизайна.
//
// Смета варианта считается детерминированно из бюджета варианта — ровно так,
// как это делала заглушка mock API. Никаких LLM и «придуманных» цен.
package domain

import (
	"context"
	"errors"
)

// ErrNotFound — смета отсутствует.
var ErrNotFound = errors.New("estimate: смета не найдена")

// Line — строка сметы.
type Line struct {
	ID    string
	Title string
	Qty   float64
	Unit  string
	Price float64
	SKU   string
	Brand string
}

// Estimate — смета.
type Estimate struct {
	ID          string
	Status      string
	TotalRub    int64
	RangeMin    int64
	RangeMax    int64
	Days        int
	VariantID   string
	VariantName string
	ProjectID   string
	Lines       []Line
	Assumptions []string
}

// Repository — порт чтения статических смет.
type Repository interface {
	ByID(ctx context.Context, id string) (Estimate, error)
	PayloadByID(ctx context.Context, id string) ([]byte, error)
	Save(ctx context.Context, estimates []Estimate, payloads map[string][]byte) error
}

// Service — операции контекста Estimate.
type Service struct {
	repo Repository
}

// NewService собирает сервис.
func NewService(repo Repository) *Service { return &Service{repo: repo} }

// ByID возвращает статическую смету.
func (s *Service) ByID(ctx context.Context, id string) (Estimate, error) {
	return s.repo.ByID(ctx, id)
}

// JSONByID отдаёт смету как в mock API (те же байты).
func (s *Service) JSONByID(ctx context.Context, id string) ([]byte, error) {
	return s.repo.PayloadByID(ctx, id)
}
