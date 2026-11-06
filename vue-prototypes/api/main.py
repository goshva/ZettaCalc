"""
Fresh House Mock API — эндпоинты данных из прототипа А + заглушки
«4 варианта дизайна → смета + ЦПР».

Запуск:  cd vue-prototypes/api && python -m uvicorn main:app --port 8000 --reload
"""

from __future__ import annotations

import hashlib
import json
import re
import uuid
from datetime import date, datetime, timedelta, timezone
from pathlib import Path
from typing import Any

from fastapi import Depends, FastAPI, Header, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import FileResponse
from fastapi.security import OAuth2PasswordBearer, OAuth2PasswordRequestForm
from jose import JWTError, jwt
from pydantic import BaseModel, Field

STATIC_DIR = Path(__file__).parent / "static"
IMAGES_DIR = STATIC_DIR / "images"
DATA_DIR = Path(__file__).parent / "data"
MAIN_PAGE_DIR = DATA_DIR / "main_page"
APPS_DIR = DATA_DIR / "apps52"

SECRET_KEY = "fresh-house-dev-key-2026-not-for-prod"
ALGORITHM = "HS256"
TOKEN_EXPIRE = timedelta(days=7)

app = FastAPI(title="Fresh House API", version="2.0.0",
              description="Mock REST API для Zetta Home прототипов A и B")

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

oauth2_scheme = OAuth2PasswordBearer(tokenUrl="/api/auth/login")

# ── Данные ─────────────────────────────────────────────────────────────────

USERS = {"test@zetta.ru": {"id": "u1", "password": "test123", "email": "test@zetta.ru", "name": "Тест"}}

# ── Изображения ────────────────────────────────────────────────────────────
IMAGE_KEYS: dict[str, dict] = {
    # — контент (БД / PIM) —
    "idea-kitchen": {"file": "idea-kitchen.jpg", "source": "db"},
    "idea-living":  {"file": "idea-living.jpg",  "source": "db"},
    "idea-bath":    {"file": "idea-bath.jpg",    "source": "db"},
    "idea-bedroom": {"file": "idea-bedroom.jpg", "source": "db"},
    "idea-teen":    {"file": "idea-teen.jpg",    "source": "db"},
    "idea-study":   {"file": "idea-study.jpg",   "source": "db"},
    "cat-furniture":{"file": "cat-furniture.jpg","source": "db"},
    "cat-materials":{"file": "cat-materials.jpg","source": "db"},
    "cat-bath":     {"file": "cat-bath.jpg",     "source": "db"},
    "cat-decor":    {"file": "cat-decor.jpg",    "source": "db"},
    "connected":    {"file": "connected.jpg",    "source": "db"},
    "sofa":         {"file": "sofa.jpg",         "source": "db"},
    "pendant":      {"file": "pendant.jpg",      "source": "db"},
    "rug":          {"file": "rug.jpg",          "source": "db"},
    "c-sofa":       {"file": "c-sofa.jpg",       "source": "db"},
    "c-dining":     {"file": "c-dining.jpg",     "source": "db"},
    "c-lounge":     {"file": "c-lounge.jpg",     "source": "db"},
    "c-wardrobe":   {"file": "c-wardrobe.jpg",   "source": "db"},
    "c-bedroom":    {"file": "c-bedroom.jpg",    "source": "db"},
    "c-floors":     {"file": "c-floors.jpg",     "source": "db"},
    "c-stone":      {"file": "c-stone.jpg",      "source": "db"},
    "c-decor":      {"file": "c-decor.jpg",      "source": "db"},
    "c-sockets":    {"file": "c-sockets.jpg",    "source": "db"},
    "c-dark":       {"file": "c-dark.jpg",       "source": "db"},
    # — декор интерфейса (статика) —
    "hero":             {"file": "hero.jpg",             "source": "static"},
    "video":            {"file": "video.jpg",            "source": "static"},
    "app":              {"file": "app.jpg",              "source": "static"},
    "cpr":              {"file": "cpr.jpg",              "source": "static"},
    "expert":           {"file": "expert.jpg",           "source": "static"},
    "auction":          {"file": "auction.jpg",          "source": "static"},
    "ceo":              {"file": "ceo.jpg",              "source": "static"},
    "plan":             {"file": "plan.jpg",             "source": "static"},
    "repair":           {"file": "repair.jpg",           "source": "static"},
    "center-store":     {"file": "center-store.jpg",     "source": "static"},
    "center-pano":      {"file": "center-pano.jpg",      "source": "static"},
    "center-kitchen":   {"file": "center-kitchen.jpg",   "source": "static"},
    "center-stone":     {"file": "center-stone.jpg",     "source": "static"},
    "center-wardrobe":  {"file": "center-wardrobe.jpg",  "source": "static"},
    "ds-e06b4936da":    {"file": "ds-e06b4936da.jpg",    "source": "static"},
}

# ── Загружаем данные из файлов ─────────────────────────────────────────────

def _load_json(path: Path) -> Any:
    if path.exists():
        with open(path, "r", encoding="utf-8") as f:
            return json.load(f)
    return None

# hotspotData60
HOTSPOT_DATA = _load_json(DATA_DIR / "proto_a_hotspotData60.json") or []

# extraAnalogues64
EXTRA_ANALOGUES = _load_json(DATA_DIR / "proto_a_extraAnalogues64.json") or []

# dataset (interiors)
DATASET = _load_json(DATA_DIR / "apps52/interiors/dataset.json") or {}
INTERIORS = DATASET.get("items", [])
CATEGORIES = DATASET.get("categories", [])
DESCRIPTIONS = DATASET.get("descriptions", [])

