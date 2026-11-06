package seed

import (
	"context"
	"fmt"
	"math"

	bookingdomain "github.com/zetta/zettacalc/backend/internal/booking/domain"
	contentdomain "github.com/zetta/zettacalc/backend/internal/content/domain"
	designdomain "github.com/zetta/zettacalc/backend/internal/design/domain"
	designapp "github.com/zetta/zettacalc/backend/internal/design/app"
	estimatedomain "github.com/zetta/zettacalc/backend/internal/estimate/domain"
	identitydomain "github.com/zetta/zettacalc/backend/internal/identity/domain"
	mediadomain "github.com/zetta/zettacalc/backend/internal/media/domain"
	orderdomain "github.com/zetta/zettacalc/backend/internal/order/domain"
	techcarddomain "github.com/zetta/zettacalc/backend/internal/techcard/domain"
	tenderdomain "github.com/zetta/zettacalc/backend/internal/tender/domain"
	"github.com/zetta/zettacalc/backend/internal/platform/jsonx"
)

// ── Варианты дизайна, смета, ЦПР, ставки ───────────────────────────────────

func (l *Loader) saveDesign(ctx context.Context, refs []mediadomain.EntityRef, report *Report) error {
	type variantSeed struct {
		id, name, style, palette, image, description string
		rooms                                        []string
		budgetMin, budgetMax                         int64
		days                                         int
	}

	seeds := []variantSeed{
		{"variant_01", "Контemporари тёплый", "contemporari", "warm",
			"/api/images/file/a_main_img_0003.jpg",
			"Современный интерьер в тёплых тонах с натуральными материалами",
			[]string{"кухня", "гостиная", "спальня"}, 2500000, 4200000, 60},
		{"variant_02", "Неоклассика", "neoclassica", "cool",
			"/api/images/file/a_main_img_0004.jpg",
			"Элегантная неоклассика с мрамором и латунными акцентами",
			[]string{"кухня", "гостиная", "кабинет", "спальня"}, 3500000, 5800000, 75},
		{"variant_03", "Минимализм", "minimalist", "neutral",
			"/api/images/file/a_main_img_0005.jpg",
			"Чистые линии, встроенная мебель, скрытые системы хранения",
			[]string{"кухня-гостиная", "спальня"}, 1800000, 3200000, 45},
		{"variant_04", "Тёплый лофт", "loft", "warm",
			"/api/images/file/a_main_img_0006.jpg",
			"Лофт с мягкими текстурами, кирпичом и натуральным деревом",
			[]string{"кухня-гостиная", "спальня", "ванна"}, 2200000, 3800000, 50},
	}

	variants := make([]designdomain.Variant, 0, len(seeds))
	payloads := make(map[string][]byte, len(seeds))
	for i, s := range seeds {
		variant := designdomain.Variant{
			ID: s.id, Name: s.name, Style: s.style, Palette: s.palette,
			ImagePath: s.image, Description: s.description, Rooms: s.rooms,
			BudgetMin: s.budgetMin, BudgetMax: s.budgetMax, EstimatedDays: s.days,
			Order: i,
		}
		variants = append(variants, variant)
		payloads[s.id] = variantPayload(variant)
	}

	if err := l.designRepo.Save(ctx, variants, payloads); err != nil {
		return fmt.Errorf("seed: сохранение вариантов: %w", err)
	}
	report.DesignVariants = len(variants)
	return nil
}

// variantPayload собирает JSON варианта с тем же порядком ключей, что в mock API.
func variantPayload(v designdomain.Variant) []byte {
	obj := jsonx.NewObject()
	obj.Set("id", jsonx.NewString(v.ID))
	obj.Set("name", jsonx.NewString(v.Name))
	obj.Set("style", jsonx.NewString(v.Style))
	obj.Set("palette", jsonx.NewString(v.Palette))
	obj.Set("image", jsonx.NewString(v.ImagePath))
	obj.Set("description", jsonx.NewString(v.Description))
	rooms := make([]*jsonx.Value, 0, len(v.Rooms))
	for _, r := range v.Rooms {
		rooms = append(rooms, jsonx.NewString(r))
	}
	obj.Set("rooms", jsonx.NewArray(rooms...))
	obj.Set("estimated_budget_min", jsonx.NewInt(v.BudgetMin))
	obj.Set("estimated_budget_max", jsonx.NewInt(v.BudgetMax))
	obj.Set("estimated_days", jsonx.NewInt(int64(v.EstimatedDays)))
	return jsonx.Marshal(obj)
}

