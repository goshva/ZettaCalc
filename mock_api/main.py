"""
Mock REST API для Fresh House / ZettaCalc
──────────────────────────────────────────
FastAPI с Swagger UI, JWT-аутентификацией (mock), моделями из прототипов А и В.

Запуск:
    cd mock_api && python main.py
Открыть: http://127.0.0.1:8000/docs
"""

from __future__ import annotations

import hashlib
import time
from datetime import datetime, timedelta, timezone
from typing import Any

from fastapi import Depends, FastAPI, HTTPException, status
from fastapi.security import OAuth2PasswordBearer
from jose import JWTError, jwt
from pydantic import BaseModel, EmailStr, Field

# ── Конфиг ──────────────────────────────────────────────────────────────────

SECRET_KEY = "dev-mock-key-not-for-production"
ALGORITHM = "HS256"
ACCESS_TOKEN_EXPIRE = timedelta(days=1)

app = FastAPI(
    title="Fresh House Mock API",
    description="Mock REST API для прототипов Zetta Home (A и B) с JWT и Swagger",
    version="0.1.0",
)

oauth2_scheme = OAuth2PasswordBearer(tokenUrl="auth/login")


# ── JWT ─────────────────────────────────────────────────────────────────────

def create_token(user_id: str, email: str) -> str:
    payload = {
        "sub": user_id,
        "email": email,
        "exp": datetime.now(timezone.utc) + ACCESS_TOKEN_EXPIRE,
        "iat": datetime.now(timezone.utc),
    }
    return jwt.encode(payload, SECRET_KEY, algorithm=ALGORITHM)


def verify_token(token: str = Depends(oauth2_scheme)) -> dict[str, Any]:
    try:
        payload = jwt.decode(token, SECRET_KEY, algorithms=[ALGORITHM])
        return payload
    except JWTError:
        raise HTTPException(status_code=401, detail="Неверный токен")


# ── Модели ──────────────────────────────────────────────────────────────────

class LoginRequest(BaseModel):
    email: EmailStr
    password: str


class TokenResponse(BaseModel):
    access_token: str
    token_type: str = "bearer"
    user_id: str
    email: str


class BriefField(BaseModel):
    id: str
    label: str
    value: str | None = None
    type: str = "select"  # select, text, number, boolean
    options: list[str] | None = None


class BriefResponse(BaseModel):
    id: str
    status: str  # draft, complete
    fields: list[BriefField]
    created_at: str
    updated_at: str


class ReferenceItem(BaseModel):
    id: str
    title: str
    category: str  # kitchen, living, bathroom, bedroom
    image: str
    style_tags: list[str]
    price_from: int
    price_to: int
    sku_count: int


class TechCard(BaseModel):
    id: str
    code: str
    title: str
    surface: str
    material: str
    operations: list[str]
    materials_consumption: dict[str, float]
    min_area_m2: float
    max_unevenness_mm_per_m: float | None = None
    version: int


class EstimateLine(BaseModel):
    id: str
    title: str
    quantity: float
    unit: str
    price_per_unit: float
    total: float
    sku: str | None = None
    brand: str | None = None


class EstimateResponse(BaseModel):
    id: str
    status: str  # preliminary, basic_cpr, cpr_with_options
    total_rub: int
    range_min_rub: int
    range_max_rub: int
    estimated_days: int
    lines: list[EstimateLine]
    assumptions: list[str]
    created_at: str


class CPRDocument(BaseModel):
    id: str
    title: str
    type: str  # drawing, specification, schedule
    status: str  # draft, ready, pending_review
    url: str | None = None


class CPRResponse(BaseModel):
    id: str
    version: int
    project_id: str
    estimate: EstimateResponse
    documents: list[CPRDocument]
    status: str
    created_at: str


class CartItem(BaseModel):
    sku: str
    brand: str
    title: str
    price: int
    quantity: int
    image: str


class CartResponse(BaseModel):
    items: list[CartItem]
    total_rub: int


class OrderItem(BaseModel):
    sku: str
    title: str
    quantity: int
    price: int


class OrderRequest(BaseModel):
    project_id: str
    items: list[OrderItem]
    contractor_id: str | None = None