# catalog data
CATALOG_DATA = _load_json(DATA_DIR / "apps52/catalog/data.json") or {}
PRODUCTS = CATALOG_DATA.get("products", [])
ROOMS = CATALOG_DATA.get("rooms", [])

# center data
CENTER_INTERIORS = _load_json(DATA_DIR / "apps52/center/interiors.json") or []
CENTER_PHOTOS = _load_json(DATA_DIR / "apps52/center/photos.json") or []

# ── Mock-данные фронтендов (совместимость со старыми эндпоинтами) ──────────

REFERENCES = [
    {"id": "ref_001", "title": "Кухня в тёплых тонах с островом", "category": "kitchen",
     "image": "idea-kitchen", "styles": ["modern", "warm"], "price_from": 450000, "price_to": 1200000},
    {"id": "ref_002", "title": "Гостиная в тёплых тонах", "category": "living",
     "image": "idea-living", "styles": ["modern", "cozy"], "price_from": 320000, "price_to": 890000},
    {"id": "ref_003", "title": "Санузел в минимализме", "category": "bathroom",
     "image": "idea-bath", "styles": ["minimalist", "modern"], "price_from": 180000, "price_to": 520000},
    {"id": "ref_004", "title": "Спальня с мягким изголовьем", "category": "bedroom",
     "image": "idea-bedroom", "styles": ["modern", "soft"], "price_from": 250000, "price_to": 680000},
    {"id": "ref_005", "title": "Подростковая", "category": "bedroom",
     "image": "idea-teen", "styles": ["modern", "bright"], "price_from": 210000, "price_to": 540000},
    {"id": "ref_006", "title": "Кабинет и библиотека", "category": "living",
     "image": "idea-study", "styles": ["classical", "calm"], "price_from": 280000, "price_to": 720000},
]

CONNECTED = {
    "image": "connected",
    "title": "Гостиная с подборкой товаров",
    "items": [
        {"sku": "FUR-SOF-014", "brand": "ZETTA", "title": "Диван угловой", "price": 148000, "image": "sofa"},
        {"sku": "LGT-PND-007", "brand": "EGLO", "title": "Подвесной светильник", "price": 12400, "image": "pendant"},
        {"sku": "DEC-RUG-021", "brand": "Oriental Rugs", "title": "Ковёр шерстяной", "price": 34900, "image": "rug"},
    ],
}

TECH_CARDS = [
    {"id": "tc_001", "code": "FLOOR_KER_600", "title": "Укладка плитки 600×600",
     "surface": "floor", "material": "ceramic_tile",
     "operations": ["Демонтаж", "Выравнивание", "Гидроизоляция", "Укладка", "Затирка"],
     "consumption": {"klej_m2": 5.0, "grout_m2": 0.3},
     "min_area_m2": 3.0, "max_unevenness": 3.0, "version": 3},
    {"id": "tc_002", "code": "WALL_PLASTER", "title": "Штукатурка гипсовая",
     "surface": "wall", "material": "gypsum_plaster",
     "operations": ["Грунтовка", "Маяки", "Штукатурка", "Накрывка"],
     "consumption": {"primer_l_m2": 0.15, "plaster_kg_m2": 10.0},
     "min_area_m2": 5.0, "max_unevenness": None, "version": 2},
]

ESTIMATES = {
    "est_001": {
        "id": "est_001", "status": "preliminary", "total_rub": 847500,
        "range_min": 720000, "range_max": 980000, "days": 45,
        "lines": [
            {"id": "l1", "title": "Кухня ZETTA METALICO", "qty": 1, "unit": "комплект", "price": 385000, "sku": "ZET-KIT-MET-001", "brand": "ZETTA"},
            {"id": "l2", "title": "Плитка 600×600 укладка", "qty": 12.5, "unit": "м²", "price": 3200, "sku": None, "brand": None},
            {"id": "l3", "title": "Штукатурка стен", "qty": 65, "unit": "м²", "price": 850, "sku": None, "brand": None},
        ],
        "assumptions": ["Ровность пола требует проверки", "Состояние стен неизвестно"],
    },
}

# Корзина mock API: в памяти процесса, стартовое состояние — демо-позиция.
_cart_state: dict = {
    "items": [{"sku": "ZET-KIT-MET-001", "brand": "ZETTA", "title": "Кухня METALICO",
               "price": 385000, "qty": 1, "image": "idea-kitchen"}],
}


def _cart_total() -> int:
    return sum(int(it.get("price", 0)) * int(it.get("qty", 1)) for it in _cart_state["items"])

BIDS = [
    {"id": "b1", "name": "РемонтПро", "rating": 4.8, "price": 320000, "days": 40, "scope": "Кухня + санузел", "image": "repair"},
    {"id": "b2", "name": "СтройМастер", "rating": 4.5, "price": 280000, "days": 45, "scope": "Кухня + санузел", "image": "auction"},
    {"id": "b3", "name": "ДизайнРемонт", "rating": 4.9, "price": 410000, "days": 35, "scope": "Полный ремонт + дизайн", "image": "expert"},
]

PLANNERS = [
    {"id": "p1", "name": "Квартира 90 м² — стандарт", "area": 90.0,
     "rooms": [{"name": "Кухня-гостиная", "area": 28}, {"name": "Спальня", "area": 18}]},
    {"id": "p2", "name": "Квартира 90 м² — перепланировка", "area": 92.0,
     "rooms": [{"name": "Кухня-гостиная", "area": 35}, {"name": "Спальня", "area": 18}]},
]

