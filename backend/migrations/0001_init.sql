-- ZettaCalc / Fresh House — схема БД Go-сервиса (PostgreSQL 18).
--
-- Границы: сюда переносится контент mock API (vue-prototypes/api) — сущности
-- прототипов A и B и картинки, привязанные к этим сущностям. Общие статические
-- картинки сайта (hero, video, app, cpr, expert, auction, ceo, plan, repair,
-- center-*, ds-*) остаются в public/ фронтендов и в БД не дублируются:
-- для них image_key.storage = 'client', а bytes не хранятся.
--
-- Паритет с Python-mock: у каждой «файловой» сущности хранится точный исходный
-- JSON (payload_text) — ответы отдаются теми же байтами, что и раньше.

BEGIN;

-- ───────────────────────────────────────────────────────────────────────────
-- Медиа: файлы картинок и ключи PIM
-- ───────────────────────────────────────────────────────────────────────────

CREATE TABLE media_asset (
    filename     text        PRIMARY KEY,
    content_type text        NOT NULL,
    byte_size    bigint      NOT NULL CHECK (byte_size >= 0),
    sha256       text        NOT NULL,
    width        int,
    height       int,
    -- entity — картинка сущности (интерьер, товар, фото центра…),
    -- static — общая картинка сайта (в БД не хранится),
    -- client — файл раздаётся фронтендом (Vite public)
    storage      text        NOT NULL CHECK (storage IN ('entity', 'static', 'client')),
    group_prefix text        NOT NULL,           -- a_interiors / a_main / b / …
    bytes        bytea,                          -- NULL для storage='client'
    created_at   timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT media_asset_bytes_present CHECK (storage <> 'entity' OR bytes IS NOT NULL)
);

COMMENT ON TABLE media_asset IS 'Файлы изображений API: картинки сущностей и служебные файлы.';

CREATE TABLE image_key (
    key        text PRIMARY KEY,
    filename   text REFERENCES media_asset (filename),
    source     text NOT NULL CHECK (source IN ('db', 'static')),
    storage    text NOT NULL CHECK (storage IN ('db', 'client')),
    sort_order int  NOT NULL DEFAULT 0
);

COMMENT ON TABLE image_key IS 'Ключи PIM (idea-kitchen, cat-furniture, hero…), как в mock API.';

-- Связь «картинка ↔ сущность»: именно она делает картинку картинкой сущности.
CREATE TABLE media_entity_ref (
    filename    text NOT NULL REFERENCES media_asset (filename) ON DELETE CASCADE,
    entity_type text NOT NULL,   -- interior | product | room | offer | extra | hotspot | analogue
                                 -- center_interior | center_photo | catalog_photo | catalog_scene
                                 -- design_variant | connected_item | reference
    entity_id   text NOT NULL,
    role        text NOT NULL,   -- primary | gallery | plan | scene | photo | image
    position    int  NOT NULL DEFAULT 0,
    PRIMARY KEY (filename, entity_type, entity_id, role, position)
);

CREATE INDEX media_entity_ref_entity_idx ON media_entity_ref (entity_type, entity_id);

-- ───────────────────────────────────────────────────────────────────────────
-- Контекст Content: интерьеры, каталог, центр, hotspot, справочники
-- ───────────────────────────────────────────────────────────────────────────

CREATE TABLE interior (
    id           text PRIMARY KEY,
    type         text NOT NULL CHECK (type IN ('idea', 'room', 'apt')),
    title        text NOT NULL,
    room         text,
    style        text,
    palette      text,
    image_file   text REFERENCES media_asset (filename),
    source_url   text,
    area_m2      numeric(10, 2),
    parent_id    text,
    sort_order   int  NOT NULL,
    payload      jsonb NOT NULL,
    payload_text text  NOT NULL
);

CREATE INDEX interior_type_idx ON interior (type, sort_order);

CREATE TABLE interior_category (
    name       text PRIMARY KEY,
    sort_order int NOT NULL
);

