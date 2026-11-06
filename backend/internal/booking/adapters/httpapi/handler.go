// Package httpapi — HTTP-адаптеры контекста Booking.
package httpapi

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/zetta/zettacalc/backend/internal/booking/domain"
	"github.com/zetta/zettacalc/backend/internal/platform/httpx"
	"github.com/zetta/zettacalc/backend/internal/platform/idempotency"
	"github.com/zetta/zettacalc/backend/internal/platform/jsonx"
	"github.com/zetta/zettacalc/backend/internal/platform/pybody"
	"github.com/zetta/zettacalc/backend/internal/platform/pyjson"
	"github.com/zetta/zettacalc/backend/internal/platform/pyvalidate"
)

// scopeIdempotency — область ключей идемпотентности этого эндпоинта.
const scopeIdempotency = "POST /api/center/appointment"

// Handler обслуживает форму записи в Интерьерный центр.
type Handler struct {
	svc    *domain.Service
	keys   *idempotency.Store
	nowFn  func() time.Time
	random func(n int) string
}

// NewHandler создаёт обработчик записи.
func NewHandler(svc *domain.Service, keys *idempotency.Store) *Handler {
	return &Handler{svc: svc, keys: keys, nowFn: func() time.Time { return time.Now().UTC() }, random: randomHex}
}

// Options обслуживает GET /api/center/appointment/options.
func (h *Handler) Options(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	opts := domain.DefaultOptions()

	place := jsonx.NewObject()
	place.Set("address", jsonx.NewString(opts.Place.Address))
	place.Set("mall", jsonx.NewString(opts.Place.Mall))
	place.Set("phone", jsonx.NewString(opts.Place.Phone))
	place.Set("phone_href", jsonx.NewString(opts.Place.PhoneHref))
	place.Set("map", jsonx.NewString(opts.Place.Map))
	place.Set("note", jsonx.NewString(opts.Place.Note))

	obj := jsonx.NewObject()
	obj.Set("tasks", stringValues(opts.Tasks))
	obj.Set("times", stringValues(opts.Times))
	obj.Set("max_ideas", jsonx.NewInt(int64(opts.MaxIdeas)))
	obj.Set("max_files", jsonx.NewInt(int64(opts.MaxFiles)))
	obj.Set("place", place)
	obj.Set("consent_url", jsonx.NewString(opts.ConsentURL))
	obj.Set("policy_url", jsonx.NewString(opts.PolicyURL))
	obj.Set("dry_run", jsonx.NewString(opts.DryRun))
	obj.Set("create", jsonx.NewString(opts.Create))

	httpx.WriteJSON(w, http.StatusOK, obj)
}

func stringValues(items []string) *jsonx.Value {
	out := make([]*jsonx.Value, 0, len(items))
	for _, s := range items {
		out = append(out, jsonx.NewString(s))
	}
	return jsonx.NewArray(out...)
}

// DryRun обслуживает POST /api/center/appointment/dry-run: проверка без записи.
func (h *Handler) DryRun(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	errs := &pyvalidate.Errors{}
	decoded, ok := h.decodeInto(w, r, errs)
	if errs.Len() > 0 {
		httpx.WriteValidation(w, errs)
		return
	}
	if !ok {
		return
	}
	result := h.svc.DryRun(decoded.Request, h.nowFn())
	httpx.WriteJSON(w, http.StatusOK, validationBody(result))
}

