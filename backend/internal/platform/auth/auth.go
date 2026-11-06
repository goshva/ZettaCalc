// Package auth — code-based проверки доступа: JWT и роли.
//
// ACL проверяется только кодом (AGENTS.md, принцип 3). LLM в этот путь не
// допускается: ни один метод пакета не принимает текстовое решение модели.
package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ErrUnauthorized — токен отсутствует, просрочен или подписан неверным ключом.
var ErrUnauthorized = errors.New("auth: требуется авторизация")

// ErrInvalidToken — токен передан, но не принят: mock API отвечает
// «Invalid token» вместо «Not authenticated».
var ErrInvalidToken = errors.New("auth: неверный токен")

// Claims — полезная нагрузка токена.
type Claims struct {
	UserID string `json:"sub"`
	Email  string `json:"email"`
	Role   string `json:"role,omitempty"`
}

type contextKey struct{}

// Service выпускает и проверяет токены.
type Service struct {
	secret []byte
	ttl    time.Duration
}

// New создаёт сервис токенов.
func New(secret string, ttl time.Duration) *Service {
	return &Service{secret: []byte(secret), ttl: ttl}
}

// Issue выпускает HS256-токен с теми же полями, что и mock API.
func (s *Service) Issue(userID, email, role string) (string, error) {
	claims := jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"exp":   time.Now().Add(s.ttl).Unix(),
	}
	if role != "" {
		claims["role"] = role
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
}

// Parse проверяет подпись и срок действия токена.
func (s *Service) Parse(token string) (Claims, error) {
	parsed, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrUnauthorized
		}
		return s.secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || !parsed.Valid {
		return Claims{}, ErrUnauthorized
	}
	mc, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return Claims{}, ErrUnauthorized
	}
	sub, _ := mc["sub"].(string)
	email, _ := mc["email"].(string)
	role, _ := mc["role"].(string)
	if sub == "" {
		return Claims{}, ErrUnauthorized
	}
	return Claims{UserID: sub, Email: email, Role: role}, nil
}

// Middleware проверяет Bearer-токен и кладёт Claims в контекст.
func (s *Service) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw := bearer(r)
		if raw == "" {
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), contextKey{}, Claims{})))
			return
		}
		claims, err := s.Parse(raw)
		if err != nil {
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), contextKey{}, Claims{})))
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), contextKey{}, claims)))
	})
}

// FromContext возвращает Claims, если токен был валиден.
func FromContext(ctx context.Context) (Claims, bool) {
	c, ok := ctx.Value(contextKey{}).(Claims)
	return c, ok && c.UserID != ""
}

// RequireToken — кодовая проверка наличия авторизации.
//
// Различает «токена нет» и «токен неверный»: mock API отвечает на это разными
// текстами, и фронтенды их показывают.
func RequireToken(r *http.Request) (Claims, error) {
	if c, ok := FromContext(r.Context()); ok {
		return c, nil
	}
	if bearer(r) != "" {
		return Claims{}, ErrInvalidToken
	}
	return Claims{}, ErrUnauthorized
}

func bearer(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if h == "" {
		return ""
	}
	const prefix = "Bearer "
	if len(h) <= len(prefix) || !strings.EqualFold(h[:len(prefix)], prefix) {
		return ""
	}
	return strings.TrimSpace(h[len(prefix):])
}
