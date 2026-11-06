// Package httpx — минимальный HTTP-слой: роутер на stdlib и JSON-ответы.
//
// Паритет с mock API важнее «красоты» роутинга, поэтому маршруты объявляются
// строками вручную, а тела ответов отдаются байт-в-байт из БД там, где это
// возможно (см. payload_text в схеме).
package httpx

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/zetta/zettacalc/backend/internal/platform/auth"
	"github.com/zetta/zettacalc/backend/internal/platform/jsonx"
	"github.com/zetta/zettacalc/backend/internal/platform/pyvalidate"
)

// HandlerFunc — обработчик с доступом к параметрам пути.
type HandlerFunc func(w http.ResponseWriter, r *http.Request, params map[string]string)

// Router — простой роутер с поддержкой параметров вида {id}.
type Router struct {
	mux    *http.ServeMux
	routes []route
}

type route struct {
	method   string
	segments []string
	handler  HandlerFunc
}

// NewRouter создаёт пустой роутер.
func NewRouter() *Router {
	return &Router{mux: http.NewServeMux()}
}

// Handle регистрирует маршрут. Шаблон: "/api/design/variants/{id}/estimate".
func (rt *Router) Handle(method, pattern string, h HandlerFunc) {
	rt.routes = append(rt.routes, route{
		method:   strings.ToUpper(method),
		segments: splitPath(pattern),
		handler:  h,
	})
}

// Get — сокращение для Handle(http.MethodGet, …).
func (rt *Router) Get(pattern string, h HandlerFunc) { rt.Handle(http.MethodGet, pattern, h) }

// Post — сокращение для Handle(http.MethodPost, …).
func (rt *Router) Post(pattern string, h HandlerFunc) { rt.Handle(http.MethodPost, pattern, h) }

// Delete — сокращение для Handle(http.MethodDelete, …).
func (rt *Router) Delete(pattern string, h HandlerFunc) { rt.Handle(http.MethodDelete, pattern, h) }

// ServeHTTP сопоставляет запрос с зарегистрированными маршрутами.
func (rt *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSuffix(r.URL.Path, "/")
	if path == "" {
		path = "/"
	}
	segs := splitPath(path)

	var allowed []string
	for _, rt2 := range rt.routes {
		params, ok := match(rt2.segments, segs)
		if !ok {
			continue
		}
		if rt2.method != r.Method {
			allowed = append(allowed, rt2.method)
			continue
		}
		rt2.handler(w, r, params)
		return
	}
	if len(allowed) > 0 {
		w.Header().Set("Allow", strings.Join(allowed, ", "))
		WriteError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		return
	}
	WriteError(w, http.StatusNotFound, "Not Found")
}

func splitPath(p string) []string {
	p = strings.Trim(p, "/")
	if p == "" {
		return nil
	}
	return strings.Split(p, "/")
}

func match(pattern, actual []string) (map[string]string, bool) {
	if len(pattern) != len(actual) {
		return nil, false
	}
	params := make(map[string]string, 2)
	for i, seg := range pattern {
		if strings.HasPrefix(seg, "{") && strings.HasSuffix(seg, "}") {
			params[strings.Trim(seg, "{}")] = actual[i]
			continue
		}
		if seg != actual[i] {
			return nil, false
		}
	}
	return params, true
}

// ContentTypeJSON — content-type ответов mock API (FastAPI JSONResponse).
const ContentTypeJSON = "application/json"

// WriteJSON отдаёт значение как компактный JSON без HTML-экранирования.
//
// Значение типа *jsonx.Value пишется упорядоченным сериализатором: порядок
// ключей и запись чисел должны совпадать с ответом Python-mock байт-в-байт,
// а encoding/json сортирует ключи map и не воспроизводит repr() чисел.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	if value, ok := v.(*jsonx.Value); ok {
		WriteRaw(w, status, jsonx.Marshal(value))
		return
	}
	body, err := marshalJSON(v)
	if err != nil {
		slog.Error("httpx: сериализация JSON", "err", err)
		WriteError(w, http.StatusInternalServerError, "Внутренняя ошибка сервиса")
		return
	}
	WriteRaw(w, status, body)
}

