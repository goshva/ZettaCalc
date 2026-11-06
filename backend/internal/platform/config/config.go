// Package config собирает конфигурацию сервиса из переменных окружения.
//
// Секреты не логируются: значения читаются один раз и наружу не печатаются.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config — настройки HTTP-сервиса и подключения к PostgreSQL.
type Config struct {
	Addr            string
	DatabaseURL     string
	JWTSecret       string
	TokenTTL        time.Duration
	IdempotencyTTL  time.Duration
	MaxUploadBytes  int64
	SeedDataDir     string
	SeedImagesDir   string
	StaticImagesDir string
	ShutdownTimeout time.Duration
}

// Load читает окружение и валидирует обязательные значения.
func Load() (Config, error) {
	cfg := Config{
		Addr:            env("ZETTA_ADDR", "127.0.0.1:8080"),
		DatabaseURL:     env("DATABASE_URL", defaultDatabaseURL()),
		JWTSecret:       env("ZETTA_JWT_SECRET", "fresh-house-dev-key-2026-not-for-prod"),
		TokenTTL:        envDuration("ZETTA_TOKEN_TTL", 7*24*time.Hour),
		IdempotencyTTL:  envDuration("ZETTA_IDEMPOTENCY_TTL", 7*24*time.Hour),
		MaxUploadBytes:  envInt64("ZETTA_MAX_UPLOAD_BYTES", 32<<20),
		SeedDataDir:     env("ZETTA_SEED_DATA_DIR", defaultSeedDataDir()),
		SeedImagesDir:   env("ZETTA_SEED_IMAGES_DIR", defaultSeedImagesDir()),
		StaticImagesDir: env("ZETTA_STATIC_IMAGES_DIR", defaultSeedImagesDir()),
		ShutdownTimeout: envDuration("ZETTA_SHUTDOWN_TIMEOUT", 10*time.Second),
	}
	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("config: не задан DATABASE_URL")
	}
	return cfg, nil
}

func defaultDatabaseURL() string {
	host := env("PGHOST", "127.0.0.1")
	port := env("PGPORT", "5432")
	user := env("PGUSER", "postgres")
	pass := env("PGPASSWORD", "postgres")
	db := env("PGDATABASE", "zettacalc")
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, db)
}

func defaultSeedDataDir() string {
	if v := env("ZETTA_REPO_ROOT", ""); v != "" {
		return strings.TrimRight(v, `\/`) + string(os.PathSeparator) + "vue-prototypes" +
			string(os.PathSeparator) + "api" + string(os.PathSeparator) + "data"
	}
	return ""
}

func defaultSeedImagesDir() string {
	if v := env("ZETTA_REPO_ROOT", ""); v != "" {
		return strings.TrimRight(v, `\/`) + string(os.PathSeparator) + "vue-prototypes" +
			string(os.PathSeparator) + "api" + string(os.PathSeparator) + "static" +
			string(os.PathSeparator) + "images"
	}
	return ""
}

func env(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}

func envInt64(key string, fallback int64) int64 {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return fallback
	}
	return n
}
