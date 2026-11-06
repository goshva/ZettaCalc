// Package pg — PostgreSQL-адаптер контекста Content.
package pg

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/zetta/zettacalc/backend/internal/content/domain"
)

// Repository читает контент из PostgreSQL.
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository создаёт репозиторий контента.
func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

// Interiors возвращает интерьеры с фильтрами mock API.
//
// Фильтр category повторяет поведение Python: сравнивается поле "category"
// в самой сущности, а его в dataset.json нет — значит ответ пустой список.
func (r *Repository) Interiors(ctx context.Context, f domain.InteriorFilter) ([]domain.Interior, error) {
	const q = `SELECT id, type, title, COALESCE(room, ''), COALESCE(style, ''), COALESCE(palette, ''),
	                  COALESCE(image_file, ''), COALESCE(source_url, ''), area_m2, COALESCE(parent_id, ''),
	                  sort_order, payload_text
	             FROM interior
	            WHERE ($1 = '' OR payload_text::jsonb ->> 'category' = $1)
	              AND ($2 = '' OR room = $2)
	              AND ($3 = '' OR style = $3)
	            ORDER BY sort_order`
	rows, err := r.pool.Query(ctx, q, f.Category, f.Room, f.Style)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.Interior, 0, 32)
	for rows.Next() {
		var it domain.Interior
		if err := rows.Scan(&it.ID, &it.Type, &it.Title, &it.Room, &it.Style, &it.Palette,
			&it.ImageFile, &it.SourceURL, &it.AreaM2, &it.ParentID, &it.Order, &it.Payload); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

// InteriorByID возвращает интерьер по идентификатору.
func (r *Repository) InteriorByID(ctx context.Context, id string) (domain.Interior, error) {
	const q = `SELECT id, type, title, COALESCE(room, ''), COALESCE(style, ''), COALESCE(palette, ''),
	                  COALESCE(image_file, ''), COALESCE(source_url, ''), area_m2, COALESCE(parent_id, ''),
	                  sort_order, payload_text
	             FROM interior WHERE id = $1`
	var it domain.Interior
	err := r.pool.QueryRow(ctx, q, id).Scan(&it.ID, &it.Type, &it.Title, &it.Room, &it.Style,
		&it.Palette, &it.ImageFile, &it.SourceURL, &it.AreaM2, &it.ParentID, &it.Order, &it.Payload)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Interior{}, domain.ErrNotFound
	}
	return it, err
}

// InteriorCategories возвращает категории интерьеров.
func (r *Repository) InteriorCategories(ctx context.Context) ([]domain.Named, error) {
	return r.named(ctx, `SELECT name, sort_order FROM interior_category ORDER BY sort_order`)
}

// InteriorDescriptions возвращает описания категорий.
func (r *Repository) InteriorDescriptions(ctx context.Context) ([]domain.Named, error) {
	return r.named(ctx, `SELECT name, sort_order FROM interior_description ORDER BY sort_order`)
}

func (r *Repository) named(ctx context.Context, q string) ([]domain.Named, error) {
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.Named, 0, 8)
	for rows.Next() {
		var n domain.Named
		if err := rows.Scan(&n.Name, &n.Order); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

// Products возвращает товары каталога с фильтром по group/shape.
func (r *Repository) Products(ctx context.Context, group string) ([]domain.Product, error) {
	const q = `SELECT id, name, COALESCE(url, ''), price, COALESCE(brand, ''), COALESCE(shape, ''),
	                  COALESCE(image_file, ''), sort_order, payload_text
	             FROM catalog_product
	            WHERE ($1 = '' OR shape = $1 OR payload_text::jsonb ->> 'group' = $1)
	            ORDER BY sort_order`
	rows, err := r.pool.Query(ctx, q, group)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.Product, 0, 16)
	for rows.Next() {
		var p domain.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.URL, &p.Price, &p.Brand, &p.Shape,
			&p.ImageFile, &p.Order, &p.Payload); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// CatalogRooms возвращает комнаты комплектации.
func (r *Repository) CatalogRooms(ctx context.Context) ([]domain.Room, error) {
	const q = `SELECT id, title, COALESCE(room, ''), COALESCE(palette, ''), area_m2,
	                  COALESCE(parent_id, ''), COALESCE(image_file, ''), sort_order, payload_text
	             FROM catalog_room ORDER BY sort_order`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.Room, 0, 16)
	for rows.Next() {
		var rm domain.Room
		if err := rows.Scan(&rm.ID, &rm.Title, &rm.Room, &rm.Palette, &rm.AreaM2,
			&rm.ParentID, &rm.ImageFile, &rm.Order, &rm.Payload); err != nil {
			return nil, err
		}
		out = append(out, rm)
	}
	return out, rows.Err()
}

// CatalogRoomByID возвращает комнату по идентификатору.
func (r *Repository) CatalogRoomByID(ctx context.Context, id string) (domain.Room, error) {
	const q = `SELECT id, title, COALESCE(room, ''), COALESCE(palette, ''), area_m2,
	                  COALESCE(parent_id, ''), COALESCE(image_file, ''), sort_order, payload_text
	             FROM catalog_room WHERE id = $1`
	var rm domain.Room
	err := r.pool.QueryRow(ctx, q, id).Scan(&rm.ID, &rm.Title, &rm.Room, &rm.Palette, &rm.AreaM2,
		&rm.ParentID, &rm.ImageFile, &rm.Order, &rm.Payload)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Room{}, domain.ErrNotFound
	}
	return rm, err
}

