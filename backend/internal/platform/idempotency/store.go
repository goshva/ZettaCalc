// Package idempotency — повтор write-запросов с тем же Idempotency-Key.
//
// Правило проекта (AGENTS.md, принцип 4): write-операции выполняются только с
// ключом идемпотентности. Повтор с тем же телом возвращает сохранённый ответ,
// повтор с другим телом — 409.
package idempotency

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrConflict — ключ уже использован с другим телом запроса.
var ErrConflict = errors.New("idempotency: ключ использован с другим телом запроса")

// Record — сохранённый ответ на write-запрос.
type Record struct {
	Key         string
	StatusCode  int
	Response    json.RawMessage
	RequestHash string
}

// Store хранит ответы идемпотентности в PostgreSQL.
type Store struct {
	pool *pgxpool.Pool
	ttl  time.Duration
}

// NewStore создаёт хранилище ключей идемпотентности.
func NewStore(pool *pgxpool.Pool, ttl time.Duration) *Store {
	return &Store{pool: pool, ttl: ttl}
}

// HashBody считает стабильный хэш тела запроса.
func HashBody(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

// LookupByKey возвращает сохранённый ответ по ключу без сверки тела запроса.
//
// Так работает mock API: повтор с известным ключом отдаёт прежний ответ, даже
// если тело отличается. Сверка тела нужна отдельно — см. Lookup.
func (s *Store) LookupByKey(ctx context.Context, scope, key string) (*Record, error) {
	var (
		storedHash string
		status     int
		response   []byte
	)
	err := s.pool.QueryRow(ctx,
		`SELECT request_hash, status_code, response
		   FROM idempotency_key
		  WHERE key = $1 AND scope = $2 AND expires_at > now()`,
		key, scope).Scan(&storedHash, &status, &response)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &Record{Key: key, StatusCode: status, Response: response, RequestHash: storedHash}, nil
}

// Lookup возвращает сохранённый ответ, если ключ уже применялся.
//
// Если ключ использован с другим телом — ErrConflict.
func (s *Store) Lookup(ctx context.Context, scope, key, bodyHash string) (*Record, error) {
	record, err := s.LookupByKey(ctx, scope, key)
	if err != nil || record == nil {
		return record, err
	}
	if record.RequestHash != bodyHash {
		return nil, ErrConflict
	}
	return record, nil
}

// Save сохраняет ответ для ключа.
func (s *Store) Save(ctx context.Context, scope, key, bodyHash string, status int, response []byte) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO idempotency_key (key, scope, request_hash, status_code, response, expires_at)
		 VALUES ($1, $2, $3, $4, $5, now() + $6::interval)
		 ON CONFLICT (key) DO NOTHING`,
		key, scope, bodyHash, status, response, s.ttl.String())
	return err
}

// Purge удаляет просроченные ключи.
func (s *Store) Purge(ctx context.Context) (int64, error) {
	tag, err := s.pool.Exec(ctx, `DELETE FROM idempotency_key WHERE expires_at <= now()`)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
