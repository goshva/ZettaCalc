// Package seed переносит данные mock API в PostgreSQL.
//
// Загрузчик намеренно «тупой»: он читает исходные файлы и сохраняет их
// представление, не пытаясь улучшить структуру. Паритет ответов с Python-mock
// важнее любой нормализации, поэтому для каждой сущности сохраняется готовый
// JSON (payload_text) и он же отдаётся клиенту.
package seed

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	assetsdomain "github.com/zetta/zettacalc/backend/internal/assets/domain"
	bookingdomain "github.com/zetta/zettacalc/backend/internal/booking/domain"
	contentdomain "github.com/zetta/zettacalc/backend/internal/content/domain"
	designdomain "github.com/zetta/zettacalc/backend/internal/design/domain"
	estimatedomain "github.com/zetta/zettacalc/backend/internal/estimate/domain"
	identitydomain "github.com/zetta/zettacalc/backend/internal/identity/domain"
	mediadomain "github.com/zetta/zettacalc/backend/internal/media/domain"
	orderdomain "github.com/zetta/zettacalc/backend/internal/order/domain"
	techcarddomain "github.com/zetta/zettacalc/backend/internal/techcard/domain"
	tenderdomain "github.com/zetta/zettacalc/backend/internal/tender/domain"

	"github.com/zetta/zettacalc/backend/internal/platform/jsonx"
)

// Options — источники данных для загрузки.
type Options struct {
	DataDir   string
	ImagesDir string
	Pool      *pgxpool.Pool
}

// Report — что именно перенесено.
type Report struct {
	Images          int
	ImageBytes      int64
	ImageKeys       int
	EntityRefs      int
	Interiors       int
	Products        int
	Rooms           int
	Extras          int
	CenterInteriors int
	CenterPhotos    int
	HotspotItems    int
	Analogues       int
	References      int
	Planners        int
	Techcards       int
	DesignVariants  int
	ContentFiles    int
}

// Loader выполняет загрузку данных.
type Loader struct {
	opts Options

	mediaRepo    mediadomain.Repository
	contentRepo  contentdomain.SavePort
	assetsRepo   assetsdomain.Repository
	designRepo   designRepository
	estimateRepo estimateRepository
	techcardRepo techcardRepository
	tenderRepo   tenderRepository
	bookingRepo  bookingRepository
	identityRepo identityRepository
	orderRepo    orderRepository
}

type designRepository interface {
	Save(ctx context.Context, variants []designdomain.Variant, payloads map[string][]byte) error
}

type estimateRepository interface {
	Save(ctx context.Context, estimates []estimatedomain.Estimate, payloads map[string][]byte) error
}

type techcardRepository interface {
	Save(ctx context.Context, cards []techcarddomain.Techcard) error
}

type tenderRepository interface {
	Save(ctx context.Context, bids []tenderdomain.Bid) error
}

type bookingRepository interface {
	Save(ctx context.Context, a bookingdomain.Appointment) error
}

type identityRepository interface {
	Save(ctx context.Context, users []identitydomain.User) error
}

type orderRepository interface {
	Save(ctx context.Context, cart orderdomain.Cart) error
}

// New собирает загрузчик.
func New(opts Options, mediaRepo mediadomain.Repository, contentRepo contentdomain.SavePort,
	assetsRepo assetsdomain.Repository, designRepo designRepository,
	estimateRepo estimateRepository, techcardRepo techcardRepository,
	tenderRepo tenderRepository, bookingRepo bookingRepository,
	identityRepo identityRepository, orderRepo orderRepository) *Loader {
	return &Loader{
		opts:         opts,
		mediaRepo:    mediaRepo,
		contentRepo:  contentRepo,
		assetsRepo:   assetsRepo,
		designRepo:   designRepo,
		estimateRepo: estimateRepo,
		techcardRepo: techcardRepo,
		tenderRepo:   tenderRepo,
		bookingRepo:  bookingRepo,
		identityRepo: identityRepo,
		orderRepo:    orderRepo,
	}
}