// ── Сметы и ЦПР ────────────────────────────────────────────────────────────

// staticEstimateSeed — демонстрационная смета из mock API (ESTIMATES["est_001"]).
var staticEstimateSeed = struct {
	ID       string
	Status   string
	TotalRub int64
	RangeMin int64
	RangeMax int64
	Days     int
	Lines    []estimatedomain.Line
	Assumps  []string
}{
	ID: "est_001", Status: "preliminary", TotalRub: 847500,
	RangeMin: 720000, RangeMax: 980000, Days: 45,
	Lines: []estimatedomain.Line{
		{ID: "l1", Title: "Кухня ZETTA METALICO", Qty: 1, Unit: "комплект", Price: 385000,
			SKU: "ZET-KIT-MET-001", Brand: "ZETTA"},
		{ID: "l2", Title: "Плитка 600×600 укладка", Qty: 12.5, Unit: "м²", Price: 3200},
		{ID: "l3", Title: "Штукатурка стен", Qty: 65, Unit: "м²", Price: 850},
	},
	Assumps: []string{"Ровность пола требует проверки", "Состояние стен неизвестно"},
}

func (l *Loader) saveEstimates(ctx context.Context, report *Report) error {
	estimates := make([]estimatedomain.Estimate, 0, 1)
	payloads := make(map[string][]byte, 1)

	seed := staticEstimateSeed
	estimate := estimatedomain.Estimate{
		ID: seed.ID, Status: seed.Status, TotalRub: seed.TotalRub,
		RangeMin: seed.RangeMin, RangeMax: seed.RangeMax, Days: seed.Days,
		ProjectID: "proj_001", Lines: seed.Lines, Assumptions: seed.Assumps,
	}
	estimates = append(estimates, estimate)
	payloads[seed.ID] = staticEstimatePayload(estimate)

	if err := l.estimateRepo.Save(ctx, estimates, payloads); err != nil {
		return fmt.Errorf("seed: сохранение смет: %w", err)
	}
	return nil
}

// staticEstimatePayload собирает ответ GET /api/estimate/{id} как в mock API.
func staticEstimatePayload(e estimatedomain.Estimate) []byte {
	obj := jsonx.NewObject()
	obj.Set("id", jsonx.NewString(e.ID))
	obj.Set("status", jsonx.NewString(e.Status))
	obj.Set("total_rub", jsonx.NewInt(e.TotalRub))
	obj.Set("range_min", jsonx.NewInt(e.RangeMin))
	obj.Set("range_max", jsonx.NewInt(e.RangeMax))
	obj.Set("days", jsonx.NewInt(int64(e.Days)))
	obj.Set("lines", linesPayload(e.Lines))
	obj.Set("assumptions", stringArray(e.Assumptions))
	return jsonx.Marshal(obj)
}

func linesPayload(lines []estimatedomain.Line) *jsonx.Value {
	items := make([]*jsonx.Value, 0, len(lines))
	for _, l := range lines {
		line := jsonx.NewObject()
		line.Set("id", jsonx.NewString(l.ID))
		line.Set("title", jsonx.NewString(l.Title))
		line.Set("qty", numberOrInt(l.Qty))
		line.Set("unit", jsonx.NewString(l.Unit))
		line.Set("price", numberOrInt(l.Price))
		line.Set("sku", nullableString(l.SKU))
		line.Set("brand", nullableString(l.Brand))
		items = append(items, line)
	}
	return jsonx.NewArray(items...)
}

