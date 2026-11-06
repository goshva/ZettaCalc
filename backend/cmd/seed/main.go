// Command seed переносит данные mock API в PostgreSQL.
//
// Запуск (из каталога backend):
//
//	go run ./cmd/seed -data ../vue-prototypes/api/data -images ../vue-prototypes/api/static/images
package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"time"

	"github.com/zetta/zettacalc/backend/internal/platform/config"
	"github.com/zetta/zettacalc/backend/internal/platform/pgxdb"
	"github.com/zetta/zettacalc/backend/internal/seed"

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
)

func main() {
	dataDir := flag.String("data", "", "каталог данных mock API (vue-prototypes/api/data)")
	imagesDir := flag.String("images", "", "каталог картинок (vue-prototypes/api/static/images)")
	migrationsDir := flag.String("migrations", "migrations", "каталог SQL-миграций")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	if err := run(*dataDir, *imagesDir, *migrationsDir); err != nil {
		slog.Error("загрузка данных не удалась", "err", err)
		os.Exit(1)
	}
}

func run(dataDir, imagesDir, migrationsDir string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if dataDir == "" {
		dataDir = cfg.SeedDataDir
	}
	if imagesDir == "" {
		imagesDir = cfg.SeedImagesDir
	}
	if dataDir == "" || imagesDir == "" {
		slog.Error("не заданы каталоги данных", "data", dataDir, "images", imagesDir)
		os.Exit(2)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	pool, err := pgxdb.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	if applied, err := pgxdb.Migrate(ctx, pool, migrationsDir); err != nil {
		return err
	} else if len(applied) > 0 {
		slog.Info("применены миграции", "versions", applied)
	}

	loader := seed.New(
		seed.Options{DataDir: dataDir, ImagesDir: imagesDir, Pool: pool},
		mediapg.NewRepository(pool, imagesDir),
		contentpg.NewRepository(pool),
		assetspg.NewRepository(pool),
		designpg.NewRepository(pool),
		estimatepg.NewRepository(pool),
		techcardpg.NewRepository(pool),
		tenderpg.NewRepository(pool),
		bookingpg.NewRepository(pool),
		identitypg.NewRepository(pool),
		orderpg.NewRepository(pool),
	)

	started := time.Now()
	report, err := loader.Run(ctx)
	if err != nil {
		return err
	}
	slog.Info("данные перенесены",
		"images", report.Images,
		"image_bytes", report.ImageBytes,
		"image_keys", report.ImageKeys,
		"entity_refs", report.EntityRefs,
		"interiors", report.Interiors,
		"products", report.Products,
		"rooms", report.Rooms,
		"extras", report.Extras,
		"center_interiors", report.CenterInteriors,
		"center_photos", report.CenterPhotos,
		"hotspot_items", report.HotspotItems,
		"analogues", report.Analogues,
		"references", report.References,
		"planners", report.Planners,
		"techcards", report.Techcards,
		"design_variants", report.DesignVariants,
		"content_files", report.ContentFiles,
		"elapsed", time.Since(started).String(),
	)
	return nil
}