CREATE TABLE interior_description (
    name       text PRIMARY KEY,
    sort_order int NOT NULL
);

CREATE TABLE interior_color (
    interior_id text NOT NULL REFERENCES interior (id) ON DELETE CASCADE,
    position    int  NOT NULL,
    color       text NOT NULL,
    PRIMARY KEY (interior_id, position)
);

CREATE TABLE interior_variant (
    interior_id text NOT NULL REFERENCES interior (id) ON DELETE CASCADE,
    position    int  NOT NULL,
    name        text NOT NULL,
    PRIMARY KEY (interior_id, position)
);

-- costs: массив строк (бюджетные срезы по категориям)
CREATE TABLE interior_cost (
    interior_id text NOT NULL REFERENCES interior (id) ON DELETE CASCADE,
    row_no      int  NOT NULL,
    column_no   int  NOT NULL,
    amount      numeric(14, 2) NOT NULL,
    PRIMARY KEY (interior_id, row_no, column_no)
);

-- ───────────────────────────────────────────────────────────────────────────
-- Контекст Content: каталог (products, offers, rooms, extra, photos, scenes)
-- ───────────────────────────────────────────────────────────────────────────

CREATE TABLE catalog_product (
    id           text PRIMARY KEY,
    name         text NOT NULL,
    url          text,
    price        numeric(14, 2),
    brand        text,
    shape        text,
    image_file   text REFERENCES media_asset (filename),
    sort_order   int  NOT NULL,
    payload      jsonb NOT NULL,
    payload_text text  NOT NULL
);

CREATE INDEX catalog_product_shape_idx ON catalog_product (shape);

CREATE TABLE catalog_offer (
    id           text PRIMARY KEY,
    width        int,
    color        text,
    fabric       text,
    price        numeric(14, 2),
    sort_order   int NOT NULL,
    payload      jsonb NOT NULL,
    payload_text text  NOT NULL
);

CREATE TABLE catalog_room (
    id           text PRIMARY KEY,
    title        text NOT NULL,
    room         text,
    palette      text,
    area_m2      numeric(10, 2),
    parent_id    text,
    image_file   text REFERENCES media_asset (filename),
    sort_order   int  NOT NULL,
    payload      jsonb NOT NULL,
    payload_text text  NOT NULL
);

CREATE TABLE catalog_room_color (
    room_id  text NOT NULL REFERENCES catalog_room (id) ON DELETE CASCADE,
    position int  NOT NULL,
    color    text NOT NULL,
    PRIMARY KEY (room_id, position)
);

CREATE TABLE catalog_room_cost (
    room_id   text NOT NULL REFERENCES catalog_room (id) ON DELETE CASCADE,
    row_no    int  NOT NULL,
    column_no int  NOT NULL,
    amount    numeric(14, 2) NOT NULL,
    PRIMARY KEY (room_id, row_no, column_no)
);

CREATE TABLE catalog_extra (
    id           text PRIMARY KEY,
    name         text NOT NULL,
    short_name   text,
    price        numeric(14, 2),
    group_name   text,
    unit         text,
    url          text,
    custom       jsonb,
    rooms        jsonb,
    image_file   text REFERENCES media_asset (filename),
    sort_order   int  NOT NULL,
    payload      jsonb NOT NULL,
    payload_text text  NOT NULL
);

-- photos и scenes каталога: слот → картинка
CREATE TABLE catalog_slot_image (
    scope        text NOT NULL CHECK (scope IN ('photo', 'scene')),
    slot         text NOT NULL,
    image_file   text REFERENCES media_asset (filename),
    sort_order   int  NOT NULL,
    payload      jsonb NOT NULL,
    payload_text text  NOT NULL,
    PRIMARY KEY (scope, slot)
);