class OrderResponse(BaseModel):
    id: str
    status: str  # pending, confirmed, shipped
    total_rub: int
    items: list[OrderItem]
    created_at: str


class ContractorBid(BaseModel):
    id: str
    contractor_name: str
    rating: float
    price_rub: int
    estimated_days: int
    scope: str


class Planner(BaseModel):
    id: str
    name: str
    area_m2: float
    rooms: list[dict[str, Any]]
    image: str


# ── Mock данные (из прототипов А и В) ───────────────────────────────────────

USERS = {
    "test@zetta.ru": {"id": "user_001", "password": "test123", "email": "test@zetta.ru"},
}

REFERENCES = [
    ReferenceItem(
        id="idea_kitchen_001",
        title="Современная кухня с островом",
        category="kitchen",
        image="https://images.unsplash.com/photo-1556909114-f6e7ad7d3136?w=800&h=600&fit=crop",
        style_tags=["modern", "minimalist", "scandinavian"],
        price_from=450000,
        price_to=1200000,
        sku_count=87,
    ),
    ReferenceItem(
        id="idea_living_001",
        title="Гостиная в тёплых тонах",
        category="living",
        image="https://images.unsplash.com/photo-1618221195710-dd6b41faaea6?w=800&h=600&fit=crop",
        style_tags=["modern", "warm", "cozy"],
        price_from=320000,
        price_to=890000,
        sku_count=64,
    ),
    ReferenceItem(
        id="idea_bathroom_001",
        title="Санузел в минималистичном стиле",
        category="bathroom",
        image="https://images.unsplash.com/photo-1552321554-5fefe8c9ef14?w=800&h=600&fit=crop",
        style_tags=["minimalist", "modern", "light"],
        price_from=180000,
        price_to=520000,
        sku_count=42,
    ),
    ReferenceItem(
        id="idea_kitchen_002",
        title="Классическая кухня с фрезеровкой",
        category="kitchen",
        image="https://images.unsplash.com/photo-1556909114-44e3e70034e2?w=800&h=600&fit=crop",
        style_tags=["classical", "traditional", "warm"],
        price_from=680000,
        price_to=2100000,
        sku_count=124,
    ),
    ReferenceItem(
        id="idea_living_002",
        title="Скандинавская гостиная",
        category="living",
        image="https://images.unsplash.com/photo-1586023492125-27b2c045efd7?w=800&h=600&fit=crop",
        style_tags=["scandinavian", "light", "minimalist"],
        price_from=280000,
        price_to=750000,
        sku_count=56,
    ),
    ReferenceItem(
        id="idea_bedroom_001",
        title="Спальня с мягким изголовьем",
        category="bedroom",
        image="https://images.unsplash.com/photo-1616594039964-ae9021a400a0?w=800&h=600&fit=crop",
        style_tags=["modern", "cozy", "soft"],
        price_from=250000,
        price_to=680000,
        sku_count=48,
    ),
]

TECH_CARDS = [
    TechCard(
        id="tc_001",
        code="FLOOR_KER_600",
        title="Укладка керамической плитки 600×600",
        surface="floor",
        material="ceramic_tile",
        operations=[
            "Демонтаж старого покрытия",
            "Выравнивание основания",
            "Гидроизоляция",
            "Приклеивание плитки",
            "Затирка швов",
        ],
        materials_consumption={"klej_m2": 5.0, "grout_m2": 0.3, "hydro_layer_m2": 1.05},
        min_area_m2=3.0,
        max_unevenness_mm_per_m=3.0,
        version=3,
    ),
    TechCard(
        id="tc_002",
        code="WALL_PLASTER_GYP",
        title="Штукатурка гипсовая стены",
        surface="wall",
        material="gypsum_plaster",
        operations=[
            "Грунтовка основания",
            "Установка маяков",
            "Нанесение штукатурки",
            "Накрывка и затирка",
        ],
        materials_consumption={"primer_l_m2": 0.15, "plaster_kg_m2": 10.0},
        min_area_m2=5.0,
        max_unevenness_mm_per_m=None,
        version=2,
    ),
    TechCard(
        id="tc_003",
        code="WALL_PAINT_LATEX",
        title="Покраска латексной краской",
        surface="wall",
        material="latex_paint",
        operations=[
            "Шпатлёвка",
            "Шлифовка",
            "Грунтовка",
            "Покраска (2 слоя)",
        ],
        materials_consumption={"primer_l_m2": 0.1, "paint_l_m2": 0.2},
        min_area_m2=2.0,
        max_unevenness_mm_per_m=2.0,
        version=1,
    ),
]