# main page sections
MAIN_PAGE_SECTIONS: dict[str, str] = {}
if MAIN_PAGE_DIR.exists():
    for f in sorted(MAIN_PAGE_DIR.glob("*.html")):
        MAIN_PAGE_SECTIONS[f.stem] = f.read_text(encoding="utf-8")

# ── Заглушки: 4 варианта дизайна → смета + ЦПР ────────────────────────────
# Вместо загрузки фото/анализа — клиент выбирает 1 из 4 вариантов,
# для каждого сразу генерируется смета и ЦПР-документация.

DESIGN_VARIANTS = [
    {
        "id": "variant_01",
        "name": "Контemporари тёплый",
        "style": "contemporari",
        "palette": "warm",
        "image": "/api/images/file/a_main_img_0003.jpg",
        "description": "Современный интерьер в тёплых тонах с натуральными материалами",
        "rooms": ["кухня", "гостиная", "спальня"],
        "estimated_budget_min": 2500000,
        "estimated_budget_max": 4200000,
        "estimated_days": 60,
    },
    {
        "id": "variant_02",
        "name": "Неоклассика",
        "style": "neoclassica",
        "palette": "cool",
        "image": "/api/images/file/a_main_img_0004.jpg",
        "description": "Элегантная неоклассика с мрамором и латунными акцентами",
        "rooms": ["кухня", "гостиная", "кабинет", "спальня"],
        "estimated_budget_min": 3500000,
        "estimated_budget_max": 5800000,
        "estimated_days": 75,
    },
    {
        "id": "variant_03",
        "name": "Минимализм",
        "style": "minimalist",
        "palette": "neutral",
        "image": "/api/images/file/a_main_img_0005.jpg",
        "description": "Чистые линии, встроенная мебель, скрытые системы хранения",
        "rooms": ["кухня-гостиная", "спальня"],
        "estimated_budget_min": 1800000,
        "estimated_budget_max": 3200000,
        "estimated_days": 45,
    },
    {
        "id": "variant_04",
        "name": "Тёплый лофт",
        "style": "loft",
        "palette": "warm",
        "image": "/api/images/file/a_main_img_0006.jpg",
        "description": "Лофт с мягкими текстурами, кирпичом и натуральным деревом",
        "rooms": ["кухня-гостиная", "спальня", "ванна"],
        "estimated_budget_min": 2200000,
        "estimated_budget_max": 3800000,
        "estimated_days": 50,
    },
]

def _make_estimate(variant: dict) -> dict:
    """Генерирует смету для варианта дизайна."""
    budget_min = variant["estimated_budget_min"]
    budget_max = variant["estimated_budget_max"]
    total = (budget_min + budget_max) // 2
    return {
        "id": f"est_{variant['id']}",
        "variant_id": variant["id"],
        "variant_name": variant["name"],
        "status": "preliminary",
        "total_rub": total,
        "range_min": budget_min,
        "range_max": budget_max,
        "days": variant["estimated_days"],
        "lines": [
            {"id": "l1", "title": "Мебель и материалы", "qty": 1, "unit": "комплект", "price": total * 0.45},
            {"id": "l2", "title": "Ремонт и отделка", "qty": 1, "unit": "комплект", "price": total * 0.30},
            {"id": "l3", "title": "Сантехника и электрика", "qty": 1, "unit": "комплект", "price": total * 0.10},
            {"id": "l4", "title": "Декор и аксессуары", "qty": 1, "unit": "комплект", "price": total * 0.05},
            {"id": "l5", "title": "Проектирование", "qty": 1, "unit": "услуга", "price": total * 0.10},
        ],
        "assumptions": [
            "Ровность пола требует проверки",
            "Состояние стен неизвестно",
            "Цены указаны ориентировочно",
        ],
    }

def _make_cpr(variant: dict, estimate: dict) -> dict:
    """Генерирует ЦПР для варианта дизайна."""
    cpr_id = "cpr_" + hashlib.md5(variant["id"].encode()).hexdigest()[:8]
    return {
        "id": cpr_id,
        "version": 1,
        "project_id": f"proj_{variant['id']}",
        "estimate": estimate,
        "documents": [
            {"id": "d1", "title": "Планировочное решение", "type": "drawing", "status": "ready"},
            {"id": "d2", "title": "Ведомость отделки", "type": "specification", "status": "ready"},
            {"id": "d3", "title": "Электрическая схема", "type": "drawing", "status": "ready"},
            {"id": "d4", "title": "Смета на работы", "type": "estimate", "status": "ready"},
            {"id": "d5", "title": "График производства работ", "type": "schedule", "status": "pending_review"},
        ],
        "status": "ready",
    }

def _make_bids(variant: dict, estimate: dict) -> list[dict]:
    """Аукцион подрядчиков для варианта дизайна."""
    base_price = estimate["total_rub"]
    return [
        {"id": "b1", "name": "РемонтПро", "rating": 4.8, "price": int(base_price * 0.95), "days": estimate["days"] + 5, "scope": f"Полный ремонт: {', '.join(variant['rooms'])}", "image": "repair"},
        {"id": "b2", "name": "СтройМастер", "rating": 4.5, "price": int(base_price * 0.88), "days": estimate["days"] + 10, "scope": f"Чистовой ремонт: {', '.join(variant['rooms'])}", "image": "auction"},
        {"id": "b3", "name": "ДизайнРемонт", "rating": 4.9, "price": int(base_price * 1.10), "days": estimate["days"] - 5, "scope": f"Дизайн + ремонт: {', '.join(variant['rooms'])}", "image": "expert"},
    ]

