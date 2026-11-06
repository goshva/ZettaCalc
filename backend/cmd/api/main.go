// Command api — точка входа Go-сервиса ZettaCalc.
//
// Здесь единственное место, где собираются конкретные реализации контекстов
// (composition root): домены и app-слои о PostgreSQL и HTTP не знают.
package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/zetta/zettacalc/backend/internal/httpapi"
	"github.com/zetta/zettacalc/backend/internal/platform/config"
	"github.com/zetta/zettacalc/backend/internal/platform/pgxdb"

	assetsdomain "github.com/zetta/zettacalc/backend/internal/assets/domain"
	bookingdomain "github.com/zetta/zettacalc/backend/internal/booking/domain"
	designdomain "github.com/zetta/zettacalc/backend/internal/design/domain"
	estimatedomain "github.com/zetta/zettacalc/backend/internal/estimate/domain"
	identitydomain "github.com/zetta/zettacalc/backend/internal/identity/domain"
	mediadomain "github.com/zetta/zettacalc/backend/internal/media/domain"
	orderdomain "github.com/zetta/zettacalc/backend/internal/order/domain"
	techcarddomain "github.com/zetta/zettacalc/backend/internal/techcard/domain"
	tenderdomain "github.com/zetta/zettacalc/backend/internal/tender/domain"

	designapp "github.com/zetta/zettacalc/backend/internal/design/app"

	assetspg "github.com/zetta/zettacalc/backend/internal/assets/adapters/pg"
	bookingpg "github.com/zetta/zettacalc/backend/internal/booking/adapters/pg"
	contentpg "github.com/zetta/zettacalc/backend/internal/content/adapters/pg"
	designpg "github.com/zetta/zettacalc/backend/internal/design/adapters/pg"
	estimatepg "github.com/zetta/zettacalc/backend/internal/estimate/adapters/pg"
	identitypg "github.com/zetta/zettacalc/backend/internal/identity/adapters/pg"
	mediapg "github.com/zetta/zettacalc/backend/internal/media/adapters/pg"
	orderpg "github.com/zetta/zettacalc/backend/internal/order/adapters/pg"
	techcardpg "github.com/zetta/zettacalc/backend/internal/techcard/adapters/pg"
	tenderpg "github.com/zetta/zettacalc/backend/internal/tender/adapters/pg"

	"github.com/zetta/zettacalc/backend/internal/platform/auth"
	"github.com/zetta/zettacalc/backend/internal/platform/idempotency"
)

func main() {
	migrationsDir := flag.String("migrations", "migrations", "каталог SQL-миграций")
	flag.Parse()

	if err := run(*migrationsDir); err != nil {
		slog.Error("запуск сервиса не удался", "err", err)
		os.Exit(1)
	}
}

func run(migrationsDir string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := pgxdb.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	applied, err := pgxdb.Migrate(ctx, pool, migrationsDir)
	if err != nil {
		return err
	}
	if len(applied) > 0 {
		slog.Info("применены миграции", "versions", applied)
	}

	// ── Репозитории (adapters) ────────────────────────────────────────────
	mediaRepo := mediapg.NewRepository(pool, cfg.StaticImagesDir)
	contentRepo := contentpg.NewRepository(pool)
	assetsRepo := assetspg.NewRepository(pool)
	designRepo := designpg.NewRepository(pool)
	estimateRepo := estimatepg.NewRepository(pool)
	techcardRepo := techcardpg.NewRepository(pool)
	tenderRepo := tenderpg.NewRepository(pool)
	bookingRepo := bookingpg.NewRepository(pool)
	identityRepo := identitypg.NewRepository(pool)
	orderRepo := orderpg.NewRepository(pool)

	// ── Сервисы (domain) ──────────────────────────────────────────────────
	tokens := auth.New(cfg.JWTSecret, cfg.TokenTTL)

	deps := httpapi.Deps{
		Pool:       pool,
		Media:      mediadomain.NewService(mediaRepo),
		Content:    contentRepo,
		Assets:     assetsdomain.NewService(assetsRepo),
		Design:     designdomain.NewService(designRepo),
		Derive:     designapp.NewDerive(designdomain.NewService(designRepo)),
		Estimate:   estimatedomain.NewService(estimateRepo),
		Techcards:  techcarddomain.NewService(techcardRepo),
		Tender:     tenderdomain.NewService(tenderRepo),
		Booking:    bookingdomain.NewService(bookingRepo),
		Identity:   identitydomain.NewService(identityRepo),
		Order:      orderdomain.NewService(orderRepo),
		Tokens:     tokens,
		Idempotent: idempotency.NewStore(pool, cfg.IdempotencyTTL),
		StartedAt:  time.Now(),
	}

	handler := httpapi.New(deps)
	server := &http.Server{
		Addr:              cfg.Addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       90 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("сервис запущен", "addr", cfg.Addr, "engine", "go")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		slog.Info("останавливаю сервис")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	return server.Shutdown(shutdownCtx)
}
