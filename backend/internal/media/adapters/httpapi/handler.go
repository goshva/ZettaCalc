// Package httpapi — HTTP-адаптеры контекста Media.
package httpapi

import (
	"errors"
	"net/http"
	"strings"

	"github.com/zetta/zettacalc/backend/internal/media/domain"
	"github.com/zetta/zettacalc/backend/internal/platform/httpx"
)

// Handler отдаёт изображения по ключу и по имени файла.
type Handler struct {
	svc *domain.Service
}

// NewHandler создаёт обработчик изображений.
func NewHandler(svc *domain.Service) *Handler { return &Handler{svc: svc} }

// ByKey обслуживает GET /api/images/{key}.
func (h *Handler) ByKey(w http.ResponseWriter, r *http.Request, params map[string]string) {
	key := params["key"]
	asset, err := h.svc.ByKey(r.Context(), key)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			httpx.WriteError(w, http.StatusNotFound, "Изображение '"+key+"' не найдено")
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "Ошибка хранилища изображений")
		return
	}
	httpx.WriteBytes(w, http.StatusOK, asset.ContentType, asset.Bytes, "public, max-age=86400")
}

// ByFilename обслуживает GET /api/images/file/{filename}.
func (h *Handler) ByFilename(w http.ResponseWriter, r *http.Request, params map[string]string) {
	name := params["filename"]
	if name == "" || strings.ContainsAny(name, `/\`) || strings.Contains(name, "..") {
		httpx.WriteError(w, http.StatusNotFound, "Файл "+name+" отсутствует в хранилище")
		return
	}
	asset, err := h.svc.ByFilename(r.Context(), name)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			httpx.WriteError(w, http.StatusNotFound, "Файл "+name+" отсутствует в хранилище")
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "Ошибка хранилища изображений")
		return
	}
	httpx.WriteBytes(w, http.StatusOK, asset.ContentType, asset.Bytes, "public, max-age=86400")
}
