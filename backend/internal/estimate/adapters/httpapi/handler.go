// Package httpapi — HTTP-адаптеры контекста Estimate.
package httpapi

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"net/http"

	"github.com/zetta/zettacalc/backend/internal/estimate/domain"
	"github.com/zetta/zettacalc/backend/internal/platform/auth"
	"github.com/zetta/zettacalc/backend/internal/platform/httpx"
	"github.com/zetta/zettacalc/backend/internal/platform/pybody"
	"github.com/zetta/zettacalc/backend/internal/platform/pyjson"
	"github.com/zetta/zettacalc/backend/internal/platform/pyvalidate"
)

// staticEstimateID — идентификатор демонстрационной сметы mock API.
const staticEstimateID = "est_001"

// Handler отдаёт сметы и ЦПР.
type Handler struct {
	svc *domain.Service
}

// NewHandler создаёт обработчик смет.
func NewHandler(svc *domain.Service) *Handler { return &Handler{svc: svc} }

// ByID обслуживает GET /api/estimate/{id}.
func (h *Handler) ByID(w http.ResponseWriter, r *http.Request, params map[string]string) {
	payload, err := h.svc.JSONByID(r.Context(), params["id"])
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			httpx.WriteError(w, http.StatusNotFound, "Not found")
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "Ошибка чтения сметы")
		return
	}
	httpx.WriteRaw(w, http.StatusOK, payload)
}

// Calculate обслуживает POST /api/estimate/calculate.
func (h *Handler) Calculate(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	h.ByID(w, r, map[string]string{"id": staticEstimateID})
}

// CPR обслуживает GET /api/cpr/{id}: ЦПР собирается из демонстрационной сметы.
func (h *Handler) CPR(w http.ResponseWriter, r *http.Request, params map[string]string) {
	estimatePayload, err := h.svc.JSONByID(r.Context(), staticEstimateID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Ошибка чтения сметы")
		return
	}
	body := cprPayload(params["id"], pyjson.NewString("proj_001"), estimatePayload, []cprDoc{
		{"d1", "План", "drawing", "ready"},
		{"d2", "Ведомость", "specification", "ready"},
	})
	httpx.WriteRaw(w, http.StatusOK, body)
}

// GenerateCPR обслуживает POST /api/cpr/generate (write, требует токен).
//
// Порядок проверок как в FastAPI: тело запроса разбирается до зависимостей,
// поэтому синтаксическая ошибка JSON отвечает 422 и без токена. Идентификатор
// ЦПР считается как md5(str(payload)) — от repr разобранного тела.
func (h *Handler) GenerateCPR(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	body := pybody.Read(r.Body, r.Header.Get("Content-Type"))
	if body.DecodeError {
		httpx.WriteError(w, http.StatusBadRequest, "There was an error parsing the body")
		return
	}
	if body.Err != nil {
		errs := &pyvalidate.Errors{}
		errs.Add(*body.Err)
		httpx.WriteValidation(w, errs)
		return
	}
	if _, err := auth.RequireToken(r); err != nil {
		httpx.WriteUnauthorized(w, err)
		return
	}
	if body.Empty {
		errs := &pyvalidate.Errors{}
		errs.Add(*pybody.MissingBody())
		httpx.WriteValidation(w, errs)
		return
	}
	if !pybody.IsObject(body.Value) {
		errs := &pyvalidate.Errors{}
		errs.Add(*pybody.NotADict(body.Value, false))
		httpx.WriteValidation(w, errs)
		return
	}

	sum := md5.Sum([]byte(pyjson.Repr(body.Value)))
	cprID := "cpr_" + hex.EncodeToString(sum[:])[:8]

	// project_id переносится в ответ как есть; non-finite литерал в нём ломает
	// сериализацию ответа — mock API отвечает 500.
	projectID := pyjson.NewString("proj_001")
	if field := body.Value.Field("project_id"); field != nil {
		if pyjson.HasNonFinite(field) {
			httpx.WriteInternalError(w)
			return
		}
		projectID = field
	}

	estimatePayload, err := h.svc.JSONByID(r.Context(), staticEstimateID)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Ошибка чтения сметы")
		return
	}
	out := cprPayload(cprID, projectID, estimatePayload, []cprDoc{
		{"d1", "План", "drawing", "ready"},
		{"d2", "Ведомость", "specification", "pending_review"},
	})
	httpx.WriteRaw(w, http.StatusOK, out)
}

type cprDoc struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Type   string `json:"type"`
	Status string `json:"status"`
}

func cprPayload(id string, projectID *pyjson.Value, estimate []byte, docs []cprDoc) []byte {
	body := make([]byte, 0, len(estimate)+256)
	body = append(body, `{"id":`...)
	body = appendJSONString(body, id)
	body = append(body, `,"version":1,"project_id":`...)
	body = append(body, pyjson.Encode(projectID)...)
	body = append(body, `,"estimate":`...)
	body = append(body, estimate...)
	body = append(body, `,"documents":[`...)
	for i, d := range docs {
		if i > 0 {
			body = append(body, ',')
		}
		body = append(body, `{"id":`...)
		body = appendJSONString(body, d.ID)
		body = append(body, `,"title":`...)
		body = appendJSONString(body, d.Title)
		body = append(body, `,"type":`...)
		body = appendJSONString(body, d.Type)
		body = append(body, `,"status":`...)
		body = appendJSONString(body, d.Status)
		body = append(body, '}')
	}
	body = append(body, `],"status":"ready"}`...)
	return body
}
