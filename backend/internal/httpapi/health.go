package httpapi

import (
	"net/http"

	"github.com/zetta/zettacalc/backend/internal/content/domain"
	"github.com/zetta/zettacalc/backend/internal/platform/httpx"
	"github.com/zetta/zettacalc/backend/internal/platform/jsonx"
)

// HealthHandler отдаёт состояние сервиса.
type HealthHandler struct {
	deps Deps
}

// NewHealthHandler создаёт обработчик /health.
func NewHealthHandler(deps Deps) *HealthHandler { return &HealthHandler{deps: deps} }

// Handle обслуживает GET /health и GET /api/health.
//
// Набор и порядок полей повторяют mock API: тела ответов сравниваются
// байт-в-байт, поэтому «лишние» счётчики Go-сервиса тут недопустимы.
func (h *HealthHandler) Handle(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	ctx := r.Context()
	stats, err := h.deps.Media.Stats(ctx)
	if err != nil {
		httpx.WriteError(w, http.StatusServiceUnavailable, "Хранилище изображений недоступно")
		return
	}
	interiors, _ := h.deps.Content.Interiors(ctx, domain.InteriorFilter{})
	products, _ := h.deps.Content.Products(ctx, "")
	variants, _ := h.deps.Design.Variants(ctx)
	sections, _ := h.deps.Assets.MainSectionCount(ctx)
	appointments, _ := h.deps.Booking.Count(ctx)

	obj := jsonx.NewObject()
	obj.Set("status", jsonx.NewString("ok"))
	obj.Set("mock", jsonx.NewBool(false))
	obj.Set("images", jsonx.NewInt(int64(stats.Keys)))
	obj.Set("interiors", jsonx.NewInt(int64(len(interiors))))
	obj.Set("products", jsonx.NewInt(int64(len(products))))
	obj.Set("design_variants", jsonx.NewInt(int64(len(variants))))
	obj.Set("main_sections", jsonx.NewInt(int64(sections)))
	obj.Set("apps", jsonx.NewInt(int64(h.deps.Assets.AppCount())))
	obj.Set("appointments", jsonx.NewInt(int64(appointments)))

	httpx.WriteJSON(w, http.StatusOK, obj)
}
