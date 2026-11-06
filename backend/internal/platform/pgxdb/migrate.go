package pgxdb

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Migrate применяет неприменённые SQL-файлы из каталога dir по возрастанию имени.
//
// Каждый файл выполняется в отдельной транзакции и сам фиксирует свой номер в
// schema_migration, поэтому повторный запуск безопасен.
func Migrate(ctx context.Context, pool *pgxpool.Pool, dir string) (applied []string, err error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("pgxdb: чтение каталога миграций %s: %w", dir, err)
	}

	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)

	for _, name := range names {
		version := strings.TrimSuffix(name, filepath.Ext(name))

		var exists bool
		err := pool.QueryRow(ctx,
			`SELECT EXISTS (SELECT 1 FROM information_schema.tables
			                 WHERE table_schema = 'public' AND table_name = 'schema_migration')`).Scan(&exists)
		if err != nil {
			return applied, fmt.Errorf("pgxdb: проверка schema_migration: %w", err)
		}
		if exists {
			var done bool
			err = pool.QueryRow(ctx,
				`SELECT EXISTS (SELECT 1 FROM schema_migration WHERE version = $1)`, version).Scan(&done)
			if err != nil {
				return applied, fmt.Errorf("pgxdb: проверка версии %s: %w", version, err)
			}
			if done {
				continue
			}
		}

		body, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return applied, fmt.Errorf("pgxdb: чтение %s: %w", name, err)
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			return applied, fmt.Errorf("pgxdb: начало транзакции %s: %w", name, err)
		}
		if _, err := tx.Exec(ctx, string(body)); err != nil {
			_ = tx.Rollback(ctx)
			return applied, fmt.Errorf("pgxdb: миграция %s: %w", name, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return applied, fmt.Errorf("pgxdb: коммит %s: %w", name, err)
		}
		applied = append(applied, version)
	}
	return applied, nil
}