CREATE TABLE catalog_offer_image (
    offer_id  text NOT NULL REFERENCES catalog_offer (id) ON DELETE CASCADE,
    position  int  NOT NULL,
    filename  text REFERENCES media_asset (filename),
    PRIMARY KEY (offer_id, position)
);

-- ───────────────────────────────────────────────────────────────────────────
-- Контекст Content: центр, hotspot, подборки, планировщики
-- ───────────────────────────────────────────────────────────────────────────

CREATE TABLE center_interior (
    id           text PRIMARY KEY,
    title        text NOT NULL,
    image_file   text REFERENCES media_asset (filename),
    sort_order   int  NOT NULL,
    payload      jsonb NOT NULL,
    payload_text text  NOT NULL
);

CREATE TABLE center_photo (
    id           text PRIMARY KEY,
    name         text NOT NULL,
    image_file   text REFERENCES media_asset (filename),
    sort_order   int  NOT NULL,
    payload      jsonb NOT NULL,
    payload_text text  NOT NULL
);

CREATE TABLE hotspot_item (
    id           text PRIMARY KEY,
    name         text NOT NULL,
    category     text,
    price        numeric(14, 2),
    unit         text,
    image_file   text REFERENCES media_asset (filename),
    point_x      double precision,
    point_y      double precision,
    sort_order   int  NOT NULL,
    payload      jsonb NOT NULL,
    payload_text text  NOT NULL
);

CREATE TABLE hotspot_analogue (
    id           text PRIMARY KEY,
    name         text NOT NULL,
    short_name   text,
    price        numeric(14, 2),
    group_name   text,
    unit         text,
    url          text,
    is_external  boolean NOT NULL DEFAULT false,
    image_file   text REFERENCES media_asset (filename),
    sort_order   int  NOT NULL,
    payload      jsonb NOT NULL,
    payload_text text  NOT NULL
);

CREATE TABLE reference_item (
    id           text PRIMARY KEY,
    title        text NOT NULL,
    category     text,
    image_key    text REFERENCES image_key (key),
    styles       jsonb,
    price_from   numeric(14, 2),
    price_to     numeric(14, 2),
    sort_order   int  NOT NULL,
    payload      jsonb NOT NULL,
    payload_text text  NOT NULL
);

CREATE TABLE connected_pick (
    id           text PRIMARY KEY,
    image_key    text REFERENCES image_key (key),
    title        text NOT NULL,
    sort_order   int  NOT NULL,
    payload      jsonb NOT NULL,
    payload_text text  NOT NULL
);

CREATE TABLE connected_item (
    id         text NOT NULL REFERENCES connected_pick (id) ON DELETE CASCADE,
    position   int  NOT NULL,
    sku        text,
    brand      text,
    title      text NOT NULL,
    price      numeric(14, 2),
    image_key  text REFERENCES image_key (key),
    payload    jsonb NOT NULL,
    PRIMARY KEY (id, position)
);

CREATE TABLE planner (
    id         text PRIMARY KEY,
    name       text NOT NULL,
    area_m2    numeric(10, 2),
    sort_order int NOT NULL,
    payload      jsonb NOT NULL,
    payload_text text  NOT NULL
);

CREATE TABLE planner_room (
    planner_id text NOT NULL REFERENCES planner (id) ON DELETE CASCADE,
    position   int  NOT NULL,
    name       text NOT NULL,
    area_m2    numeric(10, 2),
    PRIMARY KEY (planner_id, position)
);

-- ───────────────────────────────────────────────────────────────────────────
-- Контекст Assets: CSS, разметка приложений и фрагменты прототипов
-- ───────────────────────────────────────────────────────────────────────────

CREATE TABLE content_file (
    path         text PRIMARY KEY,
    kind         text NOT NULL CHECK (kind IN ('app', 'main_section', 'page', 'css')),
    app_name     text,
    proto        text CHECK (proto IN ('a', 'b')),
    name         text NOT NULL,
    content_type text NOT NULL,
    body         text NOT NULL,
    sha256       text NOT NULL,
    byte_size    bigint NOT NULL,
    sort_order   int NOT NULL DEFAULT 0
);