ESTIMATES: dict[str, EstimateResponse] = {
    "est_prelim_001": EstimateResponse(
        id="est_prelim_001",
        status="preliminary",
        total_rub=847500,
        range_min_rub=720000,
        range_max_rub=980000,
        estimated_days=45,
        lines=[
            EstimateLine(id="l1", title="Кухонный гарнитур ZETTA METALICO", quantity=1, unit="комплект",
                         price_per_unit=385000, total=385000, sku="ZET-KIT-MET-001", brand="ZETTA"),
            EstimateLine(id="l2", title="Укладка плитки 600×600", quantity=12.5, unit="м²",
                         price_per_unit=3200, total=40000, sku=None, brand=None),
            EstimateLine(id="l3", title="Штукатурка стен", quantity=65.0, unit="м²",
                         price_per_unit=850, total=55250, sku=None, brand=None),
            EstimateLine(id="l4", title="Покраска стен (2 слоя)", quantity=65.0, unit="м²",
                         price_per_unit=620, total=40300, sku="LP-TEX-001", brand="LUXENSE"),
            EstimateLine(id="l5", title="Светотехника", quantity=8, unit="точек",
                         price_per_unit=4500, total=36000, sku="LGT-SP-008", brand="EGLO"),
            EstimateLine(id="l6", title="Сантехника", quantity=1, unit="комплект",
                         price_per_unit=120000, total=120000, sku="SNT-BTH-001", brand="Villeroy&Boch"),
            EstimateLine(id="l7", title="Доставка и подъём", quantity=1, unit="услуга",
                         price_per_unit=15000, total=15000, sku=None, brand=None),
        ],
        assumptions=[
            "Ровность пола требует проверки на объекте",
            "Состояние стен неизвестно — возможен доп. расход штукатурки",
            "Электропроводка — предположение: замена не требуется",
        ],
        created_at="2026-11-04T10:00:00Z",
    ),
}

CPRS: dict[str, CPRResponse] = {
    "cpr_001": CPRResponse(
        id="cpr_001",
        version=1,
        project_id="proj_001",
        estimate=ESTIMATES["est_prelim_001"],
        documents=[
            CPRDocument(id="doc_1", title="План помещений", type="drawing", status="ready", url="/docs/cpr_001_plan.pdf"),
            CPRDocument(id="doc_2", title="Развёртка стен", type="drawing", status="ready", url="/docs/cpr_001_walls.pdf"),
            CPRDocument(id="doc_3", title="Ведомость материалов", type="specification", status="ready", url="/docs/cpr_001_materials.pdf"),
            CPRDocument(id="doc_4", title="Календарный план", type="schedule", status="pending_review", url=None),
        ],
        status="ready",
        created_at="2026-11-04T11:00:00Z",
    ),
}

CARTS: dict[str, CartResponse] = {
    "default": CartResponse(
        items=[
            CartItem(sku="ZET-KIT-MET-001", brand="ZETTA", title="Кухонный гарнитур METALICO",
                     price=385000, quantity=1, image="img_kitchen.jpg"),
            CartItem(sku="LGT-SP-008", brand="EGLO", title="Точечный светильник LED x8",
                     price=36000, quantity=1, image="img_light.jpg"),
        ],
        total_rub=421000,
    ),
}