// numberOrInt записывает целое значение без дробной части.
//
// Python печатает 1 как «1», а 12.5 как «12.5»: в mock API эти числа приходят
// из словаря, где int и float различимы, и паритет требует той же записи.
func numberOrInt(f float64) *jsonx.Value {
	if f == math.Trunc(f) && math.Abs(f) < 1e15 {
		return jsonx.NewInt(int64(f))
	}
	return jsonx.NewFloat(f)
}

func stringArray(items []string) *jsonx.Value {
	out := make([]*jsonx.Value, 0, len(items))
	for _, s := range items {
		out = append(out, jsonx.NewString(s))
	}
	return jsonx.NewArray(out...)
}

func nullableString(s string) *jsonx.Value {
	if s == "" {
		return jsonx.NewNull()
	}
	return jsonx.NewString(s)
}

// ── Техкарты ───────────────────────────────────────────────────────────────

func (l *Loader) saveTechcards(ctx context.Context, report *Report) error {
	cards := make([]techcarddomain.Techcard, 0, 2)

	first := techcarddomain.Techcard{
		ID: "tc_001", Code: "FLOOR_KER_600", Title: "Укладка плитки 600×600",
		Surface: "floor", Material: "ceramic_tile", Version: 3,
		Operations:  []string{"Демонтаж", "Выравнивание", "Гидроизоляция", "Укладка", "Затирка"},
		Consumption: map[string]float64{"klej_m2": 5.0, "grout_m2": 0.3},
		Order:       0,
	}
	first.MinAreaM2 = floatPtr(3.0)
	first.MaxUnevenness = floatPtr(3.0)
	first.Payload = techcardPayload(first)

	second := techcarddomain.Techcard{
		ID: "tc_002", Code: "WALL_PLASTER", Title: "Штукатурка гипсовая",
		Surface: "wall", Material: "gypsum_plaster", Version: 2,
		Operations:  []string{"Грунтовка", "Маяки", "Штукатурка", "Накрывка"},
		Consumption: map[string]float64{"primer_l_m2": 0.15, "plaster_kg_m2": 10.0},
		Order:       1,
	}
	second.MinAreaM2 = floatPtr(5.0)
	second.Payload = techcardPayload(second)

	cards = append(cards, first, second)
	if err := l.techcardRepo.Save(ctx, cards); err != nil {
		return fmt.Errorf("seed: сохранение техкарт: %w", err)
	}
	report.Techcards = len(cards)
	return nil
}

// techcardPayload повторяет порядок ключей ответа /api/techcards.
func techcardPayload(c techcarddomain.Techcard) []byte {
	obj := jsonx.NewObject()
	obj.Set("id", jsonx.NewString(c.ID))
	obj.Set("code", jsonx.NewString(c.Code))
	obj.Set("title", jsonx.NewString(c.Title))
	obj.Set("surface", jsonx.NewString(c.Surface))
	obj.Set("material", jsonx.NewString(c.Material))
	obj.Set("operations", stringArray(c.Operations))

	consumption := jsonx.NewObject()
	for _, name := range []string{"klej_m2", "grout_m2", "primer_l_m2", "plaster_kg_m2"} {
		if value, ok := c.Consumption[name]; ok {
			consumption.Set(name, jsonx.NewFloat(value))
		}
	}
	obj.Set("consumption", consumption)
	obj.Set("min_area_m2", jsonx.NewFloat(derefFloat(c.MinAreaM2)))
	obj.Set("max_unevenness", nullableFloat(c.MaxUnevenness))
	obj.Set("version", jsonx.NewInt(int64(c.Version)))
	return jsonx.Marshal(obj)
}

// ── Ставки подрядчиков ─────────────────────────────────────────────────────

