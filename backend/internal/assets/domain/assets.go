// Package domain контекста Assets: CSS, разметка приложений и фрагменты прототипов.
//
// Это generic-контекст: он отдаёт текстовые ресурсы по пути и ничего не знает
// о бизнес-смысле содержимого.
package domain

import (
	"context"
	"errors"
)

// ErrNotFound — ресурс отсутствует.
var ErrNotFound = errors.New("assets: ресурс не найден")

// Kind — вид ресурса.
type Kind string

const (
	// KindApp — файл приложения прототипа А (main/interiors/catalog/center/repair/about).
	KindApp Kind = "app"
	// KindMainSection — секция главной страницы прототипа А.
	KindMainSection Kind = "main_section"
	// KindPage — фрагмент разметки прототипа (a|b).
	KindPage Kind = "page"
	// KindCSS — таблица стилей прототипа (a_all.css, a_shell.css, …).
	KindCSS Kind = "css"
)

// File — текстовый ресурс.
type File struct {
	Path        string
	Kind        Kind
	AppName     string
	Proto       string
	Name        string
	ContentType string
	Body        string
	Order       int
}

// Repository — порт чтения ресурсов.
type Repository interface {
	ByPath(ctx context.Context, path string) (File, error)
	AppFile(ctx context.Context, app, ext string) (File, error)
	MainSection(ctx context.Context, name string) (File, error)
	PageFragment(ctx context.Context, proto, name string) (File, error)
	ProtoCSS(ctx context.Context, name string) (File, error)
	List(ctx context.Context, kind Kind, proto string) ([]File, error)
	Save(ctx context.Context, files []File) error
	SaveMeta(ctx context.Context, values map[string]string) error
	Meta(ctx context.Context, key string) (string, error)
}

// AppFileNames — какие файлы есть у приложения (совпадает с mock API).
var AppFileNames = map[string]struct{ HTML, CSS string }{
	"main":      {"", "main.css"},
	"interiors": {"interiors.html", "interiors.css"},
	"catalog":   {"catalog.html", "catalog.css"},
	"center":    {"center.html", "center.css"},
	"repair":    {"repair.html", "repair.css"},
	"about":     {"about.html", "about.css"},
}

// Service — операции контекста Assets.
type Service struct {
	repo Repository
}

// NewService собирает сервис.
func NewService(repo Repository) *Service { return &Service{repo: repo} }

// AppCSS отдаёт CSS приложения.
func (s *Service) AppCSS(ctx context.Context, app string) (File, error) {
	entry, ok := AppFileNames[app]
	if !ok {
		return File{}, ErrNotFound
	}
	return s.repo.AppFile(ctx, app, entry.CSS)
}

// AppHTML отдаёт разметку приложения.
func (s *Service) AppHTML(ctx context.Context, app string) (File, error) {
	entry, ok := AppFileNames[app]
	if !ok || entry.HTML == "" {
		return File{}, ErrNotFound
	}
	return s.repo.AppFile(ctx, app, entry.HTML)
}
// MainSection отдаёт секцию главной страницы.
func (s *Service) MainSection(ctx context.Context, name string) (File, error) {
	return s.repo.MainSection(ctx, name)
}

// PageFragment отдаёт фрагмент разметки прототипа.
func (s *Service) PageFragment(ctx context.Context, proto, name string) (File, error) {
	if proto != "a" && proto != "b" {
		return File{}, ErrNotFound
	}
	return s.repo.PageFragment(ctx, proto, name)
}

// ProtoCSS отдаёт таблицу стилей прототипа.
func (s *Service) ProtoCSS(ctx context.Context, name string) (File, error) {
	return s.repo.ProtoCSS(ctx, name)
}

// List отдаёт список ресурсов.
func (s *Service) List(ctx context.Context, kind Kind, proto string) ([]File, error) {
	return s.repo.List(ctx, kind, proto)
}

// PageDirs возвращает каталоги фрагментов разметки прототипов.
//
// Пути хранятся в seed_meta: они часть ответа /api/page и зависят от того,
// откуда запущен загрузчик, поэтому их нельзя «вычислить» в обработчике.
func (s *Service) PageDirs(ctx context.Context) (map[string]string, error) {
	out := map[string]string{"a": "a_page", "b": "b_page"}
	for _, proto := range []string{"a", "b"} {
		value, err := s.repo.Meta(ctx, "page_dir_"+proto)
		if err != nil {
			return nil, err
		}
		if value != "" {
			out[proto] = value
		}
	}
	return out, nil
}

// MainSectionCount возвращает число секций главной страницы (для /api/health).
func (s *Service) MainSectionCount(ctx context.Context) (int, error) {
	files, err := s.repo.List(ctx, KindMainSection, "")
	if err != nil {
		return 0, err
	}
	return len(files), nil
}

// AppCount возвращает число приложений прототипа А (для /api/health).
func (s *Service) AppCount() int { return len(AppFileNames) }
