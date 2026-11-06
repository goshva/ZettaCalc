// Package domain контекста Content: интерьеры, каталог, центр, hotspot.
//
// Все сущности хранят исходный JSON (Payload). Это осознанное решение: цель
// миграции — паритет с mock API байт-в-байт, а не «более красивая» схема.
// Нормализованные колонки нужны для фильтров и связей с картинками.
package domain

import (
	"context"
	"encoding/json"
	"errors"
)

// ErrNotFound — сущность отсутствует.
var ErrNotFound = errors.New("content: сущность не найдена")

// Payload — исходное представление сущности из mock API.
type Payload struct {
	// Text — компактный JSON сущности, отдаётся клиенту как есть.
	Text []byte
}

// InteriorFilter — фильтры списка интерьеров.
type InteriorFilter struct {
	Category string
	Room     string
	Style    string
}

// Interior — интерьерная идея, комната или квартира из dataset.json.
type Interior struct {
	ID        string
	Type      string
	Title     string
	Room      string
	Style     string
	Palette   string
	ImageFile string
	SourceURL string
	AreaM2    *float64
	ParentID  string
	Order     int
	Payload   []byte
}

// Named — категория или описание интерьеров.
type Named struct {
	Name  string
	Order int
}

// Product — товар каталога.
type Product struct {
	ID        string
	Name      string
	URL       string
	Price     *float64
	Brand     string
	Shape     string
	ImageFile string
	Order     int
	Payload   []byte
}

// Room — комната для комплектации.
type Room struct {
	ID        string
	Title     string
	Room      string
	Palette   string
	AreaM2    *float64
	ParentID  string
	ImageFile string
	Order     int
	Payload   []byte
}

// Extra — дополнительная позиция каталога (напольные покрытия и пр.).
type Extra struct {
	ID        string
	Name      string
	ShortName string
	Price     *float64
	Group     string
	Unit      string
	URL       string
	ImageFile string
	Order     int
	Payload   []byte
}

// SlotImage — картинка слота каталога (photos/scenes).
type SlotImage struct {
	Scope   string
	Slot    string
	File    string
	Order   int
	Payload []byte
}

// CenterInterior — интерьер Интерьерного центра.
type CenterInterior struct {
	ID      string
	Title   string
	Image   string
	Order   int
	Payload []byte
}

// CenterPhoto — фото центра.
type CenterPhoto struct {
	ID      string
	Name    string
	Image   string
	Order   int
	Payload []byte
}

// HotspotItem — товар на hotspot-подборке.
type HotspotItem struct {
	ID      string
	Name    string
	Cat     string
	Price   *float64
	Unit    string
	Image   string
	PointX  *float64
	PointY  *float64
	Order   int
	Payload []byte
}

// Analogue — аналог товара.
type Analogue struct {
	ID       string
	Name     string
	Short    string
	Price    *float64
	Group    string
	Unit     string
	URL      string
	External bool
	Image    string
	Order    int
	Payload  []byte
}

// Reference — референс интерьера для старых эндпоинтов.
type Reference struct {
	ID        string
	Title     string
	Category  string
	ImageKey  string
	PriceFrom *float64
	PriceTo   *float64
	Order     int
	Payload   []byte
}

// ConnectedPick — подборка «точно к интерьеру».
type ConnectedPick struct {
	ID       string
	ImageKey string
	Title    string
	Order    int
	Payload  []byte
}

// ConnectedItem — товар внутри подборки.
type ConnectedItem struct {
	SKU      string
	Brand    string
	Title    string
	Price    *float64
	ImageKey string
	Order    int
}

// Planner — планировочное решение.
type Planner struct {
	ID      string
	Name    string
	AreaM2  *float64
	Order   int
	Payload []byte
}

// Repository — порт чтения контента.
type Repository interface {
	Interiors(ctx context.Context, f InteriorFilter) ([]Interior, error)
	InteriorByID(ctx context.Context, id string) (Interior, error)
	InteriorCategories(ctx context.Context) ([]Named, error)
	InteriorDescriptions(ctx context.Context) ([]Named, error)

	Products(ctx context.Context, group string) ([]Product, error)
	CatalogRooms(ctx context.Context) ([]Room, error)
	CatalogRoomByID(ctx context.Context, id string) (Room, error)

	CenterInteriors(ctx context.Context) ([]CenterInterior, error)
	CenterPhotos(ctx context.Context) ([]CenterPhoto, error)

	HotspotItems(ctx context.Context) ([]HotspotItem, error)
	AnalogueItems(ctx context.Context) ([]Analogue, error)

	References(ctx context.Context, category string) ([]Reference, error)
	ReferenceByID(ctx context.Context, id string) (Reference, error)

	Connected(ctx context.Context) (ConnectedPick, []ConnectedItem, error)
	Planners(ctx context.Context) ([]Planner, error)
}
// SavePort — порт записи, нужен только загрузчику данных.
type SavePort interface {
	SaveInteriors(ctx context.Context, items []Interior, categories, descriptions []Named) error
	SaveCatalog(ctx context.Context, products []Product, rooms []Room, extras []Extra, slots []SlotImage) error
	SaveCenter(ctx context.Context, interiors []CenterInterior, photos []CenterPhoto) error
	SaveHotspot(ctx context.Context, items []HotspotItem, analogues []Analogue) error
	SaveReferences(ctx context.Context, refs []Reference) error
	SaveConnected(ctx context.Context, pick ConnectedPick, items []ConnectedItem) error
	SavePlanners(ctx context.Context, planners []Planner) error
}

// JSONArray склеивает сохранённые JSON-фрагменты в массив.
//
// Так ответ совпадает с ответом Python-mock байт-в-байт: FastAPI отдаёт
// список компактным JSON без ASCII-экранирования.
func JSONArray(payloads [][]byte) []byte {
	out := make([]byte, 0, 2+len(payloads)*64)
	out = append(out, '[')
	for i, p := range payloads {
		if i > 0 {
			out = append(out, ',')
		}
		out = append(out, p...)
	}
	out = append(out, ']')
	return out
}

// MarshalCompact сериализует значение как компактный JSON без экранирования.
func MarshalCompact(v any) ([]byte, error) {
	buf := &jsonBuffer{}
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	b := buf.Bytes()
	if n := len(b); n > 0 && b[n-1] == '\n' {
		b = b[:n-1]
	}
	return b, nil
}

type jsonBuffer struct{ b []byte }

func (j *jsonBuffer) Write(p []byte) (int, error) {
	j.b = append(j.b, p...)
	return len(p), nil
}

func (j *jsonBuffer) Bytes() []byte { return j.b }