func (l *Loader) saveTender(ctx context.Context, report *Report) error {
	type bidSeed struct {
		id, name, scope, image string
		rating, price          float64
		days                   int
	}
	seeds := []bidSeed{
		{"b1", "РемонтПро", "Кухня + санузел", "repair", 4.8, 320000, 40},
		{"b2", "СтройМастер", "Кухня + санузел", "auction", 4.5, 280000, 45},
		{"b3", "ДизайнРемонт", "Полный ремонт + дизайн", "expert", 4.9, 410000, 35},
	}

	bids := make([]tenderdomain.Bid, 0, len(seeds))
	for i, s := range seeds {
		rating, price, days := s.rating, s.price, s.days
		bid := tenderdomain.Bid{
			ID: s.id, Name: s.name, Rating: &rating, Price: &price, Days: &days,
			Scope: s.scope, ImageKey: s.image, Order: i,
		}
		bid.Payload = bidPayload(bid)
		bids = append(bids, bid)
	}

	if err := l.tenderRepo.Save(ctx, bids); err != nil {
		return fmt.Errorf("seed: сохранение ставок: %w", err)
	}
	return nil
}

// bidPayload повторяет порядок ключей ответа /api/bids.
func bidPayload(b tenderdomain.Bid) []byte {
	obj := jsonx.NewObject()
	obj.Set("id", jsonx.NewString(b.ID))
	obj.Set("name", jsonx.NewString(b.Name))
	obj.Set("rating", jsonx.NewFloat(derefFloat(b.Rating)))
	obj.Set("price", jsonx.NewInt(int64(derefFloat(b.Price))))
	obj.Set("days", jsonx.NewInt(int64(derefInt(b.Days))))
	obj.Set("scope", jsonx.NewString(b.Scope))
	obj.Set("image", jsonx.NewString(b.ImageKey))
	return jsonx.Marshal(obj)
}

// ── Корзина ────────────────────────────────────────────────────────────────

func (l *Loader) saveCart(ctx context.Context, report *Report) error {
	cart := orderdomain.Cart{
		Items: []orderdomain.Item{{
			SKU: "ZET-KIT-MET-001", Brand: "ZETTA", Title: "Кухня METALICO",
			Price: 385000, Qty: 1, ImageKey: "idea-kitchen",
		}},
		TotalRub: 385000,
	}
	if err := l.orderRepo.Save(ctx, cart); err != nil {
		return fmt.Errorf("seed: сохранение корзины: %w", err)
	}
	return nil
}

// ── Пользователи и опции записи ────────────────────────────────────────────

// demoUsers — пользователи mock API. Пароль демо-пользователя совпадает с mock
// (test123); для продакшена пользователь создаётся с bcrypt-хэшем.
var demoUsers = []identitydomain.User{
	{ID: "u1", Email: "test@zetta.ru", Name: "Тест", PasswordHash: "test123", Role: "client"},
}

func (l *Loader) saveUsers(ctx context.Context) error {
	if err := l.identityRepo.Save(ctx, demoUsers); err != nil {
		return err
	}
	return nil
}

func (l *Loader) saveBookingOptions(ctx context.Context) error {
	pool := l.opts.Pool
	if _, err := pool.Exec(ctx, `DELETE FROM booking_option`); err != nil {
		return err
	}
	for i, task := range bookingdomain.TaskOptions {
		if _, err := pool.Exec(ctx,
			`INSERT INTO booking_option (kind, value, sort_order) VALUES ('task', $1, $2)`, task, i); err != nil {
			return err
		}
	}
	for i, t := range bookingdomain.TimeOptions {
		if _, err := pool.Exec(ctx,
			`INSERT INTO booking_option (kind, value, sort_order) VALUES ('time', $1, $2)`, t, i); err != nil {
			return err
		}
	}
	return nil
}

// ── Мелкие помощники ───────────────────────────────────────────────────────

func floatPtr(v float64) *float64 { return &v }

func derefFloat(v *float64) float64 {
	if v == nil {
		return 0
	}
	return *v
}

func derefInt(v *int) int {
	if v == nil {
		return 0
	}
	return *v
}

func nullableFloat(v *float64) *jsonx.Value {
	if v == nil {
		return jsonx.NewNull()
	}
	return jsonx.NewFloat(*v)
}

var _ = designapp.EstimatePayload
var _ = contentdomain.JSONArray