// Run переносит все данные и возвращает отчёт.
func (l *Loader) Run(ctx context.Context) (Report, error) {
	var report Report

	if err := l.saveBookingOptions(ctx); err != nil {
		return report, fmt.Errorf("seed: опции записи: %w", err)
	}
	if err := l.saveUsers(ctx); err != nil {
		return report, fmt.Errorf("seed: пользователи: %w", err)
	}

	refs, err := l.saveImages(ctx, &report)
	if err != nil {
		return report, err
	}

	if err := l.saveInteriors(ctx, refs, &report); err != nil {
		return report, err
	}
	if err := l.saveCatalog(ctx, refs, &report); err != nil {
		return report, err
	}
	if err := l.saveCenter(ctx, refs, &report); err != nil {
		return report, err
	}
	if err := l.saveHotspot(ctx, refs, &report); err != nil {
		return report, err
	}
	if err := l.saveReferences(ctx, &report); err != nil {
		return report, err
	}
	if err := l.savePlanners(ctx, &report); err != nil {
		return report, err
	}
	if err := l.saveDesign(ctx, refs, &report); err != nil {
		return report, err
	}
	if err := l.saveEstimates(ctx, &report); err != nil {
		return report, err
	}
	if err := l.saveTechcards(ctx, &report); err != nil {
		return report, err
	}
	if err := l.saveTender(ctx, &report); err != nil {
		return report, err
	}
	if err := l.saveCart(ctx, &report); err != nil {
		return report, err
	}
	if err := l.saveAssets(ctx, &report); err != nil {
		return report, err
	}
	if err := l.saveMeta(ctx, &report); err != nil {
		return report, err
	}
	if err := l.mediaRepo.SaveRefs(ctx, refs); err != nil {
		return report, fmt.Errorf("seed: привязки картинок: %w", err)
	}
	report.EntityRefs = len(refs)
	return report, nil
}

// ── Медиа ──────────────────────────────────────────────────────────────────

// entityKeyOrder — ключи PIM картинок сущностей в порядке объявления mock API.
//
// Порядок важен: GET /api/images отдаёт словарь ключей, и паритет проверяется
// сравнением тел ответов, поэтому итерация по map здесь недопустима.
var entityKeyOrder = []string{
	"idea-kitchen", "idea-living", "idea-bath", "idea-bedroom", "idea-teen", "idea-study",
	"cat-furniture", "cat-materials", "cat-bath", "cat-decor", "connected",
	"sofa", "pendant", "rug",
	"c-sofa", "c-dining", "c-lounge", "c-wardrobe", "c-bedroom",
	"c-floors", "c-stone", "c-decor", "c-sockets", "c-dark",
}

// staticKeyOrder — ключи общей статики сайта в порядке объявления mock API.
var staticKeyOrder = []string{
	"hero", "video", "app", "cpr", "expert", "auction", "ceo", "plan", "repair",
	"center-store", "center-pano", "center-kitchen", "center-stone", "center-wardrobe",
	"ds-e06b4936da",
}

// entityKeyFiles — картинки сущностей, адресуемые по ключу PIM (source=db).
var entityKeyFiles = map[string]string{
	"idea-kitchen":  "idea-kitchen.jpg",
	"idea-living":   "idea-living.jpg",
	"idea-bath":     "idea-bath.jpg",
	"idea-bedroom":  "idea-bedroom.jpg",
	"idea-teen":     "idea-teen.jpg",
	"idea-study":    "idea-study.jpg",
	"cat-furniture": "cat-furniture.jpg",
	"cat-materials": "cat-materials.jpg",
	"cat-bath":      "cat-bath.jpg",
	"cat-decor":     "cat-decor.jpg",
	"connected":     "connected.jpg",
	"sofa":          "sofa.jpg",
	"pendant":       "pendant.jpg",
	"rug":           "rug.jpg",
	"c-sofa":        "c-sofa.jpg",
	"c-dining":      "c-dining.jpg",
	"c-lounge":      "c-lounge.jpg",
	"c-wardrobe":    "c-wardrobe.jpg",
	"c-bedroom":     "c-bedroom.jpg",
	"c-floors":      "c-floors.jpg",
	"c-stone":       "c-stone.jpg",
	"c-decor":       "c-decor.jpg",
	"c-sockets":     "c-sockets.jpg",
	"c-dark":        "c-dark.jpg",
}

