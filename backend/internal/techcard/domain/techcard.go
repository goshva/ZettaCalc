// Package domain контекста Techcard: техкарты ремонтных работ.
//
// Техкарта — версионированная запись (code + version + owner). Правило проекта:
// обязательные операции не удаляются, у каждой карты есть версия и ответственный
// (проверяется scripts/check_techcards.ps1).
package domain

import (
	"context"
	"errors"
)

// ErrNotFound — техкарта отсутствует.
var ErrNotFound = errors.New("techcard: техкарта не найдена")

// Techcard — технологическая карта.
type Techcard struct {
	ID             string
	Code           string
	Title          string
	Surface        string
	Material       string
	Version        int
	MinAreaM2      *float64
	MaxUnevenness  *float64
	Operations     []string
	Consumption    map[string]float64
	Order          int
	Payload        []byte
}

// Repository — порт чтения техкарт.
type Repository interface {
	List(ctx context.Context, surface string) ([]Techcard, error)
	Save(ctx context.Context, cards []Techcard) error
}

// Service — операции контекста Techcard.
type Service struct {
	repo Repository
}

// NewService собирает сервис.
func NewService(repo Repository) *Service { return &Service{repo: repo} }

// List возвращает техкарты с фильтром по поверхности.
func (s *Service) List(ctx context.Context, surface string) ([]Techcard, error) {
	return s.repo.List(ctx, surface)
}