# ── Модели ─────────────────────────────────────────────────────────────────

class DesignVariant(BaseModel):
    id: str
    name: str
    style: str
    palette: str
    image: str
    description: str
    rooms: list[str]
    estimated_budget_min: int
    estimated_budget_max: int
    estimated_days: int

class EstimateResponse(BaseModel):
    id: str
    variant_id: str
    variant_name: str
    status: str
    total_rub: int
    range_min: int
    range_max: int
    days: int
    lines: list[dict]
    assumptions: list[str]

class CPRResponse(BaseModel):
    id: str
    version: int
    project_id: str
    estimate: EstimateResponse
    documents: list[dict]
    status: str

class ContractorBid(BaseModel):
    id: str
    name: str
    rating: float
    price: int
    days: int
    scope: str
    image: str = ""

# ── Запись в Интерьерный центр (write-операция) ────────────────────────────
#
# Форма из оригинала (раздел «Центр», блок #appointment): набор задач,
# имя, телефон, дата и время, до 5 сохранённых идей, файлы дизайн-проекта,
# пожелания и согласие на обработку данных.
# Любая запись — только с Idempotency-Key и approval-gate: сначала dry-run,
# затем подтверждение менеджером (approved=true).

CONSENT_REQUIRED = "Не подтверждено согласие на обработку персональных данных"
PHONE_RE = re.compile(r"^[+0-9\s()\-]{10,22}$")
MAX_PICKED_IDEAS = 5
MAX_FILES = 10

TASK_OPTIONS = ["Подбор интерьера", "Мебель и материалы", "Ремонт", "Свой дизайн-проект"]
TIME_OPTIONS = ["Уточнить по телефону", "Первая половина дня", "Вторая половина дня"]

# Записи хранятся в памяти: это mock, перезапуск API очищает журнал.
APPOINTMENTS: dict[str, dict[str, Any]] = {}
# Идемпотентность: повторный POST с тем же ключом возвращает ту же запись.
IDEMPOTENCY_INDEX: dict[str, str] = {}


class AppointmentRequest(BaseModel):
    name: str = Field(min_length=1, max_length=80)
    phone: str = Field(min_length=10, max_length=22)
    tasks: list[str] = Field(default_factory=list)
    preferred_date: date | None = None
    preferred_time: str = TIME_OPTIONS[0]
    ideas: list[str] = Field(default_factory=list, max_length=MAX_PICKED_IDEAS)
    files: list[str] = Field(default_factory=list, max_length=MAX_FILES)
    wish: str = Field(default="", max_length=1200)
    consent: bool = False


class AppointmentValidation(BaseModel):
    ok: bool
    errors: list[str]
    warnings: list[str]
    summary: list[str]


class AppointmentResponse(BaseModel):
    id: str
    status: str
    created_at: str
    idempotent_replay: bool = False
    appointment: dict[str, Any]


def _validate_appointment(req: AppointmentRequest) -> AppointmentValidation:
    """Code-based проверка формы. Никакой LLM: правила только в коде."""
    errors: list[str] = []
    warnings: list[str] = []

    if not req.name.strip():
        errors.append("Укажите имя")
    if not PHONE_RE.match(req.phone.strip()):
        errors.append("Телефон должен содержать 10-22 символа: цифры, +, скобки, дефисы")
    if not req.consent:
        errors.append(CONSENT_REQUIRED)
    if not req.tasks:
        warnings.append("Не выбрана ни одна задача — менеджер уточнит её на связи")
    unknown = [t for t in req.tasks if t not in TASK_OPTIONS]
    if unknown:
        errors.append(f"Неизвестные задачи: {', '.join(unknown)}")
    if req.preferred_time not in TIME_OPTIONS:
        errors.append("Недопустимое удобное время")
    if req.preferred_date and req.preferred_date < date.today():
        errors.append("Дата визита уже прошла")
    if len(req.ideas) > MAX_PICKED_IDEAS:
        errors.append(f"Для подготовки подборки можно выбрать не более {MAX_PICKED_IDEAS} идей")
    if len(req.files) > MAX_FILES:
        errors.append(f"Не более {MAX_FILES} файлов дизайн-проекта")

    summary = [f"Задачи: {', '.join(req.tasks) or 'не выбраны'}"]
    if req.preferred_date:
        summary.append(f"Дата: {req.preferred_date.isoformat()}, {req.preferred_time}")
    else:
        summary.append(f"Время: {req.preferred_time}")
    if req.ideas:
        summary.append(f"Идей к подборке: {len(req.ideas)}")
    if req.files:
        summary.append(f"Файлов дизайн-проекта: {len(req.files)}")
    if req.wish.strip():
        summary.append("Пожелания заполнены")

    return AppointmentValidation(
        ok=not errors, errors=errors, warnings=warnings, summary=summary
    )


def _appointment_payload(req: AppointmentRequest) -> dict[str, Any]:
    return {
        "name": req.name.strip(),
        "phone": req.phone.strip(),
        "tasks": req.tasks,
        "preferred_date": req.preferred_date.isoformat() if req.preferred_date else None,
        "preferred_time": req.preferred_time,
        "ideas": req.ideas,
        "files": req.files,
        "wish": req.wish.strip(),
        "consent": req.consent,
        "place": "Москва, Аминьевское шоссе, 6, ТЦ «Квартал», 2 этаж",
    }

# ── JWT ─────────────────────────────────────────────────────────────────────

