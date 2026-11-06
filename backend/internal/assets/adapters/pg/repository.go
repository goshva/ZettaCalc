// Package pg — PostgreSQL-адаптер контекста Assets.
package pg

import (
	"context"
	"errors"
	"path"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zetta/zettacalc/backend/internal/assets/domain"
)

// Repository читает текстовые ресурсы из PostgreSQL.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository создаёт репозиторий ресурсов.
func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

const selectFile = `SELECT path, kind, COALESCE(app_name, ''), COALESCE(proto, ''), name,
                           content_type, body, sort_order
                      FROM content_file`

func (r *Repository) scanFile(ctx context.Context, q string, args ...any) (domain.File, error) {
	var f domain.File
	err := r.pool.QueryRow(ctx, q, args...).Scan(&f.Path, &f.Kind, &f.AppName, &f.Proto,
		&f.Name, &f.ContentType, &f.Body, &f.Order)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.File{}, domain.ErrNotFound
	}
	return f, err
}

// ByPath возвращает ресурс по пути.
func (r *Repository) ByPath(ctx context.Context, path string) (domain.File, error) {
	return r.scanFile(ctx, selectFile+` WHERE path = $1`, path)
}

// AppFile возвращает файл приложения по имени (main.css, interiors.html).
func (r *Repository) AppFile(ctx context.Context, app, name string) (domain.File, error) {
	return r.scanFile(ctx, selectFile+` WHERE kind = 'app' AND app_name = $1 AND path LIKE $2`,
		app, "%/"+name)
}

// MainSection возвращает секцию главной страницы.
//
// mock API читает файл <name>.html напрямую из каталога main_page, поэтому
// поиск идёт по пути, а не по имени без расширения.
func (r *Repository) MainSection(ctx context.Context, name string) (domain.File, error) {
	return r.scanFile(ctx, selectFile+` WHERE kind = 'main_section' AND path = $1`,
		"main_section/"+name+".html")
}

// PageFragment возвращает фрагмент разметки прототипа.
//
// В таблице имя хранится без расширения, а маршрут принимает имя файла
// (00_shell.html), поэтому расширение отбрасывается при поиске.
func (r *Repository) PageFragment(ctx context.Context, proto, name string) (domain.File, error) {
	base := strings.TrimSuffix(name, path.Ext(name))
	return r.scanFile(ctx, selectFile+` WHERE kind = 'page' AND proto = $1 AND name = $2`, proto, base)
}

// ProtoCSS возвращает таблицу стилей прототипа по имени файла.
func (r *Repository) ProtoCSS(ctx context.Context, name string) (domain.File, error) {
	return r.scanFile(ctx, selectFile+` WHERE kind = 'css' AND path = $1`, "css/"+name)
}

// List возвращает список ресурсов заданного вида.
func (r *Repository) List(ctx context.Context, kind domain.Kind, proto string) ([]domain.File, error) {
	rows, err := r.pool.Query(ctx, selectFile+`
	                                 WHERE kind = $1 AND ($2 = '' OR proto = $2)
	                                 ORDER BY sort_order, name`, string(kind), proto)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.File, 0, 16)
	for rows.Next() {
		var f domain.File
		if err := rows.Scan(&f.Path, &f.Kind, &f.AppName, &f.Proto, &f.Name,
			&f.ContentType, &f.Body, &f.Order); err != nil {
			return nil, err
		}
		out = append(out, f)
	}
	return out, rows.Err()
}

// Save сохраняет ресурсы пачкой.
func (r *Repository) Save(ctx context.Context, files []domain.File) error {
	if len(files) == 0 {
		return nil
	}
	batch := &pgx.Batch{}
	for _, f := range files {
		batch.Queue(
			`INSERT INTO content_file (path, kind, app_name, proto, name, content_type, body,
			                           sha256, byte_size, sort_order)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
			 ON CONFLICT (path) DO UPDATE
			    SET kind = EXCLUDED.kind, app_name = EXCLUDED.app_name, proto = EXCLUDED.proto,
			        name = EXCLUDED.name, content_type = EXCLUDED.content_type,
			        body = EXCLUDED.body, sha256 = EXCLUDED.sha256,
			        byte_size = EXCLUDED.byte_size, sort_order = EXCLUDED.sort_order`,
			f.Path, string(f.Kind), nullStr(f.AppName), nullStr(f.Proto), f.Name,
			f.ContentType, f.Body, sha256Hex(f.Body), len(f.Body), f.Order)
	}
	results := r.pool.SendBatch(ctx, batch)
	defer func() { _ = results.Close() }()
	for range files {
		if _, err := results.Exec(); err != nil {
			return err
		}
	}
	return nil
}

// SaveMeta сохраняет метаданные развёртывания (каталоги фрагментов, счётчики).
func (r *Repository) SaveMeta(ctx context.Context, values map[string]string) error {
	if len(values) == 0 {
		return nil
	}
	batch := &pgx.Batch{}
	for key, value := range values {
		batch.Queue(
			`INSERT INTO seed_meta (key, value, updated_at) VALUES ($1, $2, now())
			 ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = now()`,
			key, value)
	}
	results := r.pool.SendBatch(ctx, batch)
	defer func() { _ = results.Close() }()
	for range values {
		if _, err := results.Exec(); err != nil {
			return err
		}
	}
	return nil
}

// Meta читает значение метаданных; отсутствие ключа — не ошибка.
func (r *Repository) Meta(ctx context.Context, key string) (string, error) {
	var value string
	err := r.pool.QueryRow(ctx, `SELECT value FROM seed_meta WHERE key = $1`, key).Scan(&value)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	return value, err
}
