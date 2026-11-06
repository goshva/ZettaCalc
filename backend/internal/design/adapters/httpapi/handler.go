// Package httpapi — HTTP-адаптеры контекста Design.
package httpapi

import (
	"errors"
	"net/http"

	"github.com/zetta/zettacalc/backend/internal/design/app"
	"github.com/zetta/zettacalc/backend/internal/design/domain"
	"github.com/zetta/zettacalc/backend/internal/platform/httpx"
)

// Handler отдаёт варианты дизайна и производные сметы, ЦПР и ставки.
type Handler struct {
	svc    *domain.Service
	derive *app.Derive
}

// NewHandler создаёт обработчик вариантов.
func NewHandler(svc *domain.Service, derive *app.Derive) *Handler {
	return &Handler{svc: svc, derive: derive}
}

// Variants обслуживает GET /api/design/variants.
func (h *Handler) Variants(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	payload, err := h.svc.VariantsJSON(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Ошибка чтения вариантов")
		return
	}
	httpx.WriteRaw(w, http.StatusOK, payload)
}

// Estimate обслуживает GET /api/design/variants/{id}/estimate.
func (h *Handler) Estimate(w http.ResponseWriter, r *http.Request, params map[string]string) {
	id := params["id"]
	payload, err := h.derive.EstimateJSON(r.Context(), id)
	if err != nil {
		h.writeVariantError(w, err, id)
		return
	}
	httpx.WriteRaw(w, http.StatusOK, payload)
}

// CPR обслуживает GET /api/design/variants/{id}/cpr.
func (h *Handler) CPR(w http.ResponseWriter, r *http.Request, params map[string]string) {
	id := params["id"]
	payload, err := h.derive.CPRJSON(r.Context(), id)
	if err != nil {
		h.writeVariantError(w, err, id)
		return
	}
	httpx.WriteRaw(w, http.StatusOK, payload)
}

// Bids обслуживает GET /api/design/variants/{id}/bids.
func (h *Handler) Bids(w http.ResponseWriter, r *http.Request, params map[string]string) {
	id := params["id"]
	payload, err := h.derive.BidsJSON(r.Context(), id)
	if err != nil {
		h.writeVariantError(w, err, id)
		return
	}
	httpx.WriteRaw(w, http.StatusOK, payload)
}

// writeVariantError повторяет текст ошибки mock API.
func (h *Handler) writeVariantError(w http.ResponseWriter, err error, id string) {
	if errors.Is(err, domain.ErrNotFound) {
		httpx.WriteError(w, http.StatusNotFound, "Variant "+id+" not found")
		return
	}
	httpx.WriteError(w, http.StatusInternalServerError, "Ошибка чтения варианта")
}