def create_token(user_id: str, email: str) -> str:
    return jwt.encode({"sub": user_id, "email": email, "exp": datetime.now(timezone.utc) + TOKEN_EXPIRE}, SECRET_KEY, algorithm=ALGORITHM)

def verify_token(token: str = Depends(oauth2_scheme)) -> dict[str, Any]:
    try: return jwt.decode(token, SECRET_KEY, algorithms=[ALGORITHM])
    except JWTError: raise HTTPException(status_code=401, detail="Invalid token")

# ── Эндпоинты ───────────────────────────────────────────────────────────────

@app.post("/api/auth/login")
async def login(form: OAuth2PasswordRequestForm = Depends()):
    user = USERS.get(form.username)
    if not user or user["password"] != form.password: raise HTTPException(401, "Неверный email или пароль")
    return {"access_token": create_token(user["id"], user["email"]), "token_type": "bearer", "user_id": user["id"], "email": user["email"]}

@app.get("/api/auth/me")
async def me(payload: dict = Depends(verify_token)):
    u = USERS[payload["email"]]
    return {"user_id": payload["sub"], "email": payload["email"], "name": u["name"]}

# ── Дизайн-варианты (заглушка вместо загрузки фото) ────────────────────────

@app.get("/api/design/variants", response_model=list[DesignVariant])
async def design_variants():
    """4 варианта дизайна → клиент выбирает один."""
    return DESIGN_VARIANTS

@app.get("/api/design/variants/{variant_id}/estimate", response_model=EstimateResponse)
async def design_estimate(variant_id: str):
    """Смета для выбранного варианта дизайна."""
    variant = next((v for v in DESIGN_VARIANTS if v["id"] == variant_id), None)
    if not variant: raise HTTPException(404, f"Variant {variant_id} not found")
    return _make_estimate(variant)

@app.get("/api/design/variants/{variant_id}/cpr", response_model=CPRResponse)
async def design_cpr(variant_id: str):
    """ЦПР для выбранного варианта дизайна."""
    variant = next((v for v in DESIGN_VARIANTS if v["id"] == variant_id), None)
    if not variant: raise HTTPException(404, f"Variant {variant_id} not found")
    estimate = _make_estimate(variant)
    return _make_cpr(variant, estimate)

@app.get("/api/design/variants/{variant_id}/bids", response_model=list[ContractorBid])
async def design_bids(variant_id: str):
    """Аукцион подрядчиков для выбранного варианта дизайна."""
    variant = next((v for v in DESIGN_VARIANTS if v["id"] == variant_id), None)
    if not variant: raise HTTPException(404, f"Variant {variant_id} not found")
    estimate = _make_estimate(variant)
    return _make_bids(variant, estimate)

# ── Интерьеры (из dataset.json) ────────────────────────────────────────────
@app.get("/api/interiors")
async def interiors(category: str | None = None, room: str | None = None, style: str | None = None):
    """Интерьеры из dataset.json (28 items: 12 ideas, 14 rooms, 2 apts)."""
    items = INTERIORS
    if category:
        items = [i for i in items if i.get("category") == category]
    if room:
        items = [i for i in items if i.get("room") == room]
    if style:
        items = [i for i in items if i.get("style") == style]
    return items

@app.get("/api/interiors/categories")
async def interiors_categories():
    """Категории интерьеров из dataset.json."""
    return CATEGORIES

@app.get("/api/interiors/descriptions")
async def interiors_descriptions():
    """Описания категорий интерьеров из dataset.json."""
    return DESCRIPTIONS

@app.get("/api/interiors/{interior_id}")
async def interior_detail(interior_id: str):
    """Детали интерьера по ID."""
    for i in INTERIORS:
        if i.get("id") == interior_id:
            return i
    raise HTTPException(404, f"Interior {interior_id} not found")

# ── Каталог (из catalog/data.json) ─────────────────────────────────────────

@app.get("/api/catalog/products")
async def catalog_products(group: str | None = None):
    """Товары каталога из data.json (12 products)."""
    products = PRODUCTS
    if group:
        products = [p for p in products if p.get("group") == group or p.get("shape") == group]
    return products

@app.get("/api/catalog/rooms")
async def catalog_rooms():
    """Комнаты для комплектации (14 rooms)."""
    return ROOMS

@app.get("/api/catalog/rooms/{room_id}")
async def catalog_room_detail(room_id: str):
    """Детали комнаты."""
    for r in ROOMS:
        if r.get("id") == room_id:
            return r
    raise HTTPException(404, f"Room {room_id} not found")

# ── Центр (center) ─────────────────────────────────────────────────────────

@app.get("/api/center/interiors")
async def center_interiors():
    """Интерьеры центра (28 items)."""
    return CENTER_INTERIORS

@app.get("/api/center/photos")
async def center_photos():
    """Фото центра (7 items)."""
    return CENTER_PHOTOS

# ── Hotspot data ───────────────────────────────────────────────────────────

@app.get("/api/hotspot/data")
async def hotspot_data():
    """Hotspot данные для подборки товаров (10 items)."""
    return HOTSPOT_DATA

@app.get("/api/hotspot/analogues")
async def hotspot_analogues():
    """Аналоги товаров (2 items)."""
    return EXTRA_ANALOGUES

# ── Главная страница (sections) ────────────────────────────────────────────

@app.get("/api/main-page/start")
async def main_page_start():
    """Секция «С чего начнём?»"""
    f = MAIN_PAGE_DIR / "start.html"
    return f.read_text(encoding="utf-8") if f.exists() else ""