CREATE INDEX content_file_kind_idx ON content_file (kind, proto, app_name);

-- ───────────────────────────────────────────────────────────────────────────
-- Контекст Design: 4 варианта дизайна (заглушка вместо загрузки фото)
-- ───────────────────────────────────────────────────────────────────────────

CREATE TABLE design_variant (
    id            text PRIMARY KEY,
    name          text NOT NULL,
    style         text NOT NULL,
    palette       text NOT NULL,
    image_path    text NOT NULL,   -- /api/images/file/… как в API
    image_file    text REFERENCES media_asset (filename),
    description   text NOT NULL,
    budget_min    bigint NOT NULL,
    budget_max    bigint NOT NULL,
    estimated_days int   NOT NULL,
    sort_order    int    NOT NULL,
    payload_text  text   NOT NULL   -- ответ GET /api/design/variants байт-в-байт
);

CREATE TABLE design_variant_room (
    variant_id text NOT NULL REFERENCES design_variant (id) ON DELETE CASCADE,
    position   int  NOT NULL,
    room       text NOT NULL,
    PRIMARY KEY (variant_id, position)
);

-- ───────────────────────────────────────────────────────────────────────────
-- Контекст Estimate и CPR
-- ───────────────────────────────────────────────────────────────────────────

CREATE TABLE estimate (
    id           text PRIMARY KEY,
    status       text NOT NULL,
    total_rub    bigint NOT NULL,
    range_min    bigint NOT NULL,
    range_max    bigint NOT NULL,
    days         int    NOT NULL,
    variant_id   text REFERENCES design_variant (id),
    variant_name text,
    project_id   text,
    payload_text text   NOT NULL   -- ответ GET /api/estimate/{id} байт-в-байт
);

CREATE TABLE estimate_line (
    estimate_id text NOT NULL REFERENCES estimate (id) ON DELETE CASCADE,
    id          text NOT NULL,
    position    int  NOT NULL,
    title       text NOT NULL,
    qty         numeric(14, 3) NOT NULL,
    unit        text NOT NULL,
    price       numeric(14, 2) NOT NULL,
    sku         text,
    brand       text,
    PRIMARY KEY (estimate_id, id)
);

CREATE TABLE estimate_assumption (
    estimate_id text NOT NULL REFERENCES estimate (id) ON DELETE CASCADE,
    position    int  NOT NULL,
    text        text NOT NULL,
    PRIMARY KEY (estimate_id, position)
);

CREATE TABLE cpr (
    id         text PRIMARY KEY,
    version    int  NOT NULL,
    project_id text NOT NULL,
    estimate_id text REFERENCES estimate (id),
    status     text NOT NULL
);

CREATE TABLE cpr_document (
    cpr_id   text NOT NULL REFERENCES cpr (id) ON DELETE CASCADE,
    id       text NOT NULL,
    position int  NOT NULL,
    title    text NOT NULL,
    type     text NOT NULL,
    status   text NOT NULL,
    PRIMARY KEY (cpr_id, id)
);

-- ───────────────────────────────────────────────────────────────────────────
-- Контекст Techcard
-- ───────────────────────────────────────────────────────────────────────────

CREATE TABLE techcard (
    id             text PRIMARY KEY,
    code           text NOT NULL,
    title          text NOT NULL,
    surface        text NOT NULL,
    material       text NOT NULL,
    version        int  NOT NULL,
    min_area_m2    numeric(10, 2),
    max_unevenness numeric(10, 2),
    sort_order     int  NOT NULL,
    payload        jsonb NOT NULL,
    payload_text   text  NOT NULL
);

CREATE INDEX techcard_surface_idx ON techcard (surface);

CREATE TABLE techcard_operation (
    techcard_id text NOT NULL REFERENCES techcard (id) ON DELETE CASCADE,
    position    int  NOT NULL,
    name        text NOT NULL,
    PRIMARY KEY (techcard_id, position)
);

