package pg

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"

	"github.com/zetta/zettacalc/backend/internal/content/domain"
)

// SaveInteriors сохраняет интерьеры, категории и описания.
func (r *Repository) SaveInteriors(ctx context.Context, items []domain.Interior, categories, descriptions []domain.Named) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `DELETE FROM interior`); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM interior_category`); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM interior_description`); err != nil {
		return err
	}

	for _, it := range items {
		if _, err := tx.Exec(ctx,
			`INSERT INTO interior (id, type, title, room, style, palette, image_file, source_url,
			                       area_m2, parent_id, sort_order, payload, payload_text)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12::jsonb,$13)`,
			it.ID, it.Type, it.Title, nullStr(it.Room), nullStr(it.Style), nullStr(it.Palette),
			nullStr(it.ImageFile), nullStr(it.SourceURL), it.AreaM2, nullStr(it.ParentID),
			it.Order, string(it.Payload), string(it.Payload)); err != nil {
			return err
		}
		if err := saveInteriorDetails(ctx, tx, it); err != nil {
			return err
		}
	}

	for _, c := range categories {
		if _, err := tx.Exec(ctx,
			`INSERT INTO interior_category (name, sort_order) VALUES ($1,$2)
			 ON CONFLICT (name) DO UPDATE SET sort_order = EXCLUDED.sort_order`, c.Name, c.Order); err != nil {
			return err
		}
	}
	for _, d := range descriptions {
		if _, err := tx.Exec(ctx,
			`INSERT INTO interior_description (name, sort_order) VALUES ($1,$2)
			 ON CONFLICT (name) DO UPDATE SET sort_order = EXCLUDED.sort_order`, d.Name, d.Order); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func saveInteriorDetails(ctx context.Context, tx pgx.Tx, it domain.Interior) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(it.Payload, &raw); err != nil {
		return err
	}

	if colors, ok := raw["colors"]; ok {
		var list []string
		if err := json.Unmarshal(colors, &list); err == nil {
			for i, c := range list {
				if _, err := tx.Exec(ctx,
					`INSERT INTO interior_color (interior_id, position, color) VALUES ($1,$2,$3)
					 ON CONFLICT DO NOTHING`, it.ID, i, c); err != nil {
					return err
				}
			}
		}
	}
	if variants, ok := raw["variants"]; ok {
		var list []string
		if err := json.Unmarshal(variants, &list); err == nil {
			for i, v := range list {
				if _, err := tx.Exec(ctx,
					`INSERT INTO interior_variant (interior_id, position, name) VALUES ($1,$2,$3)
					 ON CONFLICT DO NOTHING`, it.ID, i, v); err != nil {
					return err
				}
			}
		}
	}
	if costs, ok := raw["costs"]; ok {
		var rows [][]float64
		if err := json.Unmarshal(costs, &rows); err == nil {
			for ri, row := range rows {
				for ci, amount := range row {
					if _, err := tx.Exec(ctx,
						`INSERT INTO interior_cost (interior_id, row_no, column_no, amount)
						 VALUES ($1,$2,$3,$4) ON CONFLICT DO NOTHING`, it.ID, ri, ci, amount); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// SaveCatalog сохраняет товары, комнаты, доппозиции и слоты каталога.
func (r *Repository) SaveCatalog(ctx context.Context, products []domain.Product, rooms []domain.Room,
	extras []domain.Extra, slots []domain.SlotImage) error {

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	for _, table := range []string{"catalog_product", "catalog_room", "catalog_extra", "catalog_slot_image"} {
		if _, err := tx.Exec(ctx, `DELETE FROM `+table); err != nil {
			return err
		}
	}

	for _, p := range products {
		if _, err := tx.Exec(ctx,
			`INSERT INTO catalog_product (id, name, url, price, brand, shape, image_file, sort_order, payload, payload_text)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9::jsonb,$10)`,
			p.ID, p.Name, nullStr(p.URL), p.Price, nullStr(p.Brand), nullStr(p.Shape),
			nullStr(p.ImageFile), p.Order, string(p.Payload), string(p.Payload)); err != nil {
			return err
		}
	}

	for _, rm := range rooms {
		if _, err := tx.Exec(ctx,
			`INSERT INTO catalog_room (id, title, room, palette, area_m2, parent_id, image_file, sort_order, payload, payload_text)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9::jsonb,$10)`,
			rm.ID, rm.Title, nullStr(rm.Room), nullStr(rm.Palette), rm.AreaM2, nullStr(rm.ParentID),
			nullStr(rm.ImageFile), rm.Order, string(rm.Payload), string(rm.Payload)); err != nil {
			return err
		}
		if err := saveRoomDetails(ctx, tx, rm); err != nil {
			return err
		}
	}

	for _, e := range extras {
		if _, err := tx.Exec(ctx,
			`INSERT INTO catalog_extra (id, name, short_name, price, group_name, unit, url, image_file,
			                            sort_order, payload, payload_text)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10::jsonb,$11)`,
			e.ID, e.Name, nullStr(e.ShortName), e.Price, nullStr(e.Group), nullStr(e.Unit),
			nullStr(e.URL), nullStr(e.ImageFile), e.Order, string(e.Payload), string(e.Payload)); err != nil {
			return err
		}
	}

	for _, s := range slots {
		if _, err := tx.Exec(ctx,
			`INSERT INTO catalog_slot_image (scope, slot, image_file, sort_order, payload, payload_text)
			 VALUES ($1,$2,$3,$4,$5::jsonb,$6)`,
			s.Scope, s.Slot, nullStr(s.File), s.Order, string(s.Payload), string(s.Payload)); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func saveRoomDetails(ctx context.Context, tx pgx.Tx, rm domain.Room) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(rm.Payload, &raw); err != nil {
		return err
	}
	if colors, ok := raw["colors"]; ok {
		var list []string
		if err := json.Unmarshal(colors, &list); err == nil {
			for i, c := range list {
				if _, err := tx.Exec(ctx,
					`INSERT INTO catalog_room_color (room_id, position, color) VALUES ($1,$2,$3)
					 ON CONFLICT DO NOTHING`, rm.ID, i, c); err != nil {
					return err
				}
			}
		}
	}
	if costs, ok := raw["costs"]; ok {
		var rows [][]float64
		if err := json.Unmarshal(costs, &rows); err == nil {
			for ri, row := range rows {
				for ci, amount := range row {
					if _, err := tx.Exec(ctx,
						`INSERT INTO catalog_room_cost (room_id, row_no, column_no, amount)
						 VALUES ($1,$2,$3,$4) ON CONFLICT DO NOTHING`, rm.ID, ri, ci, amount); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// SaveCenter сохраняет интерьеры и фото центра.
func (r *Repository) SaveCenter(ctx context.Context, interiors []domain.CenterInterior, photos []domain.CenterPhoto) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `DELETE FROM center_interior`); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM center_photo`); err != nil {
		return err
	}
	for _, c := range interiors {
		if _, err := tx.Exec(ctx,
			`INSERT INTO center_interior (id, title, image_file, sort_order, payload, payload_text)
			 VALUES ($1,$2,$3,$4,$5::jsonb,$6)`,
			c.ID, c.Title, nullStr(c.Image), c.Order, string(c.Payload), string(c.Payload)); err != nil {
			return err
		}
	}
	for _, p := range photos {
		if _, err := tx.Exec(ctx,
			`INSERT INTO center_photo (id, name, image_file, sort_order, payload, payload_text)
			 VALUES ($1,$2,$3,$4,$5::jsonb,$6)`,
			p.ID, p.Name, nullStr(p.Image), p.Order, string(p.Payload), string(p.Payload)); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// SaveHotspot сохраняет hotspot-товары и аналоги.
func (r *Repository) SaveHotspot(ctx context.Context, items []domain.HotspotItem, analogues []domain.Analogue) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `DELETE FROM hotspot_item`); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM hotspot_analogue`); err != nil {
		return err
	}
	for _, h := range items {
		if _, err := tx.Exec(ctx,
			`INSERT INTO hotspot_item (id, name, category, price, unit, image_file, point_x, point_y,
			                          sort_order, payload, payload_text)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10::jsonb,$11)`,
			h.ID, h.Name, nullStr(h.Cat), h.Price, nullStr(h.Unit), nullStr(h.Image),
			h.PointX, h.PointY, h.Order, string(h.Payload), string(h.Payload)); err != nil {
			return err
		}
	}
	for _, a := range analogues {
		if _, err := tx.Exec(ctx,
			`INSERT INTO hotspot_analogue (id, name, short_name, price, group_name, unit, url, is_external,
			                              image_file, sort_order, payload, payload_text)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11::jsonb,$12)`,
			a.ID, a.Name, nullStr(a.Short), a.Price, nullStr(a.Group), nullStr(a.Unit), nullStr(a.URL),
			a.External, nullStr(a.Image), a.Order, string(a.Payload), string(a.Payload)); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// SaveReferences сохраняет референсы.
func (r *Repository) SaveReferences(ctx context.Context, refs []domain.Reference) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `DELETE FROM reference_item`); err != nil {
		return err
	}
	for _, ref := range refs {
		if _, err := tx.Exec(ctx,
			`INSERT INTO reference_item (id, title, category, image_key, price_from, price_to,
			                            sort_order, payload, payload_text)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8::jsonb,$9)`,
			ref.ID, ref.Title, nullStr(ref.Category), nullStr(ref.ImageKey), ref.PriceFrom, ref.PriceTo,
			ref.Order, string(ref.Payload), string(ref.Payload)); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// SaveConnected сохраняет подборку и её товары.
func (r *Repository) SaveConnected(ctx context.Context, pick domain.ConnectedPick, items []domain.ConnectedItem) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `DELETE FROM connected_pick`); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx,
		`INSERT INTO connected_pick (id, image_key, title, sort_order, payload, payload_text)
		 VALUES ($1,$2,$3,$4,$5::jsonb,$6)`,
		pick.ID, nullStr(pick.ImageKey), pick.Title, pick.Order, string(pick.Payload), string(pick.Payload)); err != nil {
		return err
	}
	for _, it := range items {
		if _, err := tx.Exec(ctx,
			`INSERT INTO connected_item (id, position, sku, brand, title, price, image_key, payload)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8::jsonb)`,
			pick.ID, it.Order, nullStr(it.SKU), nullStr(it.Brand), it.Title, it.Price,
			nullStr(it.ImageKey), `{}`); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// SavePlanners сохраняет планировочные решения.
func (r *Repository) SavePlanners(ctx context.Context, planners []domain.Planner) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `DELETE FROM planner`); err != nil {
		return err
	}
	for _, p := range planners {
		if _, err := tx.Exec(ctx,
			`INSERT INTO planner (id, name, area_m2, sort_order, payload, payload_text)
			 VALUES ($1,$2,$3,$4,$5::jsonb,$6)`,
			p.ID, p.Name, p.AreaM2, p.Order, string(p.Payload), string(p.Payload)); err != nil {
			return err
		}
		var raw struct {
			Rooms []struct {
				Name string  `json:"name"`
				Area float64 `json:"area"`
			} `json:"rooms"`
		}
		if err := json.Unmarshal(p.Payload, &raw); err != nil {
			return err
		}
		for i, rm := range raw.Rooms {
			if _, err := tx.Exec(ctx,
				`INSERT INTO planner_room (planner_id, position, name, area_m2) VALUES ($1,$2,$3,$4)`,
				p.ID, i, rm.Name, rm.Area); err != nil {
				return err
			}
		}
	}
	return tx.Commit(ctx)
}

func nullStr(s string) any {
	if s == "" {
		return nil
	}
	return s
}