func marshalJSON(v any) ([]byte, error) {
	if v == nil {
		return []byte("null"), nil
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return bytes.TrimRight(buf.Bytes(), "\n"), nil
}

// WriteRaw отдаёт заранее подготовленные JSON-байты (паритет с mock API).
func WriteRaw(w http.ResponseWriter, status int, body []byte) {
	w.Header().Set("Content-Type", ContentTypeJSON)
	w.WriteHeader(status)
	if _, err := w.Write(body); err != nil {
		slog.Error("httpx: запись тела", "err", err)
	}
}

// WriteText отдаёт текст с указанным content-type.
func WriteText(w http.ResponseWriter, status int, contentType, body string) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(status)
	if _, err := w.Write([]byte(body)); err != nil {
		slog.Error("httpx: запись текста", "err", err)
	}
}

// WriteBytes отдаёт бинарные данные с кэшированием.
func WriteBytes(w http.ResponseWriter, status int, contentType string, body []byte, cacheControl string) {
	if cacheControl != "" {
		w.Header().Set("Cache-Control", cacheControl)
	}
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(status)
	if _, err := w.Write(body); err != nil {
		slog.Error("httpx: запись байтов", "err", err)
	}
}

// WriteValidation отдаёт 422 с ошибками pydantic.
//
// Использует ResponseBody: если тело запроса не разобралось (json_invalid),
// ошибки query-параметров в ответ не попадают — FastAPI разбирает тело раньше.
//
// Если в ошибке есть NaN или ±Inf (поле input), mock API падает на сериализации
// ответа и отдаёт 500 — json.dumps по умолчанию не допускает такие значения.
func WriteValidation(w http.ResponseWriter, errs *pyvalidate.Errors) {
	if errs.HasNonFinite() {
		WriteInternalError(w)
		return
	}
	WriteRaw(w, http.StatusUnprocessableEntity, errs.ResponseBody())
}

// WriteInternalError повторяет ответ FastAPI на необработанное исключение.
//
// Тело — обычный текст, а не JSON: так отвечает uvicorn на любую ошибку внутри
// обработчика, включая ValueError от json.dumps на NaN и ±Inf.
func WriteInternalError(w http.ResponseWriter) {
	WriteText(w, http.StatusInternalServerError, "text/plain; charset=utf-8", "Internal Server Error")
}

// WriteError — единый формат ошибки, совместимый с FastAPI: {"detail": …}.
func WriteError(w http.ResponseWriter, status int, detail any) {
	body, err := marshalJSON(map[string]any{"detail": detail})
	if err != nil {
		body = []byte(`{"detail":"Внутренняя ошибка сервиса"}`)
	}
	WriteRaw(w, status, body)
}

// WriteUnauthorized отдаёт 401 с текстом mock API.
//
// FastAPI различает два случая: без заголовка Authorization — «Not
// authenticated» (с заголовком WWW-Authenticate: Bearer), с неверным токеном —
// «Invalid token».
func WriteUnauthorized(w http.ResponseWriter, err error) {
	if errors.Is(err, auth.ErrInvalidToken) {
		WriteError(w, http.StatusUnauthorized, "Invalid token")
		return
	}
	w.Header().Set("WWW-Authenticate", "Bearer")
	WriteError(w, http.StatusUnauthorized, "Not authenticated")
}

// ErrHTTP — ошибка с HTTP-статусом и телом detail.
type ErrHTTP struct {
	Status int
	Detail any
}

func (e *ErrHTTP) Error() string { return "http error" }

// NewHTTPError создаёт ошибку уровня транспорта.
func NewHTTPError(status int, detail any) error {
	return &ErrHTTP{Status: status, Detail: detail}
}

// AsHTTPError извлекает ErrHTTP из цепочки ошибок.
func AsHTTPError(err error) (*ErrHTTP, bool) {
	var he *ErrHTTP
	if errors.As(err, &he) {
		return he, true
	}
	return nil, false
}

// DecodeJSON читает тело запроса как JSON.
func DecodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.UseNumber()
	if err := dec.Decode(dst); err != nil {
		return err
	}
	return nil
}