@app.get("/api/main-page/how")
async def main_page_how():
    """Секция «Как это работает»."""
    f = MAIN_PAGE_DIR / "how.html"
    return f.read_text(encoding="utf-8") if f.exists() else ""

@app.get("/api/main-page/ideas")
async def main_page_ideas():
    """Секция «Идеи интерьеров»."""
    f = MAIN_PAGE_DIR / "ideas.html"
    return f.read_text(encoding="utf-8") if f.exists() else ""

@app.get("/api/main-page/connected")
async def main_page_connected():
    """Секция «Подборка точно к интерьеру»."""
    f = MAIN_PAGE_DIR / "connected.html"
    return f.read_text(encoding="utf-8") if f.exists() else ""

# ── Изображения ────────────────────────────────────────────────────────────

# ── Совместимость со старыми фронтендами (references/estimate/cpr/cart) ────

@app.get("/api/references")
async def refs(category: str | None = None):
    r = REFERENCES
    if category: r = [x for x in r if x["category"] == category]
    return r

@app.get("/api/references/{ref_id}")
async def ref_detail(ref_id: str):
    for r in REFERENCES:
        if r["id"] == ref_id: return r
    raise HTTPException(404, "Not found")

@app.get("/api/techcards")
async def techcards(surface: str | None = None):
    cards = TECH_CARDS
    if surface: cards = [c for c in cards if c["surface"] == surface]
    return cards

def _est_dto(d: dict) -> dict:
    return {"id": d["id"], "status": d["status"], "total_rub": d["total_rub"],
            "range_min": d["range_min"], "range_max": d["range_max"], "days": d["days"],
            "lines": d["lines"], "assumptions": d["assumptions"]}

@app.get("/api/estimate/{est_id}")
async def estimate(est_id: str):
    if est_id not in ESTIMATES: raise HTTPException(404, "Not found")
    return _est_dto(ESTIMATES[est_id])