// Create обслуживает POST /api/center/appointment.
//
// Порядок проверок повторяет mock API (FastAPI): разбор query-параметра
// approved и тела запроса (422) → заголовок Idempotency-Key (400) → повтор с
// тем же ключом (200) → проверка формы (422) → approval-gate (428) → запись.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	// pydantic разбирает query-параметр раньше тела и объединяет ошибки.
	errs := &pyvalidate.Errors{}
	approved, approvedOK := parseApproved(r.URL.Query().Get("approved"), r.URL.Query().Has("approved"), errs)

	req, ok := h.decodeInto(w, r, errs)
	if errs.Len() > 0 {
		httpx.WriteValidation(w, errs)
		return
	}
	if !ok {
		return
	}

	idemKey := r.Header.Get("Idempotency-Key")
	if idemKey == "" {
		httpx.WriteError(w, http.StatusBadRequest, "Требуется заголовок Idempotency-Key")
		return
	}

	// Повтор с известным ключом возвращает сохранённый ответ независимо от
	// тела, подтверждения и доменных ошибок — как в mock API, где индекс
	// проверяется до валидации формы.
	if record, err := h.keys.LookupByKey(r.Context(), scopeIdempotency, idemKey); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Ошибка проверки идемпотентности")
		return
	} else if record != nil {
		body, err := withReplayFlag(record.Response)
		if err != nil {
			httpx.WriteError(w, http.StatusInternalServerError, "Ошибка воспроизведения ответа")
			return
		}
		httpx.WriteRaw(w, record.StatusCode, body)
		return
	}

	check := h.svc.DryRun(req.Request, h.nowFn())
	if !check.OK {
		httpx.WriteError(w, http.StatusUnprocessableEntity, map[string]any{
			"errors":   nonNil(check.Errors),
			"warnings": nonNil(check.Warnings),
		})
		return
	}

	// approval-gate: без явного подтверждения запись не создаётся
	if !approvedOK || !approved {
		obj := jsonx.NewObject()
		inner := jsonx.NewObject()
		inner.Set("detail", jsonx.NewString("Запись требует подтверждения менеджером (approval-gate)"))
		inner.Set("dry_run", validationBody(check))
		obj.Set("detail", inner)
		httpx.WriteJSON(w, http.StatusPreconditionRequired, obj)
		return
	}

	created := h.nowFn()
	appointment := domain.Appointment{
		ID:         "appt_" + h.random(5),
		Status:     "pending_confirmation",
		CreatedAt:  created,
		Request:    req.Request,
		Validation: check,
	}
	if err := h.svc.Create(r.Context(), appointment); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Ошибка сохранения записи")
		return
	}

	body := appointmentBody(appointment, false)
	if err := h.keys.Save(r.Context(), scopeIdempotency, idemKey,
		idempotency.HashBody(req.Raw), http.StatusOK, body); err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Ошибка сохранения ключа идемпотентности")
		return
	}
	httpx.WriteRaw(w, http.StatusOK, body)
}

// ByID обслуживает GET /api/center/appointments/{id}.
func (h *Handler) ByID(w http.ResponseWriter, r *http.Request, params map[string]string) {
	appointment, err := h.svc.ByID(r.Context(), params["id"])
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			httpx.WriteError(w, http.StatusNotFound, "Запись "+params["id"]+" не найдена")
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "Ошибка чтения записи")
		return
	}
	httpx.WriteRaw(w, http.StatusOK, appointmentInner(appointment))
}

// decodedRequest — разобранная форма вместе с исходным телом.
type decodedRequest struct {
	Request domain.Request
	Raw     []byte
}

// decodeInto разбирает тело запроса и валидирует модель AppointmentRequest.
//
// Ошибки накапливаются в errs (как в pydantic), а не пишутся сразу: FastAPI
// собирает все ошибки запроса в один ответ 422.
func (h *Handler) decodeInto(w http.ResponseWriter, r *http.Request, errs *pyvalidate.Errors) (decodedRequest, bool) {
	body := pybody.Read(r.Body, r.Header.Get("Content-Type"))
	if body.DecodeError {
		httpx.WriteError(w, http.StatusBadRequest, "There was an error parsing the body")
		return decodedRequest{}, false
	}
	if body.Err != nil {
		errs.Add(*body.Err)
		return decodedRequest{}, false
	}
	if body.Empty {
		errs.Add(*pybody.MissingBody())
		return decodedRequest{}, false
	}
	if !pybody.IsObject(body.Value) {
		errs.Add(*pybody.NotADict(body.Value, true))
		return decodedRequest{}, false
	}

	raw := body.Raw
	if !pybody.IsJSON(r.Header.Get("Content-Type")) {
		// Не JSON: pydantic получает строку и сообщает о неверном типе тела.
		raw = nil
	}
	model := pyvalidate.NewModel(body.Value, body.Value)

	req := domain.Request{}
	req.Name, _ = model.Str(pyvalidate.StrField{
		Name: "name", Required: true, MinLength: 1, MaxLength: 80,
	})
	req.Phone, _ = model.Str(pyvalidate.StrField{
		Name: "phone", Required: true, MinLength: 10, MaxLength: 22,
	})
	req.Tasks, _ = model.StrList(pyvalidate.ListField{Name: "tasks"})
	req.PreferredDate, _ = model.Date("preferred_date")
	req.PreferredTime, _ = model.Str(pyvalidate.StrField{
		Name: "preferred_time", Default: domain.TimeOptions[0],
	})
	req.Ideas, _ = model.StrList(pyvalidate.ListField{Name: "ideas", MaxItems: domain.MaxPickedIdeas})
	req.Files, _ = model.StrList(pyvalidate.ListField{Name: "files", MaxItems: domain.MaxFiles})
	req.Wish, _ = model.Str(pyvalidate.StrField{
		Name: "wish", MaxLength: domain.MaxWishLen,
	})
	req.Consent, _ = model.Bool(pyvalidate.BoolField{Name: "consent"})

	for _, err := range model.Errors().Items() {
		errs.Add(err)
	}
	if errs.Len() > 0 {
		return decodedRequest{}, false
	}
	return decodedRequest{Request: req, Raw: raw}, true
}

