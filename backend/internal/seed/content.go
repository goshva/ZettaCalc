package seed

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	assetsdomain "github.com/zetta/zettacalc/backend/internal/assets/domain"
	contentdomain "github.com/zetta/zettacalc/backend/internal/content/domain"
	mediadomain "github.com/zetta/zettacalc/backend/internal/media/domain"
	"github.com/zetta/zettacalc/backend/internal/platform/jsonx"
)

func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

// ── Интерьеры ──────────────────────────────────────────────────────────────

func (l *Loader) saveInteriors(ctx context.Context, refs []mediadomain.EntityRef, report *Report) error {
	root, err := readJSON(filepath.Join(l.opts.DataDir, "apps52", "interiors", "dataset.json"))
	if err != nil {
		return fmt.Errorf("seed: dataset.json: %w", err)
	}

	items := make([]contentdomain.Interior, 0, 32)
	for i, raw := range field(root, "items").ArrayItems() {
		payload := compact(raw)
		items = append(items, contentdomain.Interior{
			ID:        field(raw, "id").StringValue(),
			Type:      field(raw, "type").StringValue(),
			Title:     field(raw, "title").StringValue(),
			Room:      field(raw, "room").StringValue(),
			Style:     field(raw, "style").StringValue(),
			Palette:   field(raw, "palette").StringValue(),
			ImageFile: filenameFromPath(field(raw, "image").StringValue()),
			SourceURL: field(raw, "source").StringValue(),
			AreaM2:    intPtr(field(raw, "area")),
			ParentID:  field(raw, "parent").StringValue(),
			Order:     i,
			Payload:   payload,
		})
	}

	categories := make([]contentdomain.Named, 0, 8)
	for i, raw := range field(root, "categories").ArrayItems() {
		categories = append(categories, contentdomain.Named{Name: raw.StringValue(), Order: i})
	}
	descriptions := make([]contentdomain.Named, 0, 8)
	for i, raw := range field(root, "descriptions").ArrayItems() {
		descriptions = append(descriptions, contentdomain.Named{Name: raw.StringValue(), Order: i})
	}

	if err := l.contentRepo.SaveInteriors(ctx, items, categories, descriptions); err != nil {
		return fmt.Errorf("seed: сохранение интерьеров: %w", err)
	}
	report.Interiors = len(items)
	return nil
}

// ── Каталог ────────────────────────────────────────────────────────────────

func (l *Loader) saveCatalog(ctx context.Context, refs []mediadomain.EntityRef, report *Report) error {
	root, err := readJSON(filepath.Join(l.opts.DataDir, "apps52", "catalog", "data.json"))
	if err != nil {
		return fmt.Errorf("seed: catalog/data.json: %w", err)
	}

	products := make([]contentdomain.Product, 0, 16)
	for i, raw := range field(root, "products").ArrayItems() {
		products = append(products, contentdomain.Product{
			ID:        field(raw, "id").StringValue(),
			Name:      field(raw, "name").StringValue(),
			URL:       field(raw, "url").StringValue(),
			Price:     intPtr(field(raw, "price")),
			Brand:     field(raw, "brand").StringValue(),
			Shape:     field(raw, "shape").StringValue(),
			ImageFile: filenameFromPath(field(raw, "image").StringValue()),
			Order:     i,
			Payload:   compact(raw),
		})
	}

	rooms := make([]contentdomain.Room, 0, 16)
	for i, raw := range field(root, "rooms").ArrayItems() {
		rooms = append(rooms, contentdomain.Room{
			ID:        field(raw, "id").StringValue(),
			Title:     field(raw, "title").StringValue(),
			Room:      field(raw, "room").StringValue(),
			Palette:   field(raw, "palette").StringValue(),
			AreaM2:    intPtr(field(raw, "area")),
			ParentID:  field(raw, "parent").StringValue(),
			ImageFile: filenameFromPath(field(raw, "image").StringValue()),
			Order:     i,
			Payload:   compact(raw),
		})
	}

	extras := make([]contentdomain.Extra, 0, 64)
	for i, raw := range field(root, "extra").ArrayItems() {
		extras = append(extras, contentdomain.Extra{
			ID:        field(raw, "id").StringValue(),
			Name:      field(raw, "name").StringValue(),
			ShortName: field(raw, "short").StringValue(),
			Price:     intPtr(field(raw, "price")),
			Group:     field(raw, "group").StringValue(),
			Unit:      field(raw, "unit").StringValue(),
			URL:       field(raw, "url").StringValue(),
			ImageFile: filenameFromPath(field(raw, "image").StringValue()),
			Order:     i,
			Payload:   compact(raw),
		})
	}

	slots := make([]contentdomain.SlotImage, 0, 32)
	for i, slot := range objectKeys(field(root, "photos")) {
		value := field(field(root, "photos"), slot)
		slots = append(slots, contentdomain.SlotImage{
			Scope: "photo", Slot: slot,
			File:    filenameFromPath(value.StringValue()),
			Order:   i,
			Payload: compact(value),
		})
	}
	for i, slot := range objectKeys(field(root, "scenes")) {
		value := field(field(root, "scenes"), slot)
		slots = append(slots, contentdomain.SlotImage{
			Scope: "scene", Slot: slot,
			File:    filenameFromPath(value.StringValue()),
			Order:   i,
			Payload: compact(value),
		})
	}

	if err := l.contentRepo.SaveCatalog(ctx, products, rooms, extras, slots); err != nil {
		return fmt.Errorf("seed: сохранение каталога: %w", err)
	}
	report.Products = len(products)
	report.Rooms = len(rooms)
	report.Extras = len(extras)
	return nil
}