PLANNERS = [
    Planner(
        id="planner_001",
        name="Квартира 90 м² — стандартная",
        area_m2=90.0,
        rooms=[
            {"name": "Кухня-гостиная", "area_m2": 28.0},
            {"name": "Спальня", "area_m2": 18.0},
            {"name": "Детская", "area_m2": 14.0},
            {"name": "Санузел", "area_m2": 6.0},
            {"name": "Прихожая", "area_m2": 10.0},
        ],
        image="planner_001.jpg",
    ),
    Planner(
        id="planner_002",
        name="Квартира 90 м² — с перепланировкой",
        area_m2=92.0,
        rooms=[
            {"name": "Кухня-гостиная", "area_m2": 35.0},
            {"name": "Спальня", "area_m2": 18.0},
            {"name": "Кабинет", "area_m2": 12.0},
            {"name": "Санузел совмещённый", "area_m2": 8.0},
            {"name": "Прихожая", "area_m2": 9.0},
        ],
        image="planner_002.jpg",
    ),
    Planner(
        id="planner_003",
        name="Квартира 60 м² — студия",
        area_m2=60.0,
        rooms=[
            {"name": "Кухня-гостиная", "area_m2": 30.0},
            {"name": "Спальня", "area_m2": 16.0},
            {"name": "Санузел", "area_m2": 5.0},
            {"name": "Прихожая", "area_m2": 9.0},
        ],
        image="planner_003.jpg",
    ),
]

BIDS = [
    ContractorBid(id="bid_001", contractor_name="РемонтПро", rating=4.8,
                    price_rub=320000, estimated_days=40, scope="Полный ремонт кухни и санузла"),
    ContractorBid(id="bid_002", contractor_name="СтройМастер", rating=4.5,
                    price_rub=280000, estimated_days=45, scope="Полный ремонт кухни и санузла"),
    ContractorBid(id="bid_003", contractor_name="ДизайнРемонт", rating=4.9,
                    price_rub=410000, estimated_days=35, scope="Полный ремонт кухни и санузла + дизайн-сопровождение"),
]

# ── Эндпоинты ───────────────────────────────────────────────────────────────

@app.post("/auth/login", response_model=TokenResponse, summary="Вход (JWT)")
async def login(req: LoginRequest):
    user = USERS.get(req.email)
    if not user or user["password"] != req.password:
        raise HTTPException(status_code=401, detail="Неверный email или пароль")
    token = create_token(user["id"], user["email"])
    return TokenResponse(access_token=token, user_id=user["id"], email=user["email"])


@app.get("/auth/me", summary="Текущий пользователь")
async def get_me(payload: dict = Depends(verify_token)):
    return {"user_id": payload["sub"], "email": payload["email"]}


@app.get("/references", response_model=list[ReferenceItem], summary="Референсы интерьеров")
async def list_references(category: str | None = None, style: str | None = None):
    refs = REFERENCES
    if category:
        refs = [r for r in refs if r.category == category]
    if style:
        refs = [r for r in refs if style in r.style_tags]
    return refs


@app.get("/references/{ref_id}", response_model=ReferenceItem, summary="Референс по ID")
async def get_reference(ref_id: str):
    for r in REFERENCES:
        if r.id == ref_id:
            return r
    raise HTTPException(status_code=404, detail="Референс не найден")


@app.get("/brief", response_model=BriefResponse, summary="Бриф клиента")
async def get_brief(payload: dict = Depends(verify_token)):
    return BriefResponse(
        id="brief_001",
        status="draft",
        fields=[
            BriefField(id="area", label="Площадь квартиры, м²", value="90", type="number"),
            BriefField(id="rooms", label="Количество комнат", value="3", type="select", options=["1","2","3","4","5+"]),
            BriefField(id="family", label="Состав семьи", value="Семья с детьми", type="select", options=["Один","Пара","Семья с детьми","Поколения"]),
            BriefField(id="timeline", label="Желаемые сроки ремонта", value="3-6 месяцев", type="select", options=["1-2 мес","3-6 мес","6-12 мес","Без спешки"]),
            BriefField(id="style", label="Предпочтительный стиль", value="Современный", type="select", options=["Современный","Скандинавский","Классика","Минимализм"]),
            BriefField(id="has_project", label="Есть ли дизайн-проект", value="false", type="boolean"),
        ],
        created_at="2026-11-04T09:00:00Z",
        updated_at="2026-11-04T09:30:00Z",
    )