// CenterInteriors возвращает интерьеры центра.
func (r *Repository) CenterInteriors(ctx context.Context) ([]domain.CenterInterior, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, title, COALESCE(image_file, ''), sort_order, payload_text
		   FROM center_interior ORDER BY sort_order`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.CenterInterior, 0, 32)
	for rows.Next() {
		var c domain.CenterInterior
		if err := rows.Scan(&c.ID, &c.Title, &c.Image, &c.Order, &c.Payload); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// CenterPhotos возвращает фото центра.
func (r *Repository) CenterPhotos(ctx context.Context) ([]domain.CenterPhoto, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, COALESCE(image_file, ''), sort_order, payload_text
		   FROM center_photo ORDER BY sort_order`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.CenterPhoto, 0, 8)
	for rows.Next() {
		var c domain.CenterPhoto
		if err := rows.Scan(&c.ID, &c.Name, &c.Image, &c.Order, &c.Payload); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// HotspotItems возвращает товары hotspot-подборки.
func (r *Repository) HotspotItems(ctx context.Context) ([]domain.HotspotItem, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, COALESCE(category, ''), price, COALESCE(unit, ''), COALESCE(image_file, ''),
		        point_x, point_y, sort_order, payload_text
		   FROM hotspot_item ORDER BY sort_order`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.HotspotItem, 0, 16)
	for rows.Next() {
		var h domain.HotspotItem
		if err := rows.Scan(&h.ID, &h.Name, &h.Cat, &h.Price, &h.Unit, &h.Image,
			&h.PointX, &h.PointY, &h.Order, &h.Payload); err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

// AnalogueItems возвращает аналоги товаров.
func (r *Repository) AnalogueItems(ctx context.Context) ([]domain.Analogue, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, COALESCE(short_name, ''), price, COALESCE(group_name, ''), COALESCE(unit, ''),
		        COALESCE(url, ''), is_external, COALESCE(image_file, ''), sort_order, payload_text
		   FROM hotspot_analogue ORDER BY sort_order`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.Analogue, 0, 4)
	for rows.Next() {
		var a domain.Analogue
		if err := rows.Scan(&a.ID, &a.Name, &a.Short, &a.Price, &a.Group, &a.Unit, &a.URL,
			&a.External, &a.Image, &a.Order, &a.Payload); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// References возвращает референсы.
func (r *Repository) References(ctx context.Context, category string) ([]domain.Reference, error) {
	const q = `SELECT id, title, COALESCE(category, ''), COALESCE(image_key, ''), price_from, price_to,
	                  sort_order, payload_text
	             FROM reference_item
	            WHERE ($1 = '' OR category = $1)
	            ORDER BY sort_order`
	rows, err := r.pool.Query(ctx, q, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.Reference, 0, 8)
	for rows.Next() {
		var ref domain.Reference
		if err := rows.Scan(&ref.ID, &ref.Title, &ref.Category, &ref.ImageKey,
			&ref.PriceFrom, &ref.PriceTo, &ref.Order, &ref.Payload); err != nil {
			return nil, err
		}
		out = append(out, ref)
	}
	return out, rows.Err()
}

// ReferenceByID возвращает референс по идентификатору.
func (r *Repository) ReferenceByID(ctx context.Context, id string) (domain.Reference, error) {
	const q = `SELECT id, title, COALESCE(category, ''), COALESCE(image_key, ''), price_from, price_to,
	                  sort_order, payload_text
	             FROM reference_item WHERE id = $1`
	var ref domain.Reference
	err := r.pool.QueryRow(ctx, q, id).Scan(&ref.ID, &ref.Title, &ref.Category, &ref.ImageKey,
		&ref.PriceFrom, &ref.PriceTo, &ref.Order, &ref.Payload)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Reference{}, domain.ErrNotFound
	}
	return ref, err
}

// Connected возвращает подборку и её товары.
func (r *Repository) Connected(ctx context.Context) (domain.ConnectedPick, []domain.ConnectedItem, error) {
	var pick domain.ConnectedPick
	err := r.pool.QueryRow(ctx,
		`SELECT id, COALESCE(image_key, ''), title, sort_order, payload_text
		   FROM connected_pick ORDER BY sort_order LIMIT 1`).
		Scan(&pick.ID, &pick.ImageKey, &pick.Title, &pick.Order, &pick.Payload)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ConnectedPick{}, nil, domain.ErrNotFound
	}
	if err != nil {
		return domain.ConnectedPick{}, nil, err
	}

	rows, err := r.pool.Query(ctx,
		`SELECT COALESCE(sku, ''), COALESCE(brand, ''), title, price, COALESCE(image_key, ''), position
		   FROM connected_item WHERE id = $1 ORDER BY position`, pick.ID)
	if err != nil {
		return pick, nil, err
	}
	defer rows.Close()
	items := make([]domain.ConnectedItem, 0, 4)
	for rows.Next() {
		var it domain.ConnectedItem
		if err := rows.Scan(&it.SKU, &it.Brand, &it.Title, &it.Price, &it.ImageKey, &it.Order); err != nil {
			return pick, nil, err
		}
		items = append(items, it)
	}
	return pick, items, rows.Err()
}

// Planners возвращает планировочные решения.
func (r *Repository) Planners(ctx context.Context) ([]domain.Planner, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, area_m2, sort_order, payload_text FROM planner ORDER BY sort_order`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]domain.Planner, 0, 4)
	for rows.Next() {
		var p domain.Planner
		if err := rows.Scan(&p.ID, &p.Name, &p.AreaM2, &p.Order, &p.Payload); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
