// Package httpapi — HTTP-адаптеры контекста Order.
package httpapi

import (
	"bytes"
	"math/big"
	"net/http"

	"github.com/zetta/zettacalc/backend/internal/order/domain"
	"github.com/zetta/zettacalc/backend/internal/platform/auth"
	"github.com/zetta/zettacalc/backend/internal/platform/httpx"
	"github.com/zetta/zettacalc/backend/internal/platform/pybody"
	"github.com/zetta/zettacalc/backend/internal/platform/pyjson"
	"github.com/zetta/zettacalc/backend/internal/platform/pyvalidate"
)

// Handler отдаёт корзину.
type Handler struct {
	svc *domain.Service
}

// NewHandler создаёт обработчик корзины.
func NewHandler(svc *domain.Service) *Handler { return &Handler{svc: svc} }

// Cart обслуживает GET /api/cart.
func (h *Handler) Cart(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	if _, err := auth.RequireToken(r); err != nil {
		httpx.WriteUnauthorized(w, err)
		return
	}
	cart, err := h.svc.Cart(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Ошибка чтения корзины")
		return
	}

	body := &bytes.Buffer{}
	body.WriteString(`{"items":[`)
	for i, it := range cart.Items {
		if i > 0 {
			body.WriteByte(',')
		}
		body.WriteString(`{"sku":`)
		pyjson.WriteString(body, it.SKU)
		body.WriteString(`,"brand":`)
		pyjson.WriteString(body, it.Brand)
		body.WriteString(`,"title":`)
		pyjson.WriteString(body, it.Title)
		body.WriteString(`,"price":`)
		body.WriteString(pyjson.FormatNumber(it.Price))
		body.WriteString(`,"qty":`)
		body.WriteString(big.NewInt(int64(it.Qty)).String())
		body.WriteString(`,"image":`)
		pyjson.WriteString(body, it.ImageKey)
		body.WriteByte('}')
	}
	body.WriteString(`],"total_rub":`)
	body.WriteString(big.NewInt(cart.TotalRub).String())
	body.WriteByte('}')
	httpx.WriteRaw(w, http.StatusOK, body.Bytes())
}

// AddItem обслуживает POST /api/cart/items.
//
// Mock API отвечает только добавленной позицией, не сохраняя её: тело ответа
// повторяет присланный объект, а total_rub считается как int(price) * int(qty)
// — включая питоновские тонкости int("1_0") == 10 и int(2.7) == 2. Поведение
// сохранено, чтобы не менять контракт фронтендов.
func (h *Handler) AddItem(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	// FastAPI разбирает тело раньше зависимостей: синтаксическая ошибка JSON
	// отвечает 422 даже без токена.
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

	price, err := intOf(body.Value.Field("price"), 0)
	if err != nil {
		// mock API падает на int("abc") и отдаёт 500 Internal Server Error.
		httpx.WriteInternalError(w)
		return
	}
	qty, err := intOf(body.Value.Field("qty"), 1)
	if err != nil {
		httpx.WriteInternalError(w)
		return
	}

	total := new(big.Int).Mul(price, qty)

	// Ответ повторяет тело запроса, поэтому non-finite литерал в нём ломает
	// сериализацию: mock API отвечает 500, хотя разбор тела прошёл.
	if pyjson.HasNonFinite(body.Value) {
		httpx.WriteInternalError(w)
		return
	}

	out := &bytes.Buffer{}
	out.WriteString(`{"items":[`)
	out.Write(pyjson.Encode(body.Value))
	out.WriteString(`],"total_rub":`)
	out.WriteString(total.String())
	out.WriteByte('}')
	httpx.WriteRaw(w, http.StatusOK, out.Bytes())
}

// RemoveItem обслуживает DELETE /api/cart/items/{sku}.
func (h *Handler) RemoveItem(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if _, err := auth.RequireToken(r); err != nil {
		httpx.WriteUnauthorized(w, err)
		return
	}

	sku := params["sku"]
	if sku == "" {
		httpx.WriteError(w, http.StatusBadRequest, "sku is required")
		return
	}

	cart, err := h.svc.Cart(r.Context())
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Ошибка чтения корзины")
		return
	}

	removed := false
	for i, it := range cart.Items {
		if it.SKU == sku {
			cart.Items = append(cart.Items[:i], cart.Items[i+1:]...)
			removed = true
			break
		}
	}
	if !removed {
		httpx.WriteError(w, http.StatusNotFound, "Item not found in cart")
		return
	}

	// пересчитаем total
	var total int64
	for _, it := range cart.Items {
		total += int64(it.Price) * int64(it.Qty)
	}

	out := &bytes.Buffer{}
	out.WriteString(`{"items":[`)
	for i, it := range cart.Items {
		if i > 0 {
			out.WriteByte(',')
		}
		out.WriteString(`{"sku":`)
		pyjson.WriteString(out, it.SKU)
		out.WriteString(`,"brand":`)
		pyjson.WriteString(out, it.Brand)
		out.WriteString(`,"title":`)
		pyjson.WriteString(out, it.Title)
		out.WriteString(`,"price":`)
		out.WriteString(pyjson.FormatNumber(it.Price))
		out.WriteString(`,"qty":`)
		out.WriteString(big.NewInt(int64(it.Qty)).String())
		out.WriteString(`,"image":`)
		pyjson.WriteString(out, it.ImageKey)
		out.WriteByte('}')
	}
	out.WriteString(`],"total_rub":`)
	out.WriteString(big.NewInt(total).String())
	out.WriteByte('}')
	httpx.WriteRaw(w, http.StatusOK, out.Bytes())
}

// intOf повторяет item.get(key, default) с последующим int().
//
// Ключ с null — это int(None): TypeError, то есть 500 в mock API. Значение по
// умолчанию подставляется только при отсутствии ключа.
func intOf(v *pyjson.Value, fallback int64) (*big.Int, error) {
	if v == nil {
		return big.NewInt(fallback), nil
	}
	return pyjson.AsInt(v)
}
