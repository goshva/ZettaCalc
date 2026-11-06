// Package httpapi — HTTP-адаптеры контекста Techcard.
package httpapi

import (
	"net/http"

	"github.com/zetta/zettacalc/backend/internal/platform/httpx"
	"github.com/zetta/zettacalc/backend/internal/techcard/domain"
)

// Handler отдаёт техкарты.
type Handler struct {
	svc *domain.Service
}

// NewHandler создаёт обработчик техкарт.
func NewHandler(svc *domain.Service) *Handler { return &Handler{svc: svc} }

// List обслуживает GET /api/techcards.
func (h *Handler) List(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	cards, err := h.svc.List(r.Context(), r.URL.Query().Get("surface"))
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Ошибка чтения техкарт")
		return
	}
	payloads := make([][]byte, 0, len(cards))
	for _, c := range cards {
		payloads = append(payloads, c.Payload)
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
