// Package domain контекста Identity: пользователи и роли.
package domain

import (
	"context"
	"errors"
)

// ErrInvalidCredentials — неверный email или пароль.
var ErrInvalidCredentials = errors.New("Неверный email или пароль")

// User — пользователь сервиса.
type User struct {
	ID           string
	Email        string
	Name         string
	PasswordHash string
	Role         string
}

// Repository — порт хранения пользователей.
type Repository interface {
	ByEmail(ctx context.Context, email string) (User, error)
	Save(ctx context.Context, users []User) error
}

// Service — операции контекста Identity.
type Service struct {
	repo Repository
}

// NewService собирает сервис.
func NewService(repo Repository) *Service { return &Service{repo: repo} }

// ByEmail возвращает пользователя по email.
func (s *Service) ByEmail(ctx context.Context, email string) (User, error) {
	return s.repo.ByEmail(ctx, email)
}