// ── Центр ──────────────────────────────────────────────────────────────────

func (l *Loader) saveCenter(ctx context.Context, refs []mediadomain.EntityRef, report *Report) error {
	interiorsRaw, err := readJSONArray(filepath.Join(l.opts.DataDir, "apps52", "center", "interiors.json"))
	if err != nil {
		return fmt.Errorf("seed: center/interiors.json: %w", err)
	}
	photosRaw, err := readJSONArray(filepath.Join(l.opts.DataDir, "apps52", "center", "photos.json"))
	if err != nil {
		return fmt.Errorf("seed: center/photos.json: %w", err)
	}

	interiors := make([]contentdomain.CenterInterior, 0, len(interiorsRaw))
	for i, raw := range interiorsRaw {
		interiors = append(interiors, contentdomain.CenterInterior{
			ID:      field(raw, "id").StringValue(),
			Title:   field(raw, "title").StringValue(),
			Image:   filenameFromPath(field(raw, "image").StringValue()),
			Order:   i,
			Payload: compact(raw),
		})
	}
	photos := make([]contentdomain.CenterPhoto, 0, len(photosRaw))
	for i, raw := range photosRaw {
		photos = append(photos, contentdomain.CenterPhoto{
			ID:      "photo_" + itoa(i),
			Name:    field(raw, "name").StringValue(),
			Image:   filenameFromPath(field(raw, "src").StringValue()),
			Order:   i,
			Payload: compact(raw),
		})
	}

	if err := l.contentRepo.SaveCenter(ctx, interiors, photos); err != nil {
		return fmt.Errorf("seed: сохранение центра: %w", err)
	}
	report.CenterInteriors = len(interiors)
	report.CenterPhotos = len(photos)
	return nil
}

// ── Hotspot ────────────────────────────────────────────────────────────────

func (l *Loader) saveHotspot(ctx context.Context, refs []mediadomain.EntityRef, report *Report) error {
	itemsRaw, err := readJSONArray(filepath.Join(l.opts.DataDir, "proto_a_hotspotData60.json"))
	if err != nil {
		return fmt.Errorf("seed: hotspot: %w", err)
	}
	analoguesRaw, err := readJSONArray(filepath.Join(l.opts.DataDir, "proto_a_extraAnalogues64.json"))
	if err != nil {
		return fmt.Errorf("seed: аналоги: %w", err)
	}

	items := make([]contentdomain.HotspotItem, 0, len(itemsRaw))
	for i, raw := range itemsRaw {
		point := field(raw, "point").ArrayItems()
		item := contentdomain.HotspotItem{
			ID:      field(raw, "id").StringValue(),
			Name:    field(raw, "name").StringValue(),
			Cat:     field(raw, "category").StringValue(),
			Price:   intPtr(field(raw, "price")),
			Unit:    field(raw, "unit").StringValue(),
			Image:   filenameFromPath(field(raw, "image").StringValue()),
			Order:   i,
			Payload: compact(raw),
		}
		if len(point) > 0 {
			item.PointX = intPtr(point[0])
		}
		if len(point) > 1 {
			item.PointY = intPtr(point[1])
		}
		items = append(items, item)
	}

	analogues := make([]contentdomain.Analogue, 0, len(analoguesRaw))
	for i, raw := range analoguesRaw {
		image := field(raw, "image").StringValue()
		file := filenameFromPath(image)
		analogues = append(analogues, contentdomain.Analogue{
			ID:       field(raw, "id").StringValue(),
			Name:     field(raw, "name").StringValue(),
			Short:    field(raw, "short").StringValue(),
			Price:    intPtr(field(raw, "price")),
			Group:    field(raw, "group").StringValue(),
			Unit:     field(raw, "unit").StringValue(),
			URL:      field(raw, "url").StringValue(),
			External: field(raw, "external").BoolValue(),
			Image:    file,
			Order:    i,
			Payload:  compact(raw),
		})
	}

	if err := l.contentRepo.SaveHotspot(ctx, items, analogues); err != nil {
		return fmt.Errorf("seed: сохранение hotspot: %w", err)
	}
	report.HotspotItems = len(items)
	report.Analogues = len(analogues)
	return nil
}

