// Package httpapi — корневой HTTP-слой сервиса: маршруты и middleware.
//
// Пути и коды ответов совпадают с mock API (vue-prototypes/api/main.py):
// цель миграции — паритет, а не «более правильный» API.
package httpapi

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	assetshttp "github.com/zetta/zettacalc/backend/internal/assets/adapters/httpapi"
	assetsdomain "github.com/zetta/zettacalc/backend/internal/assets/domain"
	bookinghttp "github.com/zetta/zettacalc/backend/internal/booking/adapters/httpapi"
	bookingdomain "github.com/zetta/zettacalc/backend/internal/booking/domain"
	contenthttp "github.com/zetta/zettacalc/backend/internal/content/adapters/httpapi"
	contentdomain "github.com/zetta/zettacalc/backend/internal/content/domain"
	designhttp "github.com/zetta/zettacalc/backend/internal/design/adapters/httpapi"
	designapp "github.com/zetta/zettacalc/backend/internal/design/app"
	designdomain "github.com/zetta/zettacalc/backend/internal/design/domain"
	estimatehttp "github.com/zetta/zettacalc/backend/internal/estimate/adapters/httpapi"
	estimatedomain "github.com/zetta/zettacalc/backend/internal/estimate/domain"
	identityhttp "github.com/zetta/zettacalc/backend/internal/identity/adapters/httpapi"
	identitydomain "github.com/zetta/zettacalc/backend/internal/identity/domain"
	mediahttp "github.com/zetta/zettacalc/backend/internal/media/adapters/httpapi"
	mediadomain "github.com/zetta/zettacalc/backend/internal/media/domain"
	orderhttp "github.com/zetta/zettacalc/backend/internal/order/adapters/httpapi"
	orderdomain "github.com/zetta/zettacalc/backend/internal/order/domain"
	"github.com/zetta/zettacalc/backend/internal/platform/auth"
	"github.com/zetta/zettacalc/backend/internal/platform/httpx"
	"github.com/zetta/zettacalc/backend/internal/platform/idempotency"
	techcardhttp "github.com/zetta/zettacalc/backend/internal/techcard/adapters/httpapi"
	techcarddomain "github.com/zetta/zettacalc/backend/internal/techcard/domain"
	tenderhttp "github.com/zetta/zettacalc/backend/internal/tender/adapters/httpapi"
	tenderdomain "github.com/zetta/zettacalc/backend/internal/tender/domain"
)

// Deps — зависимости корневого HTTP-слоя.
type Deps struct {
	Pool       *pgxpool.Pool
	Media      *mediadomain.Service
	Content    contentdomain.Repository
	Assets     *assetsdomain.Service
	Design     *designdomain.Service
	Derive     *designapp.Derive
	Estimate   *estimatedomain.Service
	Techcards  *techcarddomain.Service
	Tender     *tenderdomain.Service
	Booking    *bookingdomain.Service
	Identity   *identitydomain.Service
	Order      *orderdomain.Service
	Tokens     *auth.Service
	Idempotent *idempotency.Store
	StartedAt  time.Time
}

