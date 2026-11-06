// Package domain контекста Media: файлы изображений и ключи PIM.
//
// Контекст отвечает только за хранение и выдачу картинок. Он не знает, к какой
// сущности относится файл: связь живёт в контексте Content и в таблице
// media_entity_ref. Такое разделение позволяет переносить картинки сущностей
// (storage = entity) и не тащить в БД общую статику сайта (storage = client).
package domain

import (
	"context"
	"errors"
)

// ErrNotFound — файл или ключ отсутствует в хранилище.
var ErrNotFound = errors.New("media: изображение не найдено")

// Storage — где живёт файл.
type Storage string

const (
	// StorageEntity — картинка сущности, хранится в БД.
	StorageEntity Storage = "entity"
	// StorageStatic — общая статика сайта, в БД не дублируется.
	StorageStatic Storage = "static"
	// StorageClient — файл раздаётся фронтендом (Vite public).
	StorageClient Storage = "client"
)

// Asset — файл изображения.
type Asset struct {
	Filename    string
	ContentType string
	ByteSize    int64
	SHA256      string
	Storage     Storage
	GroupPrefix string
	Bytes       []byte
}

// Key — именованный ключ картинки из PIM (idea-kitchen, hero, …).
type Key struct {
	Key      string
	Filename string
	Source   string // db | static
	Storage  string // db | client
	Order    int
}

// EntityRef — привязка картинки к сущности.
type EntityRef struct {
	Filename   string
	EntityType string
	EntityID   string
	Role       string
	Position   int
}

// Repository — порт хранения файлов.
type Repository interface {
	// AssetByFilename возвращает файл по имени (a_main_img_0001.jpg).
	AssetByFilename(ctx context.Context, filename string) (Asset, error)
	// AssetByKey возвращает файл по ключу PIM (idea-kitchen).
	AssetByKey(ctx context.Context, key string) (Asset, error)
	// Keys возвращает все ключи PIM.
	Keys(ctx context.Context) ([]Key, error)
	// SaveAssets сохраняет файлы пачкой (используется загрузчиком).
	SaveAssets(ctx context.Context, assets []Asset) error
	// SaveKeys сохраняет ключи PIM.
	SaveKeys(ctx context.Context, keys []Key) error
	// SaveRefs сохраняет привязки картинок к сущностям.
	SaveRefs(ctx context.Context, refs []EntityRef) error
	// Stats возвращает счётчики для /api/health.
	Stats(ctx context.Context) (Stats, error)
}

// Stats — агрегаты хранилища изображений.
type Stats struct {
	Assets   int
	Keys     int
	Entities int
	Bytes    int64
}

// Service — операции контекста Media.
type Service struct {
	repo Repository
}

// NewService собирает сервис поверх репозитория.
func NewService(repo Repository) *Service { return &Service{repo: repo} }

// ByKey отдаёт картинку по ключу PIM.
func (s *Service) ByKey(ctx context.Context, key string) (Asset, error) {
	return s.repo.AssetByKey(ctx, key)
}

// ByFilename отдаёт картинку по имени файла.
func (s *Service) ByFilename(ctx context.Context, filename string) (Asset, error) {
	return s.repo.AssetByFilename(ctx, filename)
}

// Stats отдаёт агрегаты хранилища.
func (s *Service) Stats(ctx context.Context) (Stats, error) { return s.repo.Stats(ctx) }

// Keys отдаёт все ключи PIM.
func (s *Service) Keys(ctx context.Context) ([]Key, error) { return s.repo.Keys(ctx) }