// ── Референсы, подборка, планировки ────────────────────────────────────────

// referenceSeed — референсы из mock API (в файлах их нет, они были в коде).
var referenceSeed = []struct {
	ID        string
	Title     string
	Category  string
	ImageKey  string
	Styles    []string
	PriceFrom int64
	PriceTo   int64
}{
	{"ref_001", "Кухня в тёплых тонах с островом", "kitchen", "idea-kitchen",
		[]string{"modern", "warm"}, 450000, 1200000},
	{"ref_002", "Гостиная в тёплых тонах", "living", "idea-living",
		[]string{"modern", "cozy"}, 320000, 890000},
	{"ref_003", "Санузел в минимализме", "bathroom", "idea-bath",
		[]string{"minimalist", "modern"}, 180000, 520000},
	{"ref_004", "Спальня с мягким изголовьем", "bedroom", "idea-bedroom",
		[]string{"modern", "soft"}, 250000, 680000},
	{"ref_005", "Подростковая", "bedroom", "idea-teen",
		[]string{"modern", "bright"}, 210000, 540000},
	{"ref_006", "Кабинет и библиотека", "living", "idea-study",
		[]string{"classical", "calm"}, 280000, 720000},
}

func (l *Loader) saveReferences(ctx context.Context, report *Report) error {
	refs := make([]contentdomain.Reference, 0, len(referenceSeed))
	for i, seed := range referenceSeed {
		from, to := float64(seed.PriceFrom), float64(seed.PriceTo)
		refs = append(refs, contentdomain.Reference{
			ID:        seed.ID,
			Title:     seed.Title,
			Category:  seed.Category,
			ImageKey:  seed.ImageKey,
			PriceFrom: &from,
			PriceTo:   &to,
			Order:     i,
			Payload:   referencePayload(seed),
		})
	}
	if err := l.contentRepo.SaveReferences(ctx, refs); err != nil {
		return fmt.Errorf("seed: сохранение референсов: %w", err)
	}
	report.References = len(refs)
	return nil
}

// referencePayload повторяет порядок ключей ответа /api/references.
func referencePayload(seed struct {
	ID        string
	Title     string
	Category  string
	ImageKey  string
	Styles    []string
	PriceFrom int64
	PriceTo   int64
}) []byte {
	obj := jsonx.NewObject()
	obj.Set("id", jsonx.NewString(seed.ID))
	obj.Set("title", jsonx.NewString(seed.Title))
	obj.Set("category", jsonx.NewString(seed.Category))
	obj.Set("image", jsonx.NewString(seed.ImageKey))
	obj.Set("styles", stringArray(seed.Styles))
	obj.Set("price_from", jsonx.NewInt(seed.PriceFrom))
	obj.Set("price_to", jsonx.NewInt(seed.PriceTo))
	return jsonx.Marshal(obj)
}

// plannerSeed — планировки из mock API: area квартиры float, площади комнат int.
var plannerSeed = []struct {
	ID    string
	Name  string
	Area  float64
	Rooms []plannerRoomSeed
}{
	{"p1", "Квартира 90 м² — стандарт", 90.0, []plannerRoomSeed{
		{"Кухня-гостиная", 28}, {"Спальня", 18},
	}},
	{"p2", "Квартира 90 м² — перепланировка", 92.0, []plannerRoomSeed{
		{"Кухня-гостиная", 35}, {"Спальня", 18},
	}},
}

type plannerRoomSeed struct {
	Name string
	Area int64
}

func (l *Loader) savePlanners(ctx context.Context, report *Report) error {
	planners := make([]contentdomain.Planner, 0, len(plannerSeed))
	for i, seed := range plannerSeed {
		area := seed.Area
		planners = append(planners, contentdomain.Planner{
			ID: seed.ID, Name: seed.Name, AreaM2: &area, Order: i,
			Payload: plannerPayload(seed),
		})
	}
	if err := l.contentRepo.SavePlanners(ctx, planners); err != nil {
		return fmt.Errorf("seed: сохранение планировок: %w", err)
	}
	report.Planners = len(planners)
	return nil
}

// plannerPayload повторяет порядок ключей ответа /api/planners.
func plannerPayload(seed struct {
	ID    string
	Name  string
	Area  float64
	Rooms []plannerRoomSeed
}) []byte {
	rooms := make([]*jsonx.Value, 0, len(seed.Rooms))
	for _, r := range seed.Rooms {
		room := jsonx.NewObject()
		room.Set("name", jsonx.NewString(r.Name))
		room.Set("area", jsonx.NewInt(r.Area))
		rooms = append(rooms, room)
	}

	obj := jsonx.NewObject()
	obj.Set("id", jsonx.NewString(seed.ID))
	obj.Set("name", jsonx.NewString(seed.Name))
	obj.Set("area", jsonx.NewFloat(seed.Area))
	obj.Set("rooms", jsonx.NewArray(rooms...))
	return jsonx.Marshal(obj)
}