// New собирает корневой маршрутизатор.
func New(deps Deps) http.Handler {
	router := httpx.NewRouter()

	media := mediahttp.NewHandler(deps.Media)
	content := contenthttp.NewHandler(deps.Content)
	assets := assetshttp.NewHandler(deps.Assets)
	design := designhttp.NewHandler(deps.Design, deps.Derive)
	estimate := estimatehttp.NewHandler(deps.Estimate)
	techcard := techcardhttp.NewHandler(deps.Techcards)
	tender := tenderhttp.NewHandler(deps.Tender)
	booking := bookinghttp.NewHandler(deps.Booking, deps.Idempotent)
	identity := identityhttp.NewHandler(deps.Identity, deps.Tokens)
	order := orderhttp.NewHandler(deps.Order)
	health := NewHealthHandler(deps)

	// ── Авторизация ────────────────────────────────────────────────────────
	router.Post("/api/auth/login", identity.Login)
	router.Get("/api/auth/me", identity.Me)

	// ── Дизайн-варианты и производные сметы ───────────────────────────────
	router.Get("/api/design/variants", design.Variants)
	router.Get("/api/design/variants/{id}/estimate", design.Estimate)
	router.Get("/api/design/variants/{id}/cpr", design.CPR)
	router.Get("/api/design/variants/{id}/bids", design.Bids)

	// ── Интерьеры ─────────────────────────────────────────────────────────
	router.Get("/api/interiors", content.Interiors)
	router.Get("/api/interiors/categories", content.InteriorsCategories)
	router.Get("/api/interiors/descriptions", content.InteriorsDescriptions)
	router.Get("/api/interiors/{id}", content.InteriorDetail)

	// ── Каталог ───────────────────────────────────────────────────────────
	router.Get("/api/catalog/products", content.Products)
	router.Get("/api/catalog/rooms", content.CatalogRooms)
	router.Get("/api/catalog/rooms/{id}", content.CatalogRoomDetail)

	// ── Интерьерный центр ─────────────────────────────────────────────────
	router.Get("/api/center/interiors", content.CenterInteriors)
	router.Get("/api/center/photos", content.CenterPhotos)

	// ── Hotspot ───────────────────────────────────────────────────────────
	router.Get("/api/hotspot/data", content.HotspotData)
	router.Get("/api/hotspot/analogues", content.HotspotAnalogues)

	// ── Секции главной страницы ───────────────────────────────────────────
	router.Get("/api/main-page/start", assets.MainSection("start"))
	router.Get("/api/main-page/how", assets.MainSection("how"))
	router.Get("/api/main-page/ideas", assets.MainSection("ideas"))
	router.Get("/api/main-page/connected", assets.MainSection("connected"))

	// ── Совместимость со старыми эндпоинтами ──────────────────────────────
	router.Get("/api/references", content.References)
	router.Get("/api/references/{id}", content.ReferenceDetail)
	router.Get("/api/techcards", techcard.List)
	// mock API объявляет расчёт сметы только как POST.
	router.Post("/api/estimate/calculate", estimate.Calculate)
	router.Get("/api/estimate/{id}", estimate.ByID)
	router.Get("/api/cpr/{id}", estimate.CPR)
	router.Post("/api/cpr/generate", estimate.GenerateCPR)
	router.Get("/api/cart", order.Cart)
        router.Post("/api/cart/items", order.AddItem)
        router.Handle("DELETE", "/api/cart/items/{sku}", order.RemoveItem)
	router.Get("/api/bids", tender.Static)
	router.Get("/api/planners", content.Planners)

	// ── Запись в Интерьерный центр (write) ────────────────────────────────
	router.Get("/api/center/appointment/options", booking.Options)
	router.Post("/api/center/appointment/dry-run", booking.DryRun)
	router.Post("/api/center/appointment", booking.Create)
	router.Get("/api/center/appointments/{id}", booking.ByID)

	// ── Приложения и разметка прототипов ──────────────────────────────────
	router.Get("/api/apps", assets.AppsIndex)
	router.Get("/api/apps/{app}/css", assets.AppCSS)
	router.Get("/api/apps/{app}/html", assets.AppHTML)
	router.Get("/api/page", assets.PageIndex)
	router.Get("/api/page/{proto}/{file}", assets.PageFragment)
	router.Get("/api/css/{proto}", assets.ProtoCSS)
	router.Get("/api/css/{proto}/shell", assets.ProtoShellCSS)

	// ── Изображения ───────────────────────────────────────────────────────
	router.Get("/api/images", media.Index)
	router.Get("/api/images/file/{filename}", media.ByFilename)
	router.Get("/api/images/{key}", media.ByKey)

	// ── Служебное ─────────────────────────────────────────────────────────
	router.Get("/health", health.Handle)
	router.Get("/api/health", health.Handle)

	return recoverer(cors(deps.Tokens.Middleware(router)))
}

// cors повторяет настройки mock API: любые источники и методы.
func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Access-Control-Allow-Origin", "*")
		h.Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		h.Set("Access-Control-Allow-Headers", "*")
		h.Set("Access-Control-Allow-Credentials", "true")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// recoverer не даёт панике в обработчике уронить сервис.
func recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				slog.Error("httpapi: паника в обработчике", "path", r.URL.Path, "panic", rec)
				httpx.WriteError(w, http.StatusInternalServerError, "Внутренняя ошибка сервиса")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
