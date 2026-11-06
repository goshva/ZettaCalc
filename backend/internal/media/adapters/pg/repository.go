// Package pg — PostgreSQL-адаптер контекста Media.
package pg

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zetta/zettacalc/backend/internal/media/domain"
)

// Repository — PostgreSQL-реализация порта domain.Repository.
type Repository struct {
	pool *pgxpool.Pool
	// staticDir — каталог общей статики сайта. Такие файлы не дублируются в БД
	// (storage = client), но API обязан их отдавать, поэтому читаем с диска.
	staticDir string
}

// NewRepository создаёт репозиторий изображений.
func NewRepository(pool *pgxpool.Pool, staticDir string) *Repository {
	return &Repository{pool: pool, staticDir: staticDir}
}

// AssetByFilename читает файл по имени.
func (r *Repository) AssetByFilename(ctx context.Context, filename string) (domain.Asset, error) {
	const q = `SELECT filename, content_type, byte_size, sha256, storage, group_prefix, bytes
	             FROM media_asset WHERE filename = $1`
	return r.scanAsset(ctx, q, filename)
}

// AssetByKey читает файл по ключу PIM.
func (r *Repository) AssetByKey(ctx context.Context, key string) (domain.Asset, error) {
	const q = `SELECT a.filename, a.content_type, a.byte_size, a.sha256, a.storage, a.group_prefix, a.bytes
	             FROM image_key k JOIN media_asset a ON a.filename = k.filename
	            WHERE k.key = $1`
	return r.scanAsset(ctx, q, key)
}

func (r *Repository) scanAsset(ctx context.Context, q string, arg any) (domain.Asset, error) {
	var a domain.Asset
	err := r.pool.QueryRow(ctx, q, arg).Scan(
		&a.Filename, &a.ContentType, &a.ByteSize, &a.SHA256, &a.Storage, &a.GroupPrefix, &a.Bytes)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Asset{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Asset{}, err
	}
	// Общая статика сайта в БД не дублируется: файл читается с диска.
	if len(a.Bytes) == 0 && r.staticDir != "" {
		body, readErr := os.ReadFile(filepath.Join(r.staticDir, a.Filename))
		if readErr != nil {
			return domain.Asset{}, domain.ErrNotFound
		}
		a.Bytes = body
	}
	return a, nil
}

// SaveAssets сохраняет файлы пачкой (идемпотентно по имени файла).
func (r *Repository) SaveAssets(ctx context.Context, assets []domain.Asset) error {
	if len(assets) == 0 {
		return nil
	}
	batch := &pgx.Batch{}
	for _, a := range assets {
		batch.Queue(
			`INSERT INTO media_asset (filename, content_type, byte_size, sha256, storage, group_prefix, bytes)
			 VALUES ($1, $2, $3, $4, $5, $6, $7)
			 ON CONFLICT (filename) DO UPDATE
			    SET content_type = EXCLUDED.content_type,
			        byte_size    = EXCLUDED.byte_size,
			        sha256       = EXCLUDED.sha256,
			        storage      = EXCLUDED.storage,
			        group_prefix = EXCLUDED.group_prefix,
			        bytes        = COALESCE(EXCLUDED.bytes, media_asset.bytes)`,
			a.Filename, a.ContentType, a.ByteSize, a.SHA256, a.Storage, a.GroupPrefix, a.Bytes)
	}
	results := r.pool.SendBatch(ctx, batch)
	defer func() { _ = results.Close() }()
	for range assets {
		if _, err := results.Exec(); err != nil {
			return err
		}
	}
	return nil
}

// SaveKeys сохраняет ключи PIM.
func (r *Repository) SaveKeys(ctx context.Context, keys []domain.Key) error {
	if len(keys) == 0 {
		return nil
	}
	batch := &pgx.Batch{}
	for _, k := range keys {
		batch.Queue(
			`INSERT INTO image_key (key, filename, source, storage, sort_order)
			 VALUES ($1, $2, $3, $4, $5)
			 ON CONFLICT (key) DO UPDATE
			    SET filename   = EXCLUDED.filename,
			        source     = EXCLUDED.source,
			        storage    = EXCLUDED.storage,
			        sort_order = EXCLUDED.sort_order`,
			k.Key, nullable(k.Filename), k.Source, k.Storage, k.Order)
	}
	results := r.pool.SendBatch(ctx, batch)
	defer func() { _ = results.Close() }()
	for range keys {
		if _, err := results.Exec(); err != nil {
			return err
		}
	}
	return nil
}

// SaveRefs сохраняет привязки картинок к сущностям.
func (r *Repository) SaveRefs(ctx context.Context, refs []domain.EntityRef) error {
	if len(refs) == 0 {
		return nil
	}
	batch := &pgx.Batch{}
	for _, ref := range refs {
		batch.Queue(
			`INSERT INTO media_entity_ref (filename, entity_type, entity_id, role, position)
			 VALUES ($1, $2, $3, $4, $5)
			 ON CONFLICT DO NOTHING`,
			ref.Filename, ref.EntityType, ref.EntityID, ref.Role, ref.Position)
	}
	results := r.pool.SendBatch(ctx, batch)
	defer func() { _ = results.Close() }()
	for range refs {
		if _, err := results.Exec(); err != nil {
			return err
		}
	}
	return nil
}

// Keys возвращает все ключи PIM по порядку.
func (r *Repository) Keys(ctx context.Context) ([]domain.Key, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT key, COALESCE(filename, ''), source, storage, sort_order
		   FROM image_key ORDER BY sort_order`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.Key, 0, 64)
	for rows.Next() {
		var k domain.Key
		if err := rows.Scan(&k.Key, &k.Filename, &k.Source, &k.Storage, &k.Order); err != nil {
			return nil, err
		}
		out = append(out, k)
	}
	return out, rows.Err()
}

// Stats возвращает агрегаты хранилища.
func (r *Repository) Stats(ctx context.Context) (domain.Stats, error) {
	var s domain.Stats
	err := r.pool.QueryRow(ctx,
		`SELECT (SELECT count(*) FROM media_asset WHERE storage = 'entity'),
		        (SELECT count(*) FROM image_key),
		        (SELECT count(DISTINCT (entity_type, entity_id)) FROM media_entity_ref),
		        (SELECT COALESCE(sum(byte_size), 0) FROM media_asset WHERE storage = 'entity')`).
		Scan(&s.Assets, &s.Keys, &s.Entities, &s.Bytes)
	return s, err
}

func nullable(s string) any {
	if s == "" {
		return nil
	}
	return s
}
