// Package httpapi — HTTP-адаптеры контекста Assets.
package httpapi

import (
	"errors"
	"net/http"
	"path"
	"strings"

	"github.com/zetta/zettacalc/backend/internal/assets/domain"
	"github.com/zetta/zettacalc/backend/internal/platform/httpx"
	"github.com/zetta/zettacalc/backend/internal/platform/jsonx"
)

// Handler отдаёт CSS, разметку и фрагменты прототипов.
type Handler struct {
	svc *domain.Service
}

// NewHandler создаёт обработчик ресурсов.
func NewHandler(svc *domain.Service) *Handler { return &Handler{svc: svc} }

// appOrder — порядок приложений в ответе /api/apps (как в mock API).
var appOrder = []string{"main", "interiors", "catalog", "center", "repair", "about"}

// AppsIndex обслуживает GET /api/apps — список приложений прототипа А.
func (h *Handler) AppsIndex(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	out := jsonx.NewObject()
	for _, name := range appOrder {
		entry, ok := domain.AppFileNames[name]
		if !ok {
			continue
		}
		files := jsonx.NewObject()
		files.Set("html", nullableJSONString(entry.HTML))
		files.Set("css", jsonx.NewString(entry.CSS))

		item := jsonx.NewObject()
		if entry.HTML == "" {
			item.Set("html", jsonx.NewNull())
		} else {
			item.Set("html", jsonx.NewString("/api/apps/"+name+"/html"))
		}
		item.Set("css", jsonx.NewString("/api/apps/"+name+"/css"))
		item.Set("files", files)
		out.Set(name, item)
	}
	httpx.WriteJSON(w, http.StatusOK, out)
}

// AppCSS обслуживает GET /api/apps/{app}/css.
func (h *Handler) AppCSS(w http.ResponseWriter, r *http.Request, params map[string]string) {
	app := params["app"]
	file, err := h.svc.AppCSS(r.Context(), app)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			if _, ok := domain.AppFileNames[app]; !ok {
				httpx.WriteError(w, http.StatusNotFound, "Приложение '"+app+"' не известно")
				return
			}
			httpx.WriteError(w, http.StatusNotFound, "CSS "+domain.AppFileNames[app].CSS+" не найден")
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "Ошибка чтения CSS")
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=3600")
	httpx.WriteText(w, http.StatusOK, file.ContentType, file.Body)
}

// AppHTML обслуживает GET /api/apps/{app}/html.
func (h *Handler) AppHTML(w http.ResponseWriter, r *http.Request, params map[string]string) {
	app := params["app"]
	entry, ok := domain.AppFileNames[app]
	if !ok {
		httpx.WriteError(w, http.StatusNotFound, "Приложение '"+app+"' не известно")
		return
	}
	if entry.HTML == "" {
		httpx.WriteError(w, http.StatusNotFound, "У приложения '"+app+"' есть только CSS")
		return
	}
	file, err := h.svc.AppHTML(r.Context(), app)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			httpx.WriteError(w, http.StatusNotFound, "Разметка "+entry.HTML+" не найдена")
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "Ошибка чтения разметки")
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=3600")
	httpx.WriteText(w, http.StatusOK, file.ContentType, file.Body)
}

// MainSection отдаёт секцию главной страницы (start, how, ideas, connected).
//
// mock API объявляет эти обработчики без response_class, поэтому FastAPI
// сериализует HTML-строку как JSON: content-type application/json и кавычки.
func (h *Handler) MainSection(name string) httpx.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
		file, err := h.svc.MainSection(r.Context(), name)
		if err != nil {
			httpx.WriteJSON(w, http.StatusOK, "")
			return
		}
		httpx.WriteJSON(w, http.StatusOK, file.Body)
	}
}

// PageIndex обслуживает GET /api/page.
func (h *Handler) PageIndex(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	dirs, err := h.svc.PageDirs(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Ошибка чтения фрагментов")
		return
	}
	out := jsonx.NewObject()
	for _, proto := range []string{"a", "b"} {
		files, err := h.svc.List(r.Context(), domain.KindPage, proto)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "Ошибка чтения фрагментов")
			return
		}
		names := make([]*jsonx.Value, 0, len(files))
		for _, f := range files {
			names = append(names, jsonx.NewString(path.Base(f.Path)))
		}
		item := jsonx.NewObject()
		item.Set("dir", jsonx.NewString(dirs[proto]))
		item.Set("files", jsonx.NewArray(names...))
		out.Set(proto, item)
	}
	out.Set("served_as", jsonx.NewString("/api/page/{proto}/{file}"))
	httpx.WriteJSON(w, http.StatusOK, out)
}

// PageFragment обслуживает GET /api/page/{proto}/{file}.
func (h *Handler) PageFragment(w http.ResponseWriter, r *http.Request, params map[string]string) {
	proto := params["proto"]
	name := params["file"]
	// порядок проверок как в mock API: сначала прототип, затем имя файла
	if proto != "a" && proto != "b" {
		httpx.WriteError(w, http.StatusNotFound, "Прототип '"+proto+"' не известен (a|b)")
		return
	}
	if strings.ContainsAny(name, `/\`) || strings.Contains(name, "..") {
		httpx.WriteError(w, http.StatusBadRequest, "Некорректное имя файла")
		return
	}
	file, err := h.svc.PageFragment(r.Context(), proto, name)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			httpx.WriteError(w, http.StatusNotFound, "Фрагмент "+proto+"/"+name+" не найден")
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "Ошибка чтения фрагмента")
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=3600")
	httpx.WriteText(w, http.StatusOK, file.ContentType, file.Body)
}

// ProtoCSS обслуживает GET /api/css/{proto}.
func (h *Handler) ProtoCSS(w http.ResponseWriter, r *http.Request, params map[string]string) {
	proto := params["proto"]
	name := map[string]string{"a": "a_all.css", "b": "b_all.css"}[proto]
	if name == "" {
		httpx.WriteError(w, http.StatusNotFound, "proto должен быть a или b")
		return
	}
	h.writeCSS(w, r, name, "CSS "+name+" не найден")
}

// ProtoShellCSS обслуживает GET /api/css/{proto}/shell.
func (h *Handler) ProtoShellCSS(w http.ResponseWriter, r *http.Request, params map[string]string) {
	proto := params["proto"]
	name := map[string]string{"a": "a_shell.css", "b": "b_shell.css"}[proto]
	if name == "" {
		httpx.WriteError(w, http.StatusNotFound, "proto должен быть a или b")
		return
	}
	h.writeCSS(w, r, name, "CSS оболочки "+name+" не найден")
}

func (h *Handler) writeCSS(w http.ResponseWriter, r *http.Request, name, notFound string) {
	file, err := h.svc.ProtoCSS(r.Context(), name)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			httpx.WriteError(w, http.StatusNotFound, notFound)
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "Ошибка чтения CSS")
		return
	}
	httpx.WriteText(w, http.StatusOK, file.ContentType, file.Body)
}

// nullableJSONString отдаёт null вместо пустой строки — как pydantic-модель.
func nullableJSONString(s string) *jsonx.Value {
	if s == "" {
		return jsonx.NewNull()
	}
	return jsonx.NewString(s)
}