@app.put("/brief", response_model=BriefResponse, summary="Обновить бриф")
async def update_brief(
    fields: dict[str, str],
    payload: dict = Depends(verify_token),
):
    # Mock: просто возвращаем обновлённый бриф
    return BriefResponse(
        id="brief_001",
        status="complete" if len(fields) >= 3 else "draft",
        fields=[
            BriefField(id=f, label="Поле " + f, value=v, type="text")
            for f, v in fields.items()
        ],
        created_at="2026-11-04T09:00:00Z",
        updated_at=datetime.now(timezone.utc).isoformat(),
    )


@app.get("/planners", response_model=list[Planner], summary="Библиотека планировок")
async def list_planners():
    return PLANNERS


@app.get("/techcards", response_model=list[TechCard], summary="Библиотека техкарт")
async def list_techcards(surface: str | None = None, material: str | None = None):
    cards = TECH_CARDS
    if surface:
        cards = [c for c in cards if c.surface == surface]
    if material:
        cards = [c for c in cards if c.material == material]
    return cards


@app.get("/estimate/{est_id}", response_model=EstimateResponse, summary="Предварительная калькуляция")
async def get_estimate(est_id: str):
    if est_id not in ESTIMATES:
        raise HTTPException(status_code=404, detail="Расчёт не найден")
    return ESTIMATES[est_id]


@app.post("/estimate/calculate", response_model=EstimateResponse, summary="Рассчитать (mock)")
async def calculate_estimate(payload: dict):
    """Mock: возвращает предзаготовленный расчёт."""
    return ESTIMATES["est_prelim_001"]


@app.get("/cpr/{cpr_id}", response_model=CPRResponse, summary="Цифровой паспорт ремонта")
async def get_cpr(cpr_id: str):
    if cpr_id not in CPRS:
        raise HTTPException(status_code=404, detail="ЦПР не найден")
    return CPRS[cpr_id]


@app.post("/cpr/generate", response_model=CPRResponse, summary="Сгенерировать ЦПР (mock)")
async def generate_cpr(
    payload: dict,
    payload_user: dict = Depends(verify_token),
):
    """Mock: генерирует базовый ЦПР."""
    cpr = CPRResponse(
        id="cpr_new_" + hashlib.md5(str(payload).encode()).hexdigest()[:8],
        version=1,
        project_id=payload.get("project_id", "proj_001"),
        estimate=ESTIMATES["est_prelim_001"],
        documents=[
            CPRDocument(id="d1", title="План помещений", type="drawing", status="ready", url="/docs/plan.pdf"),
            CPRDocument(id="d2", title="Ведомость материалов", type="specification", status="ready", url="/docs/spec.pdf"),
        ],
        status="ready",
        created_at=datetime.now(timezone.utc).isoformat(),
    )
    CPRS[cpr.id] = cpr
    return cpr


@app.get("/cart", response_model=CartResponse, summary="Корзина")
async def get_cart(payload: dict = Depends(verify_token)):
    return CARTS["default"]


@app.post("/cart/items", response_model=CartResponse, summary="Добавить в корзину")
async def add_to_cart(item: dict):
    cart = CARTS["default"]
    cart.items.append(CartItem(**item))
    cart.total_rub += item["price"] * item.get("quantity", 1)
    return cart


@app.delete("/cart/items/{sku}", summary="Удалить из корзины")
async def remove_from_cart(sku: str, payload: dict = Depends(verify_token)):
    cart = CARTS["default"]
    cart.items = [i for i in cart.items if i.sku != sku]
    cart.total_rub = sum(i.price * i.quantity for i in cart.items)
    return cart


@app.get("/bids", response_model=list[ContractorBid], summary="Предложения подрядчиков")
async def list_bids():
    return BIDS


@app.post("/orders", response_model=OrderResponse, summary="Оформить заказ")
async def create_order(
    req: OrderRequest,
    payload: dict = Depends(verify_token),
):
    order = OrderResponse(
        id="ord_" + hashlib.md5(str(req).encode()).hexdigest()[:8],
        status="pending",
        total_rub=sum(i.price * i.quantity for i in req.items),
        items=req.items,
        created_at=datetime.now(timezone.utc).isoformat(),
    )
    return order


@app.get("/health", summary="Здоровье API")
async def health():
    return {"status": "ok", "mock": True}