// staticKeyFiles — общая статика сайта: остаётся во фронтендах, в БД не пишется.
var staticKeyFiles = map[string]string{
	"hero":            "hero.jpg",
	"video":           "video.jpg",
	"app":             "app.jpg",
	"cpr":             "cpr.jpg",
	"expert":          "expert.jpg",
	"auction":         "auction.jpg",
	"ceo":             "ceo.jpg",
	"plan":            "plan.jpg",
	"repair":          "repair.jpg",
	"center-store":    "center-store.jpg",
	"center-pano":     "center-pano.jpg",
	"center-kitchen":  "center-kitchen.jpg",
	"center-stone":    "center-stone.jpg",
	"center-wardrobe": "center-wardrobe.jpg",
	"ds-e06b4936da":   "ds-e06b4936da.jpg",
}

// saveImages переносит файлы картинок сущностей и ключи PIM.
func (l *Loader) saveImages(ctx context.Context, report *Report) ([]mediadomain.EntityRef, error) {
	entries, err := os.ReadDir(l.opts.ImagesDir)
	if err != nil {
		return nil, fmt.Errorf("seed: чтение каталога картинок %s: %w", l.opts.ImagesDir, err)
	}

	assets := make([]mediadomain.Asset, 0, len(entries))
	var refs []mediadomain.EntityRef

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		group, isEntity := entityGroup(name)
		if !isEntity {
			// Общая статика сайта: файл остаётся во фронтендах, в БД только запись
			// о нём (storage = client, bytes = NULL).
			if _, ok := staticFileByName(name); ok {
				assets = append(assets, mediadomain.Asset{
					Filename:    name,
					ContentType: contentTypeFor(name),
					Storage:     mediadomain.StorageClient,
					GroupPrefix: "static",
				})
			}
			continue
		}

		body, err := os.ReadFile(filepath.Join(l.opts.ImagesDir, name))
		if err != nil {
			return nil, fmt.Errorf("seed: чтение %s: %w", name, err)
		}
		assets = append(assets, mediadomain.Asset{
			Filename:    name,
			ContentType: contentTypeFor(name),
			ByteSize:    int64(len(body)),
			SHA256:      sha256Hex(string(body)),
			Storage:     mediadomain.StorageEntity,
			GroupPrefix: group,
			Bytes:       body,
		})
		report.Images++
		report.ImageBytes += int64(len(body))
	}

	keys := make([]mediadomain.Key, 0, len(entityKeyOrder)+len(staticKeyOrder))
	order := 0
	for _, key := range entityKeyOrder {
		file := entityKeyFiles[key]
		if _, err := os.Stat(filepath.Join(l.opts.ImagesDir, file)); err != nil {
			continue
		}
		keys = append(keys, mediadomain.Key{Key: key, Filename: file, Source: "db", Storage: "db", Order: order})
		order++
	}
	for _, key := range staticKeyOrder {
		keys = append(keys, mediadomain.Key{
			Key: key, Filename: staticKeyFiles[key], Source: "static", Storage: "client", Order: order,
		})
		order++
	}
	if err := l.mediaRepo.SaveAssets(ctx, assets); err != nil {
		return nil, fmt.Errorf("seed: сохранение картинок: %w", err)
	}
	if err := l.mediaRepo.SaveKeys(ctx, keys); err != nil {
		return nil, fmt.Errorf("seed: сохранение ключей: %w", err)
	}
	report.ImageKeys = len(keys)

	refs = append(refs, imageRefsFromData(l.opts.DataDir)...)
	return refs, nil
}

