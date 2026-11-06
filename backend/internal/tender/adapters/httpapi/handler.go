// Package httpapi — HTTP-адаптеры контекста Tender.
package httpapi

import (
	"net/http"

	"github.com/zetta/zettacalc/backend/internal/platform/httpx"
	"github.com/zetta/zettacalc/backend/internal/tender/domain"
)

// Handler отдаёт ставки подрядчиков.
type Handler struct {
	svc *domain.Service
}

// NewHandler создаёт обработчик ставок.
func NewHandler(svc *domain.Service) *Handler { return &Handler{svc: svc} }

// Static обслуживает GET /api/bids — демонстрационная подборка.
func (h *Handler) Static(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	bids, err := h.svc.Static(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Ошибка чтения ставок")
		return
	}
	payloads := make([][]byte, 0, len(bids))
	for _, b := range bids {
		payloads = append(payloads, b.Payload)
	}
	httpx.WriteRaw(w, http.StatusOK, joinArray(payloads))
}

func joinArray(items [][]byte) []byte {
	out := make([]byte, 0, 2)
	out = append(out, '[')
	for i, item := range items {
		if i > 0 {
			out = append(out, ',')
		}
		out = append(out, item...)
	}
	return append(out, ']')
}
