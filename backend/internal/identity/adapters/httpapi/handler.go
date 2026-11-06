// Package httpapi — HTTP-адаптеры контекста Identity.
package httpapi

import (
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/zetta/zettacalc/backend/internal/identity/domain"
	"github.com/zetta/zettacalc/backend/internal/platform/auth"
	"github.com/zetta/zettacalc/backend/internal/platform/httpx"
	"github.com/zetta/zettacalc/backend/internal/platform/jsonx"
	"github.com/zetta/zettacalc/backend/internal/platform/pyjson"
	"github.com/zetta/zettacalc/backend/internal/platform/pyvalidate"
	"golang.org/x/crypto/bcrypt"
)

// Handler обслуживает вход и профиль.
type Handler struct {
	svc    *domain.Service
	tokens *auth.Service
}

// NewHandler создаёт обработчик авторизации.
func NewHandler(svc *domain.Service, tokens *auth.Service) *Handler {
	return &Handler{svc: svc, tokens: tokens}
}

// Login обслуживает POST /api/auth/login (форма OAuth2 password).
//
// Форма разбирается как тело модели OAuth2PasswordRequestForm: пустые значения
// считаются отсутствующими полями, поэтому mock API отвечает 422, а не 401.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	if err := r.ParseForm(); err != nil {
		httpx.WriteError(w, http.StatusBadRequest, "There was an error parsing the body")
		return
	}
	errs := &pyvalidate.Errors{}
	username, okUser := formField(r.PostForm, "username", errs)
	password, okPass := formField(r.PostForm, "password", errs)
	if !okUser || !okPass {
		httpx.WriteRaw(w, http.StatusUnprocessableEntity, errs.Body())
		return
	}

	user, err := h.svc.ByEmail(r.Context(), username)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			httpx.WriteError(w, http.StatusUnauthorized, domain.ErrInvalidCredentials.Error())
			return
		}
		httpx.WriteError(w, http.StatusInternalServerError, "Ошибка чтения пользователя")
		return
	}
	if !checkPassword(user.PasswordHash, password) {
		httpx.WriteError(w, http.StatusUnauthorized, domain.ErrInvalidCredentials.Error())
		return
	}

	token, err := h.tokens.Issue(user.ID, user.Email, user.Role)
	if err != nil {
		httpx.WriteError(w, http.StatusInternalServerError, "Ошибка выпуска токена")
		return
	}
	obj := jsonx.NewObject()
	obj.Set("access_token", jsonx.NewString(token))
	obj.Set("token_type", jsonx.NewString("bearer"))
	obj.Set("user_id", jsonx.NewString(user.ID))
	obj.Set("email", jsonx.NewString(user.Email))
	httpx.WriteJSON(w, http.StatusOK, obj)
}

// Me обслуживает GET /api/auth/me.
func (h *Handler) Me(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	claims, err := auth.RequireToken(r)
	if err != nil {
		httpx.WriteUnauthorized(w, err)
		return
	}
	user, err := h.svc.ByEmail(r.Context(), claims.Email)
	if err != nil {
		httpx.WriteUnauthorized(w, auth.ErrInvalidToken)
		return
	}
	obj := jsonx.NewObject()
	obj.Set("user_id", jsonx.NewString(user.ID))
	obj.Set("email", jsonx.NewString(user.Email))
	obj.Set("name", jsonx.NewString(user.Name))
	httpx.WriteJSON(w, http.StatusOK, obj)
}

// formField читает поле формы: пустое значение pydantic считает отсутствующим.
func formField(form url.Values, name string, errs *pyvalidate.Errors) (string, bool) {
	values, ok := form[name]
	if !ok || len(values) == 0 || values[0] == "" {
		errs.Add(pyvalidate.Error{
			Type:  "missing",
			Loc:   pyvalidate.Body(name),
			Msg:   "Field required",
			Input: pyjson.NewNull(),
		})
		return "", false
	}
	return values[0], true
}

// checkPassword поддерживает и bcrypt-хэши, и открытый пароль демо-пользователя.
func checkPassword(hash, password string) bool {
	if strings.HasPrefix(hash, "$2") {
		return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
	}
	return hash == password
}