// saveMeta сохраняет метаданные развёртывания.
//
// Каталоги фрагментов попадают в ответ /api/page, а число секций главной
// страницы — в /api/health: оба значения зависят от каталога загрузки.
func (l *Loader) saveMeta(ctx context.Context, report *Report) error {
	abs := func(dir string) string {
		if resolved, err := filepath.Abs(dir); err == nil {
			return resolved
		}
		return dir
	}
	values := map[string]string{
		"page_dir_a": abs(filepath.Join(l.opts.DataDir, "a_page")),
		"page_dir_b": abs(filepath.Join(l.opts.DataDir, "b_page")),
	}
	if err := l.assetsRepo.SaveMeta(ctx, values); err != nil {
		return fmt.Errorf("seed: сохранение метаданных: %w", err)
	}
	return nil
}

// entityGroup определяет, относится ли файл к сущностям, и к какой группе.
//
// Карты ключей адресуются ключом PIM (idea-kitchen), а на диске лежит файл
// (idea-kitchen.jpg), поэтому имя файла сверяется со значениями карты.
func entityGroup(name string) (string, bool) {
	switch {
	case strings.HasSuffix(name, ".jpg") && strings.HasPrefix(name, "a_") && strings.Contains(name, "_img_"):
		return strings.SplitN(name, "_img_", 2)[0], true
	case strings.HasPrefix(name, "b_img_"):
		return "b", true
	}
	if _, ok := keyForEntityFile(name); ok {
		return "pim", true
	}
	return "", false
}

// keyForEntityFile ищет ключ PIM по имени файла картинки сущности.
func keyForEntityFile(file string) (string, bool) {
	for key, name := range entityKeyFiles {
		if name == file {
			return key, true
		}
	}
	return "", false
}

// staticFileByName сообщает, относится ли файл к общей статике сайта.
func staticFileByName(name string) (string, bool) {
	for key, file := range staticKeyFiles {
		if file == name {
			return key, true
		}
	}
	return "", false
}

func contentTypeFor(name string) string {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	case ".svg":
		return "image/svg+xml"
	case ".gif":
		return "image/gif"
	default:
		return "image/jpeg"
	}
}

