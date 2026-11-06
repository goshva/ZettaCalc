// Package httpapi — HTTP-адаптеры контекста Media.
package httpapi

import (
	"net/http"

	"github.com/zetta/zettacalc/backend/internal/platform/httpx"
	"github.com/zetta/zettacalc/backend/internal/platform/jsonx"
)

// Index обслуживает GET /api/images: список ключей и место хранения.
//
// Формат и порядок ключей совпадают с mock API: тела ответов сравниваются
// байт-в-байт, поэтому ключи идут в порядке объявления IMAGE_KEYS.
func (h *Handler) Index(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	list, err := h.svc.Keys(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Ошибка чтения ключей изображений")
		return
	}

	keys := jsonx.NewObject()
	for _, k := range list {
		item := jsonx.NewObject()
		item.Set("file", jsonx.NewString(k.Filename))
		item.Set("source", jsonx.NewString(k.Source))
		keys.Set(k.Key, item)
	}

	obj := jsonx.NewObject()
	obj.Set("served_as", jsonx.NewString("/api/images/{key}"))
	obj.Set("storage", jsonx.NewString("БД / PIM (контент)"))
	obj.Set("keys", keys)
	httpx.WriteJSON(w, http.StatusOK, obj)
}