// parseApproved разбирает query-параметр approved как bool в lax-режиме.
func parseApproved(raw string, present bool, errs *pyvalidate.Errors) (bool, bool) {
	if !present {
		return false, true
	}
	value := pyjson.NewString(raw)
	switch strings.ToLower(raw) {
	case "true", "yes", "on", "1", "t", "y":
		return true, true
	case "false", "no", "off", "0", "f", "n":
		return false, true
	}
	errs.Add(pyvalidate.Error{
		Type:  "bool_parsing",
		Loc:   pyvalidate.Query("approved"),
		Msg:   "Input should be a valid boolean, unable to interpret input",
		Input: value,
	})
	return false, false
}

// validationBody повторяет структуру AppointmentValidation из mock API.
func validationBody(v domain.Validation) *jsonx.Value {
	obj := jsonx.NewObject()
	obj.Set("ok", jsonx.NewBool(v.OK))
	obj.Set("errors", stringValues(nonNil(v.Errors)))
	obj.Set("warnings", stringValues(nonNil(v.Warnings)))
	obj.Set("summary", stringValues(v.Summary))
	return obj
}

// appointmentBody формирует ответ AppointmentResponse.
//
// Поле idempotent_replay печатается всегда: в mock API это часть модели
// ответа, поэтому первый ответ содержит явное false.
func appointmentBody(a domain.Appointment, replay bool) []byte {
	body := make([]byte, 0, 1024)
	body = append(body, `{"id":`...)
	body = append(body, quote(a.ID)...)
	body = append(body, `,"status":`...)
	body = append(body, quote(a.Status)...)
	body = append(body, `,"created_at":`...)
	body = append(body, quote(a.CreatedAt.Format("2006-01-02T15:04:05.000000+00:00"))...)
	body = append(body, `,"idempotent_replay":`...)
	body = append(body, boolLiteral(replay)...)
	body = append(body, `,"appointment":`...)
	body = append(body, appointmentInner(a)...)
	body = append(body, '}')
	return body
}

// appointmentInner формирует сохранённую запись (порядок ключей как в mock API).
func appointmentInner(a domain.Appointment) []byte {
	body := make([]byte, 0, 1024)
	body = append(body, `{"id":`...)
	body = append(body, quote(a.ID)...)
	body = append(body, `,"status":`...)
	body = append(body, quote(a.Status)...)
	body = append(body, `,"name":`...)
	body = append(body, quote(trimSpace(a.Request.Name))...)
	body = append(body, `,"phone":`...)
	body = append(body, quote(trimSpace(a.Request.Phone))...)
	body = append(body, `,"tasks":`...)
	body = append(body, stringList(a.Request.Tasks)...)
	body = append(body, `,"preferred_date":`...)
	if a.Request.PreferredDate == "" {
		body = append(body, `null`...)
	} else {
		body = append(body, quote(a.Request.PreferredDate)...)
	}
	body = append(body, `,"preferred_time":`...)
	body = append(body, quote(a.Request.PreferredTime)...)
	body = append(body, `,"ideas":`...)
	body = append(body, stringList(a.Request.Ideas)...)
	body = append(body, `,"files":`...)
	body = append(body, stringList(a.Request.Files)...)
	body = append(body, `,"wish":`...)
	body = append(body, quote(trimSpace(a.Request.Wish))...)
	body = append(body, `,"consent":`...)
	body = append(body, boolLiteral(a.Request.Consent)...)
	body = append(body, `,"place":`...)
	body = append(body, quote(domain.PlaceAddress)...)
	body = append(body, `,"validation":`...)
	body = append(body, jsonx.Marshal(validationBody(a.Validation))...)
	body = append(body, `,"created_at":`...)
	body = append(body, quote(a.CreatedAt.Format("2006-01-02T15:04:05.000000+00:00"))...)
	body = append(body, '}')
	return body
}

// withReplayFlag добавляет idempotent_replay:true в сохранённый ответ.
func withReplayFlag(stored []byte) ([]byte, error) {
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(stored, &payload); err != nil {
		return nil, err
	}

	out := make([]byte, 0, len(stored)+32)
	out = append(out, `{"id":`...)
	out = append(out, payload["id"]...)
	out = append(out, `,"status":`...)
	out = append(out, payload["status"]...)
	out = append(out, `,"created_at":`...)
	out = append(out, payload["created_at"]...)
	out = append(out, `,"idempotent_replay":true,"appointment":`...)
	out = append(out, payload["appointment"]...)
	out = append(out, '}')
	return out, nil
}

func nonNil(list []string) []string {
	if list == nil {
		return []string{}
	}
	return list
}

func randomHex(n int) string {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "0000000000"[:n]
	}
	return hex.EncodeToString(buf)
}