// imageRefsFromData собирает связи «картинка ↔ сущность» из исходных данных.
//
// Это и есть ответ на вопрос «какие картинки принадлежат сущностям»: связь
// берётся не из имени файла, а из ссылок в данных прототипов.
func imageRefsFromData(dataDir string) []mediadomain.EntityRef {
	var refs []mediadomain.EntityRef

	add := func(entityType, entityID, role string, position int, raw string) {
		name := filenameFromPath(raw)
		if name == "" {
			return
		}
		refs = append(refs, mediadomain.EntityRef{
			Filename: name, EntityType: entityType, EntityID: entityID,
			Role: role, Position: position,
		})
	}

	// dataset.json: интерьеры (image, gallery, plan)
	if dataset, err := readJSON(filepath.Join(dataDir, "apps52", "interiors", "dataset.json")); err == nil {
		for _, item := range field(dataset, "items").ArrayItems() {
			id := field(item, "id").StringValue()
			add("interior", id, "primary", 0, field(item, "image").StringValue())
			for i, g := range field(item, "gallery").ArrayItems() {
				add("interior", id, "gallery", i, g.StringValue())
			}
			add("interior", id, "plan", 0, field(item, "plan").StringValue())
		}
	}

	// catalog/data.json: товары, комнаты, офферы, extra, photos, scenes
	if catalog, err := readJSON(filepath.Join(dataDir, "apps52", "catalog", "data.json")); err == nil {
		for _, p := range field(catalog, "products").ArrayItems() {
			add("product", field(p, "id").StringValue(), "image", 0, field(p, "image").StringValue())
		}
		for _, rm := range field(catalog, "rooms").ArrayItems() {
			add("room", field(rm, "id").StringValue(), "image", 0, field(rm, "image").StringValue())
		}
		for _, o := range field(catalog, "offers").ArrayItems() {
			for i, img := range field(o, "images").ArrayItems() {
				add("offer", field(o, "id").StringValue(), "image", i, img.StringValue())
			}
		}
		for _, e := range field(catalog, "extra").ArrayItems() {
			add("extra", field(e, "id").StringValue(), "image", 0, field(e, "image").StringValue())
		}
		for _, slot := range objectKeys(field(catalog, "photos")) {
			add("catalog_photo", slot, "image", 0, field(field(catalog, "photos"), slot).StringValue())
		}
		for _, slot := range objectKeys(field(catalog, "scenes")) {
			add("catalog_scene", slot, "image", 0, field(field(catalog, "scenes"), slot).StringValue())
		}
	}

	// center: интерьеры и фото
	if center, err := readJSONArray(filepath.Join(dataDir, "apps52", "center", "interiors.json")); err == nil {
		for _, item := range center {
			add("center_interior", field(item, "id").StringValue(), "image", 0, field(item, "image").StringValue())
		}
	}
	if photos, err := readJSONArray(filepath.Join(dataDir, "apps52", "center", "photos.json")); err == nil {
		for i, item := range photos {
			add("center_photo", field(item, "name").StringValue(), "image", i, field(item, "src").StringValue())
		}
	}

	// hotspot и аналоги
	if hotspot, err := readJSONArray(filepath.Join(dataDir, "proto_a_hotspotData60.json")); err == nil {
		for _, item := range hotspot {
			add("hotspot", field(item, "id").StringValue(), "image", 0, field(item, "image").StringValue())
		}
	}
	if analogues, err := readJSONArray(filepath.Join(dataDir, "proto_a_extraAnalogues64.json")); err == nil {
		for _, item := range analogues {
			add("analogue", field(item, "id").StringValue(), "image", 0, field(item, "image").StringValue())
		}
	}

	return refs
}

// filenameFromPath достаёт имя файла из /api/images/file/<name>.
func filenameFromPath(path string) string {
	const marker = "/api/images/file/"
	if !strings.HasPrefix(path, marker) {
		return ""
	}
	name := strings.TrimPrefix(path, marker)
	if strings.ContainsAny(name, `/\`) || strings.Contains(name, "..") {
		return ""
	}
	return name
}

// ── Чтение JSON ────────────────────────────────────────────────────────────

func readJSON(path string) (*jsonx.Value, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return jsonx.Parse(body)
}

func readJSONArray(path string) ([]*jsonx.Value, error) {
	v, err := readJSON(path)
	if err != nil {
		return nil, err
	}
	if v.Kind != jsonx.Array {
		return nil, errors.New("seed: ожидался JSON-массив в " + path)
	}
	return v.Items, nil
}

func objectKeys(v *jsonx.Value) []string {
	if v == nil || v.Kind != jsonx.Object {
		return nil
	}
	out := make([]string, 0, len(v.Keys))
	out = append(out, v.Keys...)
	return out
}

// field возвращает поле JSON-объекта; nil безопасен для всех геттеров jsonx.
func field(v *jsonx.Value, key string) *jsonx.Value {
	if v == nil {
		return nil
	}
	return v.Field(key)
}

// compact нормализует JSON так же, как это делает Python-mock.
func compact(v *jsonx.Value) []byte { return jsonx.Marshal(v) }

// marshalCompact сериализует структуру Go в компактный JSON без экранирования.
func marshalCompact(v any) ([]byte, error) {
	buf := &strings.Builder{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return []byte(strings.TrimRight(buf.String(), "\n")), nil
}

func intPtr(v *jsonx.Value) *float64 {
	if v == nil || v.Kind == jsonx.Null {
		return nil
	}
	f, ok := v.FloatValue()
	if !ok {
		return nil
	}
	return &f
}

func strOr(v *jsonx.Value, fallback string) string {
	if v == nil || v.Kind == jsonx.Null {
		return fallback
	}
	return v.StringValue()
}

var _ = fs.ValidPath