// ── Ресурсы приложений и прототипов ────────────────────────────────────────

func (l *Loader) saveAssets(ctx context.Context, report *Report) error {
	files := make([]assetsFile, 0, 128)

	// CSS и разметка приложений прототипа А
	appsDir := filepath.Join(l.opts.DataDir, "apps52")
	for app, entry := range map[string][2]string{
		"main":      {"", "main.css"},
		"interiors": {"interiors.html", "interiors.css"},
		"catalog":   {"catalog.html", "catalog.css"},
		"center":    {"center.html", "center.css"},
		"repair":    {"repair.html", "repair.css"},
		"about":     {"about.html", "about.css"},
	} {
		for _, name := range entry {
			if name == "" {
				continue
			}
			if err := l.addContentFile(&files, appsDir, name, "app", app, ""); err != nil {
				return err
			}
		}
	}

	// Секции главной страницы прототипа А
	mainPageDir := filepath.Join(l.opts.DataDir, "main_page")
	names, err := listHTML(mainPageDir)
	if err != nil {
		return fmt.Errorf("seed: main_page: %w", err)
	}
	for _, name := range names {
		if err := l.addContentFile(&files, mainPageDir, name, "main_section", "", ""); err != nil {
			return err
		}
	}

	// Фрагменты разметки прототипов A и B
	for _, proto := range []string{"a", "b"} {
		dir := filepath.Join(l.opts.DataDir, proto+"_page")
		names, err := listHTML(dir)
		if err != nil {
			return fmt.Errorf("seed: %s_page: %w", proto, err)
		}
		for _, name := range names {
			if err := l.addContentFile(&files, dir, name, "page", "", proto); err != nil {
				return err
			}
		}
	}

	// Таблицы стилей прототипов
	cssDir := filepath.Join(l.opts.DataDir, "css")
	names, err = listFiles(cssDir, ".css")
	if err != nil {
		return fmt.Errorf("seed: css: %w", err)
	}
	for _, name := range names {
		if err := l.addContentFile(&files, cssDir, name, "css", "", ""); err != nil {
			return err
		}
	}

	sort.SliceStable(files, func(i, j int) bool {
		if files[i].kind != files[j].kind {
			return files[i].kind < files[j].kind
		}
		return files[i].path < files[j].path
	})

	out := make([]assetsdomain.File, 0, len(files))
	for i, f := range files {
		out = append(out, assetsdomain.File{
			Path: f.path, Kind: assetsdomain.Kind(f.kind), AppName: f.app,
			Proto: f.proto, Name: f.name,
			ContentType: f.contentType, Body: f.body, Order: i,
		})
	}
	if err := l.assetsRepo.Save(ctx, out); err != nil {
		return fmt.Errorf("seed: сохранение ресурсов: %w", err)
	}
	report.ContentFiles = len(out)
	return nil
}

type assetsFile struct {
	path, kind, app, proto, name, contentType, body string
}

func (l *Loader) addContentFile(dst *[]assetsFile, dir, name, kind, app, proto string) error {
	body, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("seed: чтение %s: %w", name, err)
	}
	text := string(body)
	// Секции главной страницы mock API читает через read_text(), а он
	// нормализует переводы строк (universal newlines). Остальные ресурсы
	// отдаются FileResponse как есть, поэтому нормализация только здесь.
	if kind == "main_section" {
		text = normalizeNewlines(text)
	}
	contentType := "text/html; charset=utf-8"
	if strings.HasSuffix(name, ".css") {
		contentType = "text/css; charset=utf-8"
	}
	*dst = append(*dst, assetsFile{
		path:        filepath.ToSlash(filepath.Join(kind, proto, app, name)),
		kind:        kind,
		app:         app,
		proto:       proto,
		name:        strings.TrimSuffix(name, filepath.Ext(name)),
		contentType: contentType,
		body:        text,
	})
	return nil
}

// normalizeNewlines повторяет universal newlines из Python: \r\n и \r → \n.
func normalizeNewlines(s string) string {
	if !strings.ContainsRune(s, '\r') {
		return s
	}
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.ReplaceAll(s, "\r", "\n")
}

func listHTML(dir string) ([]string, error) { return listFiles(dir, ".html") }

func listFiles(dir, ext string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ext) {
			continue
		}
		out = append(out, e.Name())
	}
	sort.Strings(out)
	return out, nil
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