@app.post("/api/estimate/calculate")
async def calc_estimate(payload: dict | None = None):
    """Расчёт по выбранному референсу.

    payload.reference_id — id референса из REFERENCES; без него возвращается
    демонстрационная смета est_001 (как раньше). Итог, диапазон и срок
    детерминированно считаются от price_from/price_to референса, поэтому смена
    референса меняет результат расчёта.
    """
    ref_id = (payload or {}).get("reference_id") or (payload or {}).get("ref")
    if not ref_id:
        return _est_dto(ESTIMATES["est_001"])
    ref = next((r for r in REFERENCES if r["id"] == ref_id), None)
    if not ref:
        raise HTTPException(404, f"Reference {ref_id} not found")
    lo, hi = int(ref["price_from"]), int(ref["price_to"])
    mid = round((lo + hi) / 2 / 500) * 500  # «результат» — середина диапазона, округлённая до 500
    days = 30 + max(0, (mid - 180000) // 20000)  # чем дороже комплектация — тем дольше
    lines = [
        {"id": "l1", "title": ref["title"], "qty": 1, "unit": "компл.",
         "price": mid, "sku": f"REF-{ref_id.upper()}", "brand": "ZETTA"},
        {"id": "l2", "title": "Отделочные материалы", "qty": 12.5, "unit": "м²",
         "price": 3200, "sku": None, "brand": None},
        {"id": "l3", "title": "Работы", "qty": 65, "unit": "м²",
         "price": 850, "sku": None, "brand": None},
    ]
    total = sum(int(l["price"] * l["qty"]) for l in lines)
    return {"id": f"est_{ref_id}", "status": "preliminary", "total_rub": total,
            "range_min": lo, "range_max": hi, "days": days,
            "lines": lines,
            "assumptions": ["Состав уточняется после замера",
                            "Цены — из каталога на дату расчёта"]}

@app.get("/api/cpr/{cpr_id}")
async def cpr(cpr_id: str):
    est = _est_dto(ESTIMATES["est_001"])
    return {"id": cpr_id, "version": 1, "project_id": "proj_001", "estimate": est,
            "documents": [{"id": "d1", "title": "План", "type": "drawing", "status": "ready"},
                          {"id": "d2", "title": "Ведомость", "type": "specification", "status": "ready"}],
            "status": "ready"}

@app.post("/api/cpr/generate")
async def gen_cpr(payload: dict, user: dict = Depends(verify_token)):
    est = _est_dto(ESTIMATES["est_001"])
    cpr_id = "cpr_" + hashlib.md5(str(payload).encode()).hexdigest()[:8]
    return {"id": cpr_id, "version": 1, "project_id": payload.get("project_id", "proj_001"),
            "estimate": est,
            "documents": [{"id": "d1", "title": "План", "type": "drawing", "status": "ready"},
                          {"id": "d2", "title": "Ведомость", "type": "specification", "status": "pending_review"}],
            "status": "ready"}

@app.get("/api/cart")
async def get_cart(user: dict = Depends(verify_token)):
    return {"items": _cart_state["items"], "total_rub": _cart_total()}

@app.post("/api/cart/items")
async def add_cart(item: dict, user: dict = Depends(verify_token)):
    sku = item.get("sku", "")
    for it in _cart_state["items"]:
        if it.get("sku") == sku:
            it["qty"] = int(it.get("qty", 0)) + int(item.get("qty", 1))
            break
    else:
        _cart_state["items"].append(item)
    return {"items": _cart_state["items"], "total_rub": _cart_total()}

@app.delete("/api/cart/items/{sku}")
async def remove_cart_item(sku: str, user: dict = Depends(verify_token)):
    before = len(_cart_state["items"])
    _cart_state["items"] = [it for it in _cart_state["items"] if it.get("sku") != sku]
    if len(_cart_state["items"]) == before:
        raise HTTPException(404, f"Item {sku} not found in cart")
    return {"items": _cart_state["items"], "total_rub": _cart_total()}

@app.get("/api/bids")
async def bids():
    return BIDS

@app.get("/api/planners")
async def planners():
    return PLANNERS

# ── Запись в Интерьерный центр ─────────────────────────────────────────────

@app.get("/api/center/appointment/options", tags=["appointment"])
async def appointment_options():
    """Опции формы записи — дословно из оригинального блока #appointment."""
    return {
        "tasks": TASK_OPTIONS,
        "times": TIME_OPTIONS,
        "max_ideas": MAX_PICKED_IDEAS,
        "max_files": MAX_FILES,
        "place": {
            "address": "Москва, Аминьевское шоссе, 6",
            "mall": "ТЦ «Квартал», 2 этаж · бесплатная парковка",
            "phone": "+7 (495) 138-73-38",
            "phone_href": "tel:+74951387338",
            "map": "https://yandex.ru/maps/?text=%D0%9C%D0%BE%D1%81%D0%BA%D0%B2%D0%B0%2C%20"
                   "%D0%90%D0%BC%D0%B8%D0%BD%D1%8C%D0%B5%D0%B2%D1%81%D0%BA%D0%BE%D0%B5%20"
                   "%D1%88%D0%BE%D1%81%D1%81%D0%B5%2C%206%2C%20Zetta%20Home",
            "note": "Удобное время встречи подтвердит менеджер. Планируйте около 2–3 часов.",
        },
        "consent_url": "https://zettahome.ru/privacy-policy/consent/",
        "policy_url": "https://zettahome.ru/privacy-policy/",
        "dry_run": "/api/center/appointment/dry-run",
        "create": "/api/center/appointment",
    }


@app.post("/api/center/appointment/dry-run", response_model=AppointmentValidation, tags=["appointment"])
async def appointment_dry_run(req: AppointmentRequest):
    """Проверка формы без записи. Ничего не сохраняется."""
    return _validate_appointment(req)


@app.post("/api/center/appointment", response_model=AppointmentResponse, tags=["appointment"])
async def create_appointment(
    req: AppointmentRequest,
    approved: bool = False,
    idempotency_key: str | None = Header(default=None, alias="Idempotency-Key"),
):
    """Создать запись. Только с Idempotency-Key и подтверждением (approval-gate)."""
    if not idempotency_key:
        raise HTTPException(400, "Требуется заголовок Idempotency-Key")

    replay_id = IDEMPOTENCY_INDEX.get(idempotency_key)
    if replay_id:
        stored = APPOINTMENTS[replay_id]
        return AppointmentResponse(
            id=replay_id,
            status=stored["status"],
            created_at=stored["created_at"],
            idempotent_replay=True,
            appointment=stored,
        )

    check = _validate_appointment(req)
    if not check.ok:
        raise HTTPException(422, {"errors": check.errors, "warnings": check.warnings})

    # approval-gate: без явного подтверждения запись не создаётся
    if not approved:
        raise HTTPException(
            428,
            {
                "detail": "Запись требует подтверждения менеджером (approval-gate)",
                "dry_run": check.model_dump(),
            },
        )

    appointment_id = "appt_" + uuid.uuid4().hex[:10]
    stored = {
        "id": appointment_id,
        "status": "pending_confirmation",
        **_appointment_payload(req),
        "validation": check.model_dump(),
    }
    stored["created_at"] = datetime.now(timezone.utc).isoformat()
    APPOINTMENTS[appointment_id] = stored
    IDEMPOTENCY_INDEX[idempotency_key] = appointment_id
    return AppointmentResponse(
        id=appointment_id,
        status=stored["status"],
        created_at=stored["created_at"],
        appointment=stored,
    )


@app.get("/api/center/appointments/{appointment_id}", tags=["appointment"])
async def get_appointment(appointment_id: str):
    stored = APPOINTMENTS.get(appointment_id)
    if not stored:
        raise HTTPException(404, f"Запись {appointment_id} не найдена")
    return stored


# ── CSS и разметка приложений прототипа А (center, repair, catalog, about) ──

APP_FILES = {
    "main": (None, "main.css"),
    "interiors": ("interiors.html", "interiors.css"),
    "catalog": ("catalog.html", "catalog.css"),
    "center": ("center.html", "center.css"),
    "repair": ("repair.html", "repair.css"),
    "about": ("about.html", "about.css"),
}


@app.get("/api/apps", tags=["apps"])
async def apps_index():
    """Список приложений прототипа А и их файлов."""
    return {
        name: {
            "html": f"/api/apps/{name}/html" if html else None,
            "css": f"/api/apps/{name}/css",
            "files": {"html": html, "css": css},
        }
        for name, (html, css) in APP_FILES.items()
    }


@app.get("/api/apps/{app_name}/css", tags=["apps"], response_class=FileResponse)
async def app_css(app_name: str):
    """Оригинальный CSS приложения (селекторы под #surface52)."""
    entry = APP_FILES.get(app_name)
    if not entry:
        raise HTTPException(404, f"Приложение '{app_name}' не известно")
    path = APPS_DIR / entry[1]
    if not path.exists():
        raise HTTPException(404, f"CSS {entry[1]} не найден")
    return FileResponse(
        path,
        media_type="text/css; charset=utf-8",
        headers={"Cache-Control": "public, max-age=3600"},
    )


@app.get("/api/apps/{app_name}/html", tags=["apps"], response_class=FileResponse)
async def app_html(app_name: str):
    """Оригинальная разметка приложения (картинки уже через /api/images/file/...)."""
    entry = APP_FILES.get(app_name)
    if not entry:
        raise HTTPException(404, f"Приложение '{app_name}' не известно")
    if not entry[0]:
        raise HTTPException(404, f"У приложения '{app_name}' есть только CSS")
    path = APPS_DIR / entry[0]
    if not path.exists():
        raise HTTPException(404, f"Разметка {entry[0]} не найдена")
    return FileResponse(
        path,
        media_type="text/html; charset=utf-8",
        headers={"Cache-Control": "public, max-age=3600"},
    )

# ── Разметка оригинальных прототипов (для 1:1 переноса во Vue) ─────────────
#
# Файги HTML извлечены из «Прототип А.html» / «Прототип В.html»; все
# data:image в них уже заменены на /api/images/file/<имя>.jpg.
# Фронтенд забирает кусок разметки и вставляет его в шаблон, поэтому
# вёрстка совпадает с оригиналом, а данные (картинки) идут через API.

A_PAGE_DIR = DATA_DIR / "a_page"
B_PAGE_DIR = DATA_DIR / "b_page"

PAGE_DIRS = {"a": A_PAGE_DIR, "b": B_PAGE_DIR}


@app.get("/api/page", tags=["page"])
async def page_index():
    """Список доступных HTML-фрагментов разметки."""
    def listing(d: Path):
        if not d.exists():
            return []
        return sorted(f.name for f in d.glob("*.html"))
    return {
        "a": {"dir": str(A_PAGE_DIR), "files": listing(A_PAGE_DIR)},
        "b": {"dir": str(B_PAGE_DIR), "files": listing(B_PAGE_DIR)},
        "served_as": "/api/page/{proto}/{file}",
    }


@app.get("/api/page/{proto}/{fname}", tags=["page"], response_class=FileResponse)
async def page_fragment(proto: str, fname: str):
    """Отдать фрагмент оригинальной разметки (html|css)."""
    base = PAGE_DIRS.get(proto)
    if not base:
        raise HTTPException(404, f"Прототип '{proto}' не известен (a|b)")
    # защита от path traversal
    if "/" in fname or "\\" in fname or ".." in fname:
        raise HTTPException(400, "Некорректное имя файла")
    path = base / fname
    if not path.exists():
        raise HTTPException(404, f"Фрагмент {proto}/{fname} не найден")
    media = "text/css" if fname.endswith(".css") else "text/html"
    return FileResponse(path, media_type=media + "; charset=utf-8",
                        headers={"Cache-Control": "public, max-age=3600"})


@app.get("/api/css/{proto}", tags=["page"])
async def page_css(proto: str):
    """Оригинальный CSS прототипа одним файлом (подключается на фронте)."""
    fname = {"a": "a_all.css", "b": "b_all.css"}.get(proto)
    if not fname:
        raise HTTPException(404, "proto должен быть a или b")
    path = DATA_DIR / "css" / fname
    if not path.exists():
        raise HTTPException(404, f"CSS {fname} не найден")
    return FileResponse(path, media_type="text/css; charset=utf-8")


@app.get("/api/css/{proto}/shell", tags=["page"])
async def page_shell_css(proto: str):
    """Постоянная оболочка прототипа.

    В оригинале таблица стилей приложения подменялась целиком
    (`style52.textContent = info.css`), а оболочка оставалась на месте:
    первый `<style>` и `<style id="about-shell68">`. Отдаём только её,
    чтобы CSS приложения не тянул за собой стили главной.
    """
    fname = {"a": "a_shell.css", "b": "b_shell.css"}.get(proto)
    if not fname:
        raise HTTPException(404, "proto должен быть a или b")
    path = DATA_DIR / "css" / fname
    if not path.exists():
        raise HTTPException(404, f"CSS оболочки {fname} не найден")
    return FileResponse(path, media_type="text/css; charset=utf-8")


@app.get("/api/images", tags=["images"])
async def image_index():
    return {
        "served_as": "/api/images/{key}",
        "storage": "БД / PIM (контент)",
        "keys": IMAGE_KEYS,
    }

@app.get("/api/images/{key}", tags=["images"], response_class=FileResponse)
async def image_by_key(key: str):
    entry = IMAGE_KEYS.get(key)
    if not entry:
        raise HTTPException(404, f"Изображение '{key}' не найдено")
    path = IMAGES_DIR / entry["file"]
    if not path.exists():
        raise HTTPException(404, f"Файл {entry['file']} отсутствует в хранилище")
    return FileResponse(path, headers={"Cache-Control": "public, max-age=86400"})

@app.get("/api/images/file/{filename}", tags=["images"], response_class=FileResponse)
async def image_by_filename(filename: str):
    """Отдать файл по имени из data:image (a_interiors_img_0001.jpg и т.д.)."""
    path = IMAGES_DIR / filename
    if not path.exists():
        raise HTTPException(404, f"Файл {filename} отсутствует в хранилище")
    return FileResponse(path, headers={"Cache-Control": "public, max-age=86400"})

@app.get("/health")
@app.get("/api/health")
async def health():
    return {"status": "ok", "mock": True, "images": len(IMAGE_KEYS),
            "interiors": len(INTERIORS), "products": len(PRODUCTS),
            "design_variants": len(DESIGN_VARIANTS), "main_sections": len(MAIN_PAGE_SECTIONS),
            "apps": len(APP_FILES), "appointments": len(APPOINTMENTS)}