CREATE TABLE techcard_consumption (
    techcard_id text NOT NULL REFERENCES techcard (id) ON DELETE CASCADE,
    name        text NOT NULL,
    value       numeric(14, 4) NOT NULL,
    PRIMARY KEY (techcard_id, name)
);

-- ───────────────────────────────────────────────────────────────────────────
-- Контекст Tender: аукцион подрядчиков
-- ───────────────────────────────────────────────────────────────────────────

CREATE TABLE contractor_bid (
    id         text PRIMARY KEY,
    name       text NOT NULL,
    rating     numeric(3, 1),
    price      numeric(14, 2),
    days       int,
    scope      text,
    image_key  text REFERENCES image_key (key),
    variant_id text REFERENCES design_variant (id),  -- NULL для статичного /api/bids
    sort_order int NOT NULL,
    payload    jsonb NOT NULL,
    payload_text text NOT NULL
);

-- ───────────────────────────────────────────────────────────────────────────
-- Контекст Order: корзина (в mock — заглушка)
-- ───────────────────────────────────────────────────────────────────────────

CREATE TABLE cart_item (
    id         text PRIMARY KEY,
    sku        text NOT NULL,
    brand      text,
    title      text NOT NULL,
    price      numeric(14, 2) NOT NULL,
    qty        int NOT NULL,
    image_key  text REFERENCES image_key (key),
    sort_order int NOT NULL
);

-- ───────────────────────────────────────────────────────────────────────────
-- Контекст Identity
-- ───────────────────────────────────────────────────────────────────────────

CREATE TABLE app_user (
    id            text PRIMARY KEY,
    email         text NOT NULL UNIQUE,
    name          text NOT NULL,
    password_hash text NOT NULL,
    created_at    timestamptz NOT NULL DEFAULT now()
);

-- ───────────────────────────────────────────────────────────────────────────
-- Контекст Booking: запись в Интерьерный центр (write с Idempotency-Key)
-- ───────────────────────────────────────────────────────────────────────────

CREATE TABLE booking_option (
    kind       text NOT NULL CHECK (kind IN ('task', 'time')),
    value      text NOT NULL,
    sort_order int  NOT NULL,
    PRIMARY KEY (kind, value)
);

CREATE TABLE appointment (
    id             text PRIMARY KEY,
    status         text NOT NULL,
    created_at     timestamptz NOT NULL,
    name           text NOT NULL,
    phone          text NOT NULL,
    tasks          jsonb NOT NULL DEFAULT '[]'::jsonb,
    preferred_date date,
    preferred_time text NOT NULL,
    ideas          jsonb NOT NULL DEFAULT '[]'::jsonb,
    files          jsonb NOT NULL DEFAULT '[]'::jsonb,
    wish           text NOT NULL DEFAULT '',
    consent        boolean NOT NULL DEFAULT false,
    place          text NOT NULL,
    validation     jsonb NOT NULL
);

-- Idempotency-Key: ключ + хэш тела → сохранённый ответ (7 суток по ADR).
CREATE TABLE idempotency_key (
    key          text PRIMARY KEY,
    scope        text NOT NULL,
    request_hash text NOT NULL,
    status_code  int  NOT NULL,
    response     jsonb NOT NULL,
    created_at   timestamptz NOT NULL DEFAULT now(),
    expires_at   timestamptz NOT NULL
);

CREATE INDEX idempotency_key_expires_idx ON idempotency_key (expires_at);

-- ───────────────────────────────────────────────────────────────────────────
-- Служебное
-- ───────────────────────────────────────────────────────────────────────────

CREATE TABLE schema_migration (
    version    text PRIMARY KEY,
    applied_at timestamptz NOT NULL DEFAULT now()
);

INSERT INTO schema_migration (version) VALUES ('0001_init');

COMMIT;
