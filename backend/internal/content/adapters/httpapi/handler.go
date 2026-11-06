// Package httpapi — HTTP-адаптеры контекста Content.
package httpapi

import (
	"errors"
	"net/http"

	"github.com/zetta/zettacalc/backend/internal/content/domain"
	"github.com/zetta/zettacalc/backend/internal/platform/httpx"
)

// Handler отдаёт контент прототипов: интерьеры, каталог, центр, hotspot.
type Handler struct {
	repo domain.Repository
}

// NewHandler создаёт обработчик контента.
func NewHandler(repo domain.Repository) *Handler { return &Handler{repo: repo} }

// Interiors обслуживает GET /api/interiors.
func (h *Handler) Interiors(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	q := r.URL.Query()
	items, err := h.repo.Interiors(r.Context(), domain.InteriorFilter{
		Category: q.Get("category"),
		Room:     q.Get("room"),
		Style:    q.Get("style"),
	})
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Ошибка чтения интерьеров")
		return
	}
	payloads := make([][]byte, 0, len(items))
	for _, it := range items {
		payloads = append(payloads, it.Payload)
	}
	httpx.WriteRaw(w, http.StatusOK, domain.JSONArray(payloads))
}

// InteriorsCategories обслуживает GET /api/interiors/categories.
func (h *Handler) InteriorsCategories(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	items, err := h.repo.InteriorCategories(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Ошибка чтения категорий")
		return
	}
	httpx.WriteRaw(w, http.StatusOK, stringListPayload(items))
}

// InteriorsDescriptions обслуживает GET /api/interiors/descriptions.
func (h *Handler) InteriorsDescriptions(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	items, err := h.repo.InteriorDescriptions(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Ошибка чтения описаний")
		return
	}
	httpx.WriteRaw(w, http.StatusOK, stringListPayload(items))
}

func stringListPayload(items []domain.Named) []byte {
	out := make([]byte, 0, 2+len(items)*24)
	out = append(out, '[')
	for i, it := range items {
		if i > 0 {
			out = append(out, ',')
		}
		encoded, err := domain.MarshalCompact(it.Name)
		if err != nil {
			continue
		}
		out = append(out, encoded...)
	}
	return append(out, ']')
}

// InteriorDetail обслуживает GET /api/interiors/{id}.
func (h *Handler) InteriorDetail(w http.ResponseWriter, r *http.Request, params map[string]string) {
	item, err := h.repo.InteriorByID(r.Context(), params["id"])
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			httpx.WriteError(w, http.StatusNotFound, "Interior "+params["id"]+" not found")
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "Ошибка чтения интерьера")
		return
	}
	httpx.WriteRaw(w, http.StatusOK, item.Payload)
}

// Products обслуживает GET /api/catalog/products.
func (h *Handler) Products(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	items, err := h.repo.Products(r.Context(), r.URL.Query().Get("group"))
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Ошибка чтения товаров")
		return
	}
	payloads := make([][]byte, 0, len(items))
	for _, p := range items {
		payloads = append(payloads, p.Payload)
	}
	httpx.WriteRaw(w, http.StatusOK, domain.JSONArray(payloads))
}

// CatalogRooms обслуживает GET /api/catalog/rooms.
func (h *Handler) CatalogRooms(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	items, err := h.repo.CatalogRooms(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Ошибка чтения комнат")
		return
	}
	payloads := make([][]byte, 0, len(items))
	for _, rm := range items {
		payloads = append(payloads, rm.Payload)
	}
	httpx.WriteRaw(w, http.StatusOK, domain.JSONArray(payloads))
}

// CatalogRoomDetail обслуживает GET /api/catalog/rooms/{id}.
func (h *Handler) CatalogRoomDetail(w http.ResponseWriter, r *http.Request, params map[string]string) {
	room, err := h.repo.CatalogRoomByID(r.Context(), params["id"])
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			httpx.WriteError(w, http.StatusNotFound, "Room "+params["id"]+" not found")
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "Ошибка чтения комнаты")
		return
	}
	httpx.WriteRaw(w, http.StatusOK, room.Payload)
}

// CenterInteriors обслуживает GET /api/center/interiors.
func (h *Handler) CenterInteriors(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	items, err := h.repo.CenterInteriors(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Ошибка чтения интерьеров центра")
		return
	}
	payloads := make([][]byte, 0, len(items))
	for _, c := range items {
		payloads = append(payloads, c.Payload)
	}
	httpx.WriteRaw(w, http.StatusOK, domain.JSONArray(payloads))
}

// CenterPhotos обслуживает GET /api/center/photos.
func (h *Handler) CenterPhotos(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	items, err := h.repo.CenterPhotos(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Ошибка чтения фото центра")
		return
	}
	payloads := make([][]byte, 0, len(items))
	for _, p := range items {
		payloads = append(payloads, p.Payload)
	}
	httpx.WriteRaw(w, http.StatusOK, domain.JSONArray(payloads))
}

// HotspotData обслуживает GET /api/hotspot/data.
func (h *Handler) HotspotData(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	items, err := h.repo.HotspotItems(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Ошибка чтения hotspot")
		return
	}
	payloads := make([][]byte, 0, len(items))
	for _, it := range items {
		payloads = append(payloads, it.Payload)
	}
	httpx.WriteRaw(w, http.StatusOK, domain.JSONArray(payloads))
}

// HotspotAnalogues обслуживает GET /api/hotspot/analogues.
func (h *Handler) HotspotAnalogues(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	items, err := h.repo.AnalogueItems(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Ошибка чтения аналогов")
		return
	}
	payloads := make([][]byte, 0, len(items))
	for _, a := range items {
		payloads = append(payloads, a.Payload)
	}
	httpx.WriteRaw(w, http.StatusOK, domain.JSONArray(payloads))
}

// References обслуживает GET /api/references.
func (h *Handler) References(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	items, err := h.repo.References(r.Context(), r.URL.Query().Get("category"))
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Ошибка чтения референсов")
		return
	}
	payloads := make([][]byte, 0, len(items))
	for _, ref := range items {
		payloads = append(payloads, ref.Payload)
	}
	httpx.WriteRaw(w, http.StatusOK, domain.JSONArray(payloads))
}

// ReferenceDetail обслуживает GET /api/references/{id}.
func (h *Handler) ReferenceDetail(w http.ResponseWriter, r *http.Request, params map[string]string) {
	ref, err := h.repo.ReferenceByID(r.Context(), params["id"])
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			httpx.WriteError(w, http.StatusNotFound, "Not found")
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "Ошибка чтения референса")
		return
	}
	httpx.WriteRaw(w, http.StatusOK, ref.Payload)
}

// Planners обслуживает GET /api/planners.
func (h *Handler) Planners(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	items, err := h.repo.Planners(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Ошибка чтения планировок")
		return
	}
	payloads := make([][]byte, 0, len(items))
	for _, p := range items {
		payloads = append(payloads, p.Payload)
	}
	httpx.WriteRaw(w, http.StatusOK, domain.JSONArray(payloads))
}
