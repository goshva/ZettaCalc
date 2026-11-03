<#
.SYNOPSIS
    Инициализация AI-first проекта ZettaCalc (Fresh House CPR).
.DESCRIPTION
    Разворачивает структуру репозитория, AGENTS.md, правила для AI-агентов,
    заготовки сервисов, OpenAPI-схему, JSON Schema, тесты и CI.
    Подключает remote git@github.com:goshva/ZettaCalc.git.
.PARAMETER Root
    Корневая папка проекта. По умолчанию D:\soft\ZettaCalc
.PARAMETER RemoteUrl
    URL git-remote. По умолчанию git@github.com:goshva/ZettaCalc.git
.PARAMETER Force
    Перезаписывать существующие файлы.
.EXAMPLE
    PS D:\soft\ZettaCalc> .\Init-AiFirstProject.ps1 -Force
#>

[CmdletBinding()]
param(
    [string]$Root = "D:\soft\ZettaCalc",
    [string]$RemoteUrl = "git@github.com:goshva/ZettaCalc.git",
    [switch]$Force
)

$ErrorActionPreference = "Stop"

function Write-Step([string]$msg) {
    Write-Host "==> $msg" -ForegroundColor Cyan
}

function Write-Ok([string]$msg) {
    Write-Host "    ok: $msg" -ForegroundColor Green
}

function Write-Skip([string]$msg) {
    Write-Host "    skip: $msg" -ForegroundColor DarkYellow
}

function Ensure-Dir([string]$path) {
    if (-not (Test-Path -LiteralPath $path)) {
        New-Item -ItemType Directory -Path $path -Force | Out-Null
        Write-Ok "dir $path"
    }
}

function Write-File([string]$path, [string]$content) {
    if ((Test-Path -LiteralPath $path) -and (-not $Force)) {
        Write-Skip "file $path (exists, use -Force)"
        return
    }
    $dir = Split-Path -Parent $path
    if ($dir) { Ensure-Dir $dir }
    Set-Content -LiteralPath $path -Value $content -Encoding UTF8
    Write-Ok "file $path"
}

function Test-Command([string]$name) {
    return [bool](Get-Command $name -ErrorAction SilentlyContinue)
}

# ---------------------------------------------------------------------------
# 0. Preflight
# ---------------------------------------------------------------------------
Write-Step "Preflight: проверка инструментов"

if (-not (Test-Command "git")) {
    throw "git не найден в PATH. Установите Git for Windows."
}
Write-Ok "git $((git --version) -replace 'git version ','')"

$pythonCmd = $null
foreach ($c in @("python","py")) {
    if (Test-Command $c) { $pythonCmd = $c; break }
}
if (-not $pythonCmd) {
    Write-Warning "python не найден. Скрипт продолжит, но тесты и линтеры потребуют Python."
} else {
    $prevEAP = $ErrorActionPreference
    $ErrorActionPreference = "Continue"
    try {
        $pyVersion = (& $pythonCmd --version 2>&1 | Out-String).Trim()
    } finally {
        $ErrorActionPreference = $prevEAP
    }
    Write-Ok "$pyVersion"
}

# ---------------------------------------------------------------------------
# 1. Корень проекта
# ---------------------------------------------------------------------------
Write-Step "Корень проекта: $Root"
Ensure-Dir $Root
Set-Location -LiteralPath $Root

# Сохраняем существующие прототипы и стратегию (не трогаем)
$existing = Get-ChildItem -LiteralPath $Root -File -ErrorAction SilentlyContinue |
    Where-Object { $_.Extension -in @(".html", ".docx", ".pdf") }
if ($existing) {
    Write-Ok "найдены исходные файлы: $($existing.Name -join ', ')"
}

# ---------------------------------------------------------------------------
# 2. Структура каталогов
# ---------------------------------------------------------------------------
Write-Step "Создание структуры каталогов"

$dirs = @(
    "docs",
    "docs/architecture",
    "docs/adr",
    "docs/ai",
    "docs/api",
    "schemas",
    "schemas/json",
    "schemas/openapi",
    "src",
    "src/api",
    "src/api/routers",
    "src/api/schemas",
    "src/domain",
    "src/domain/catalog",
    "src/domain/geometry",
    "src/domain/estimate",
    "src/domain/cpr",
    "src/domain/tender",
    "src/domain/orders",
    "src/rules",
    "src/rules/techcards",
    "src/harness",
    "src/harness/tools",
    "src/harness/skills",
    "src/harness/guardrails",
    "src/harness/evals",
    "src/infra",
    "tests",
    "tests/unit",
    "tests/contract",
    "tests/evals",
    "tests/fixtures",
    "scripts",
    "infra",
    "infra/docker",
    ".github",
    ".github/workflows",
    ".ai",
    ".ai/prompts",
    ".ai/context",
    ".ai/evals",
    ".cursor",
    ".vscode"
)
foreach ($d in $dirs) { Ensure-Dir $d }

# ---------------------------------------------------------------------------
# 3. AGENTS.md — главный контракт для AI-агентов
# ---------------------------------------------------------------------------
Write-Step "AGENTS.md"

$agentsMd = @'
# AGENTS.md — правила для AI-агентов в репозитории ZettaCalc

> Этот файл — единственный источник правды для AI-агентов (Cursor, Copilot,
> Claude Code, локальные агенты). Любой агент перед изменением кода обязан
> прочитать этот файл и `docs/architecture/overview.md`.

## 1. Что это за проект

ZettaCalc — модуль Fresh House / ZETTA Home: анкета и геометрия помещения >
предварительный расчёт > ЦПР (цифровой паспорт ремонта) > аукцион подрядчиков >
заказ. См. `docs/architecture/overview.md`.

## 2. Базовые принципы (нерушимые)

1. **Детерминированное ядро не вызывает LLM.** PIM, геометрия, rule engine,
   генератор ЦПР — только детерминированный код. LLM готовит вход (контекст,
   подсказки, разбор фото), но не считает смету.
2. **Цены, SKU и габариты — только из PIM.** Никаких «придуманных» цен.
3. **Права доступа проверяются только code-based evals.** ACL нельзя оставлять
   «на усмотрение LLM».
4. **Write-операции** (заказ, изменение сметы, передача ТЗ, approve ЦПР) —
   только с `Idempotency-Key`, dry-run и approval-gate человеком.
5. **Автосборка ЦПР** оценивается метрикой `pass^k` (все k прогонов успешны),
   а не `pass@k`.
6. **Рой агентов не нужен** на линейном пути «бриф > планировка > технологии >
   ЦПР». Только на параллельных ветках (сравнение гипотез).

## 3. Что можно и нельзя менять агенту

| Зона | Разрешено | Запрещено |
|---|---|---|
| `src/domain/catalog` | Добавлять поля, тесты | Менять контракт цен без ADR |
| `src/domain/geometry` | Оптимизация, тесты | Менять единицы (1 unit = 1 m) |
| `src/rules/techcards` | Добавлять техкарты с тестами | Удалять/менять обязательные операции |
| `src/domain/cpr` | Новые поля, версии | Менять формат без миграции |
| `src/harness` | Новые tools, evals | Ослаблять guardrails |
| `schemas/` | Добавлять поля с `x-` | Удалять обязательные поля |
| `AGENTS.md` | Ничего | Любые изменения без человека |

## 4. Обязательные проверки перед PR

- `pytest tests/unit tests/contract` — зелёные.
- `ruff check src tests` — без ошибок.
- `mypy src` — без ошибок (strict для `src/domain`).
- `scripts/check_guardrails.ps1` — проверка, что guardrails не ослаблены.
- `scripts/check_techcards.ps1` — у каждой техкарты есть версия, ответственный и тестовый пример.

## 5. Формат коммитов

Conventional Commits: `feat(domain/cpr): ...`, `fix(geometry): ...`,
`test(evals): ...`, `docs(adr): ...`.

## 6. Контекст для агента

- Архитектура: `docs/architecture/overview.md`
- ADR: `docs/adr/`
- OpenAPI: `schemas/openapi/cpr.yaml`
- JSON Schema ЦПР: `schemas/json/cpr.schema.json`
- Промпты: `.ai/prompts/`
- Evals: `.ai/evals/`

## 7. Запрещённые действия

- Коммитить секреты, токены, персональные данные.
- Менять `.github/workflows` без человека.
- Отключать `Idempotency-Key` на write-эндпоинтах.
- Генерировать цены/SKU «из головы».
'@
Write-File (Join-Path $Root "AGENTS.md") $agentsMd

# ---------------------------------------------------------------------------
# 4. .cursor/rules и .vscode
# ---------------------------------------------------------------------------
Write-Step "Правила для IDE-агентов (.cursor, .vscode)"

$cursorRules = @'
---
description: ZettaCalc AI-first rules
globs: ["src/**/*.py", "schemas/**/*.yaml", "tests/**/*.py"]
---
- Читай AGENTS.md перед любым изменением.
- Детерминированное ядро не вызывает LLM.
- Цены и SKU — только из PIM (`src/domain/catalog`).
- Для write-операций — Idempotency-Key и approval-gate.
- Не ослабляй guardrails в `src/harness/guardrails`.
- Добавляй тест на каждое изменение бизнес-правила.
'@
Write-File (Join-Path $Root ".cursor/rules.mdc") $cursorRules

$vscodeSettings = @'
{
  "python.analysis.typeCheckingMode": "strict",
  "editor.formatOnSave": true,
  "files.trimTrailingWhitespace": true,
  "files.insertFinalNewline": true,
  "yaml.schemas": {
    "https://json.schemastore.org/openapi-3.1.json": "schemas/openapi/*.yaml"
  }
}
'@
Write-File (Join-Path $Root ".vscode/settings.json") $vscodeSettings

# ---------------------------------------------------------------------------
# 5. .gitignore, .editorconfig, .gitattributes
# ---------------------------------------------------------------------------
Write-Step "Базовые файлы репозитория"

$gitignore = @'
# Python
__pycache__/
*.py[cod]
.venv/
.env
.env.*
!.env.example
.pytest_cache/
.mypy_cache/
.ruff_cache/
coverage.xml
htmlcov/

# Node (если добавим фронтенд)
node_modules/
dist/
build/

# IDE
.idea/
*.swp
.DS_Store

# Артефакты
*.log
*.tmp
artifacts/
reports/
'@
Write-File (Join-Path $Root ".gitignore") $gitignore

$editorconfig = @'
root = true

[*]
charset = utf-8
end_of_line = lf
insert_final_newline = true
trim_trailing_whitespace = true
indent_style = space
indent_size = 4

[*.{yaml,yml,json,md}]
indent_size = 2
'@
Write-File (Join-Path $Root ".editorconfig") $editorconfig

$gitattributes = @'
* text=auto eol=lf
*.ps1 text eol=crlf
*.bat text eol=crlf
*.png binary
*.jpg binary
*.webp binary
*.pdf binary
*.docx binary
'@
Write-File (Join-Path $Root ".gitattributes") $gitattributes

$envExample = @'
# Fresh House CPR API
APP_ENV=local
APP_PORT=8080
DATABASE_URL=postgresql+asyncpg://freshhouse:freshhouse@localhost:5432/freshhouse
S3_ENDPOINT=http://localhost:9000
S3_BUCKET=fresh-house
S3_ACCESS_KEY=minio
S3_SECRET_KEY=minio123

# LLM (только для harness, не для ядра)
LLM_PROVIDER=openai
LLM_API_KEY=
LLM_MODEL_CHEAP=gpt-4o-mini
LLM_MODEL_STRONG=gpt-4o

# Guardrails
ALLOW_WRITE=false
APPROVAL_REQUIRED=true
'@
Write-File (Join-Path $Root ".env.example") $envExample

# ---------------------------------------------------------------------------
# 6. docs/architecture/overview.md
# ---------------------------------------------------------------------------
Write-Step "Архитектурный обзор"

$overview = @'
# Архитектура ZettaCalc (Fresh House CPR)

## Поток

Фото/чертежи + бриф > Геометрия (контур в мм, сцена) > Предварительный расчёт
(калькулятор) > ЦПР (версии) > Согласование экспертом > Аукцион подрядчиков >
Заказ товаров/услуг > Личный кабинет.

## Слои данных

- ODS: выгрузки ZETTA, Zetta Home, поставщиков, подрядчиков, клиента.
- DDS: справочник SKU + посадочные размеры, версионные цены, библиотека
  техкарт, правила расчёта, шаблоны чертежей.
- UDS: ИнтерьернаяИдея, ВариантРеализации, ПроектКлиента, Помещение+Геометрия,
  Комплектация, ЦПР (версия), Заказ/ПередачаТЗ.
- DM: калькулятор, подбор технологий, семантический слой для Text2SQL,
  аналитика спроса.

**S2T-маппинг — gate** для любого нового источника каталога.

## Сервисы

- Детерминированное ядро: Catalog & Pricing (PIM), Geometry Service,
  Rule Engine (техкарты), CPR Generator, Project Service, Order & Contractor.
- ИИ-слой (harness): Vision (выравнивание, углы, размеры), Context engineering
  (бриф, допущения, вопросы), Semantic Search (референс > вариант),
  Text2SQL (семантический слой), Render Proxy (апселл).

## Harness / guardrails

- Runtime: бюджеты шагов/токенов, ретраи, anti-loop.
- Tool layer по JSON Schema; MCP-слой к dbt/Airflow/Greenplum (read-only).
- Skills под роли (SA/DE/BI/дизайнер), subagents, hooks, sandbox (NemoClaw-подход).
- Evals: code-based (ACL, JSON, SKU, площади), LLM-as-judge (калибровка),
  human evaluation (edge-кейсы). Метрика — `pass^k`.
- Write-safety: dry-run, idempotency, approval-gate, deny-list, аудит.
'@
Write-File (Join-Path $Root "docs/architecture/overview.md") $overview

# ---------------------------------------------------------------------------
# 7. ADR
# ---------------------------------------------------------------------------
Write-Step "ADR"

$adr1 = @'
# ADR-0001: Детерминированное ядро не вызывает LLM

## Статус
Принято.

## Контекст
LLM может «додумывать» цены, SKU и операции. Для сметы и ЦПР это недопустимо.

## Решение
Расчёты, подбор технологий и генерация ЦПР выполняются только детерминированным
кодом (rule engine + PIM). LLM допускается только для подготовки входа
(контекст, разбор фото, подсказки) и проходит через harness/guardrails.

## Последствия
- Дороже в разработке rule engine, но предсказуемый результат.
- Все изменения техкарт — через версии, ответственных и тестовые примеры.
'@
Write-File (Join-Path $Root "docs/adr/0001-deterministic-core.md") $adr1

$adr2 = @'
# ADR-0002: Метрика надёжности автосборки ЦПР — pass^k

## Статус
Принято.

## Контекст
pass@k («хотя бы один из k успешен») вводит в заблуждение для автономных
агентов. Нужна стабильность, а не «иногда получается».

## Решение
Для автосборки ЦПР использовать `pass^k` (все k прогонов успешны).
Цель v1 — 50–60% без правок, стратегическая — ? 90% при ограничении
номенклатуры проверенных решений.

## Последствия
- Evals должны считаться на k прогонах, а не на одном.
- Human handoff — штатный механизм, а не «провал».
'@
Write-File (Join-Path $Root "docs/adr/0002-pass-k-metric.md") $adr2

# ---------------------------------------------------------------------------
# 8. JSON Schema ЦПР
# ---------------------------------------------------------------------------
Write-Step "JSON Schema ЦПР"

$cprSchema = @'
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://fresh-house.ru/schemas/cpr/v1.json",
  "title": "CPR (Цифровой паспорт ремонта)",
  "type": "object",
  "required": ["cpr_id", "version", "status", "source", "client", "rooms", "specification", "estimate", "deliverables", "audit"],
  "properties": {
    "cpr_id": { "type": "string", "pattern": "^CPR-[0-9]{4}-[0-9]{6}$" },
    "version": { "type": "integer", "minimum": 1 },
    "status": {
      "type": "string",
      "enum": ["draft", "pending_review", "approved_for_tender", "in_tender", "awarded", "rejected"]
    },
    "created_at": { "type": "string", "format": "date-time" },
    "updated_at": { "type": "string", "format": "date-time" },
    "source": {
      "type": "object",
      "required": ["reference_id", "implementation_variant", "brief_id", "geometry_id", "catalog_snapshot"],
      "properties": {
        "reference_id": { "type": "string" },
        "implementation_variant": { "type": "string", "enum": ["econom", "balanced", "fast", "premium"] },
        "brief_id": { "type": "string" },
        "geometry_id": { "type": "string" },
        "catalog_snapshot": { "type": "string" }
      }
    },
    "client": {
      "type": "object",
      "required": ["project_id", "assumptions"],
      "properties": {
        "project_id": { "type": "string" },
        "apartment_area_m2": { "type": "number" },
        "rooms_count": { "type": "integer" },
        "assumptions": { "type": "array", "items": { "type": "string" } }
      }
    },
    "rooms": {
      "type": "array",
      "items": {
        "type": "object",
        "required": ["room_id", "type", "geometry", "completeness_checklist"],
        "properties": {
          "room_id": { "type": "string" },
          "type": { "type": "string", "enum": ["kitchen", "bathroom", "bedroom", "living", "hallway", "kids", "office"] },
          "area_m2": { "type": "number" },
          "ceiling_height_m": { "type": "number" },
          "geometry": { "type": "object", "additionalProperties": true },
          "layout_type": { "type": "string" },
          "completeness_checklist": { "type": "object", "additionalProperties": { "type": "boolean" } },
          "surfaces": { "type": "object", "additionalProperties": true },
          "furniture": { "type": "array", "items": { "type": "object", "additionalProperties": true } },
          "lighting": { "type": "object", "additionalProperties": true },
          "works": { "type": "array", "items": { "type": "object", "additionalProperties": true } }
        }
      }
    },
    "specification": {
      "type": "object",
      "required": ["materials", "furniture", "lighting"],
      "properties": {
        "materials": { "type": "array", "items": { "type": "object", "additionalProperties": true } },
        "furniture": { "type": "array", "items": { "type": "object", "additionalProperties": true } },
        "lighting": { "type": "array", "items": { "type": "object", "additionalProperties": true } }
      }
    },
    "estimate": {
      "type": "object",
      "required": ["total_rub", "currency", "disclaimer"],
      "properties": {
        "materials_rub": { "type": "number" },
        "furniture_rub": { "type": "number" },
        "lighting_rub": { "type": "number" },
        "works_rub": { "type": "number" },
        "total_rub": { "type": "number" },
        "currency": { "type": "string", "default": "RUB" },
        "disclaimer": { "type": "string" }
      }
    },
    "deliverables": {
      "type": "object",
      "required": ["drawings", "schedules", "attachments"],
      "properties": {
        "drawings": { "type": "array", "items": { "type": "object", "additionalProperties": true } },
        "schedules": { "type": "array", "items": { "type": "object", "additionalProperties": true } },
        "attachments": { "type": "array", "items": { "type": "object", "additionalProperties": true } }
      }
    },
    "tender": { "type": ["object", "null"] },
    "audit": {
      "type": "object",
      "required": ["created_by", "changelog"],
      "properties": {
        "created_by": { "type": "string" },
        "approved_by": { "type": ["string", "null"] },
        "approval_at": { "type": ["string", "null"], "format": "date-time" },
        "changelog": {
          "type": "array",
          "items": {
            "type": "object",
            "required": ["version", "at", "note"],
            "properties": {
              "version": { "type": "integer" },
              "at": { "type": "string", "format": "date-time" },
              "note": { "type": "string" }
            }
          }
        }
      }
    }
  }
}
'@
Write-File (Join-Path $Root "schemas/json/cpr.schema.json") $cprSchema

# ---------------------------------------------------------------------------
# 9. OpenAPI-заглушка (ссылка на полную спецификацию из ответа)
# ---------------------------------------------------------------------------
Write-Step "OpenAPI-заглушка"

$openapiStub = @'
openapi: 3.1.0
info:
  title: Fresh House CPR API
  version: 1.0.0
  description: |
    Заглушка. Полная спецификация — в чате / docs/api/cpr.openapi.yaml.
paths: {}
components: {}
'@
Write-File (Join-Path $Root "schemas/openapi/cpr.yaml") $openapiStub

# ---------------------------------------------------------------------------
# 10. Domain-заготовки (Python)
# ---------------------------------------------------------------------------
Write-Step "Заготовки доменных сервисов"

$pyInit = @'
"""ZettaCalc domain package (детерминированное ядро)."""
'@

Write-File (Join-Path $Root "src/__init__.py") ""
Write-File (Join-Path $Root "src/domain/__init__.py") $pyInit
Write-File (Join-Path $Root "src/api/__init__.py") ""
Write-File (Join-Path $Root "src/harness/__init__.py") ""
Write-File (Join-Path $Root "src/rules/__init__.py") ""

$catalogPy = @'
"""Catalog & Pricing (PIM): SKU, посадочные размеры, цены.

Правило: цены и SKU — только из PIM. Никаких значений «из головы».
"""
from __future__ import annotations

from dataclasses import dataclass


@dataclass(frozen=True)
class Sku:
    sku: str
    name: str
    category: str
    width_mm: int
    depth_mm: int
    height_mm: int
    seat_width_mm: int | None = None
    seat_depth_mm: int | None = None
    price_rub: float = 0.0


def filter_by_fit(items: list[Sku], seat_width_mm: int, tolerance_mm: int = 5) -> list[Sku]:
    """Сценарий замены: отфильтровать SKU по посадочному размеру."""
    return [
        s for s in items
        if s.seat_width_mm is not None
        and abs(s.seat_width_mm - seat_width_mm) <= tolerance_mm
    ]
'@
Write-File (Join-Path $Root "src/domain/catalog/__init__.py") $catalogPy

$geometryPy = @'
"""Geometry Service: контур в мм, сцена (1 unit = 1 m)."""
from __future__ import annotations

from dataclasses import dataclass


@dataclass(frozen=True)
class Contour:
    points_mm: list[tuple[float, float]]

    def area_m2(self) -> float:
        # Shoelace
        pts = self.points_mm
        s = 0.0
        for i in range(len(pts)):
            x1, y1 = pts[i]
            x2, y2 = pts[(i + 1) % len(pts)]
            s += x1 * y2 - x2 * y1
        return abs(s) / 2.0 / 1_000_000.0
'@
Write-File (Join-Path $Root "src/domain/geometry/__init__.py") $geometryPy

$cprPy = @'
"""CPR Generator: сборка ЦПР из входа (rule engine + PIM).

Правило: обязательные операции нельзя молча убирать ради более низкой цены.
"""
from __future__ import annotations

from dataclasses import dataclass, field


@dataclass
class Cpr:
    cpr_id: str
    version: int = 1
    status: str = "draft"
    assumptions: list[str] = field(default_factory=list)
    checklist: dict[str, bool] = field(default_factory=dict)

    def can_approve(self) -> bool:
        return all(self.checklist.values()) if self.checklist else False
'@
Write-File (Join-Path $Root "src/domain/cpr/__init__.py") $cprPy

$techcardExample = @'
{
  "code": "FLOOR-KER-600",
  "version": 3,
  "owner": "expert-001",
  "title": "Укладка керамогранита 600?600",
  "conditions": {
    "surface": "floor",
    "material": "keramogranit",
    "min_area_m2": 1
  },
  "operations": [
    {"code": "PREP-01", "name": "Подготовка основания", "unit": "m2"},
    {"code": "FLOOR-01", "name": "Укладка покрытия", "unit": "m2"},
    {"code": "GROUT-01", "name": "Затирка швов", "unit": "m2"}
  ],
  "consumption": {"glue_kg_per_m2": 4.5, "grout_kg_per_m2": 0.5},
  "limitations": ["Не применять при неровности основания > 5 мм/м"],
  "test_examples": ["tests/fixtures/techcards/floor_ker_600.json"]
}
'@
Write-File (Join-Path $Root "src/rules/techcards/floor_ker_600.json") $techcardExample

# ---------------------------------------------------------------------------
# 11. Harness-заготовки
# ---------------------------------------------------------------------------
Write-Step "Harness: runtime, guardrails, evals"

$runtimePy = @'
"""Harness runtime: бюджеты, ретраи, anti-loop, трейсы.

Правило: детерминированное ядро не вызывает LLM. Harness — единственная
точка вызова LLM.
"""
from __future__ import annotations

from dataclasses import dataclass, field


@dataclass
class Budget:
    max_steps: int = 20
    max_tokens: int = 50_000
    max_cost_rub: float = 50.0


@dataclass
class Trace:
    steps: list[dict] = field(default_factory=list)

    def log(self, tool: str, args: dict, result: dict, cost_rub: float) -> None:
        self.steps.append({"tool": tool, "args": args, "result": result, "cost_rub": cost_rub})


def run_agent(task: str, budget: Budget) -> Trace:
    """Заглушка: здесь будет цикл агента с tool layer и hooks."""
    raise NotImplementedError("Реализовать runtime с бюджетами и anti-loop")
'@
Write-File (Join-Path $Root "src/harness/runtime.py") $runtimePy

$guardrailsPy = @'
"""Code-based guardrails: ACL, JSON, SKU, площади, write-safety.

Правило: права доступа проверяются только code-based, не LLM.
"""
from __future__ import annotations


def check_acl(actor_roles: list[str], required: str) -> bool:
    return required in actor_roles


def check_sku_exists(sku: str, known_skus: set[str]) -> bool:
    return sku in known_skus


def check_write_allowed(allow_write: bool, approved: bool) -> bool:
    """Write-операция допустима только при allow_write и approval."""
    return allow_write and approved
'@
Write-File (Join-Path $Root "src/harness/guardrails/__init__.py") $guardrailsPy

$evalsPy = @'
"""Evals: pass^k для автосборки ЦПР.

Правило: метрика — pass^k (все k прогонов успешны), не pass@k.
"""
from __future__ import annotations


def pass_hat_k(successes: list[bool]) -> bool:
    """Все прогоны успешны."""
    return all(successes) if successes else False


def pass_at_k(successes: list[bool]) -> bool:
    """Хотя бы один успешен (для сравнения, не для продакшна)."""
    return any(successes) if successes else False
'@
Write-File (Join-Path $Root "src/harness/evals/__init__.py") $evalsPy

# ---------------------------------------------------------------------------
# 12. AI-контекст и промпты
# ---------------------------------------------------------------------------
Write-Step "AI-контекст и промпты"

$aiContext = @'
# Контекст проекта для AI-агентов (сжатый)

## Цель
Модуль Fresh House CPR: анкета + геометрия > расчёт > ЦПР > аукцион > заказ.

## Нерушимые правила
1. Ядро не вызывает LLM.
2. Цены/SKU — только из PIM.
3. ACL — только code-based.
4. Write — только с idempotency + approval-gate.
5. pass^k, не pass@k.
6. Рой — только на параллельных ветках.

## Где что лежит
- `src/domain/*` — детерминированное ядро.
- `src/harness/*` — LLM-слой, guardrails, evals.
- `src/rules/techcards/*` — техкарты (версии, ответственные, тесты).
- `schemas/` — OpenAPI и JSON Schema.
- `tests/` — unit, contract, evals.

## Что делать агенту
- Прочитать `AGENTS.md` и этот файл.
- Проверить, не нарушает ли изменение правила.
- Добавить тест.
- Не менять `AGENTS.md` без человека.
'@
Write-File (Join-Path $Root ".ai/context/project.md") $aiContext

$promptBrief = @'
Ты — ассистент по брифу Fresh House.

Задача: из ответов клиента собрать бриф и явно выделить допущения.

Правила:
- Не придумывай размеры и данные, которых нет.
- Всё неизвестное — в `assumptions` как «требует подтверждения».
- Обязательные зоны (холодильник, мойка, закрытые шкафы, проходы)
  вынеси в `completeness_checklist` и проверь, что они отмечены.
- Если данных недостаточно — верни `open_questions`, а не догадку.

Формат ответа: JSON по схеме `schemas/json/cpr.schema.json` (секции client, rooms).
'@
Write-File (Join-Path $Root ".ai/prompts/brief.md") $promptBrief

$promptGeometry = @'
Ты — ассистент по геометрии помещения.

Вход: фото и/или чертёж, опорные точки, расстояния.
Задача: предложить контур в мм и проёмы.

Правила:
- Не меняй пропорции и не «дорисовывай» лишние двери.
- Все распознанные размеры — с флагом `confidence`.
- Если размер не распознан — `null` и `open_question`.
- Опционально: предложи размещение двери на скрытой стене с отступами от углов.

Формат: JSON (geometry).
'@
Write-File (Join-Path $Root ".ai/prompts/geometry.md") $promptGeometry

$promptCprReview = @'
Ты — ассистент по проверке ЦПР перед согласованием экспертом.

Задача: прогнать чек-лист и вернуть список нарушений.

Правила:
- Не пересчитывай смету.
- Проверь: холодильник, мойка, закрытые шкафы, проходы, открывание дверей,
  освещённость по нормам.
- Каждое нарушение — с `path`, `rule`, `message`.
- Если нарушений нет — верни `passed: true`.

Формат: JSON.
'@
Write-File (Join-Path $Root ".ai/prompts/cpr_review.md") $promptCprReview

# ---------------------------------------------------------------------------
# 13. Тесты-заготовки
# ---------------------------------------------------------------------------
Write-Step "Тесты-заготовки"

$testGeometry = @'
from src.domain.geometry import Contour


def test_area_rectangle():
    c = Contour([(0, 0), (3200, 0), (3200, 4000), (0, 4000)])
    assert round(c.area_m2(), 2) == 12.8
'@
Write-File (Join-Path $Root "tests/unit/test_geometry.py") $testGeometry

$testGuardrails = @'
from src.harness.guardrails import check_acl, check_write_allowed


def test_acl_denied():
    assert not check_acl(["client"], "expert")


def test_write_requires_approval():
    assert not check_write_allowed(allow_write=True, approved=False)
    assert check_write_allowed(allow_write=True, approved=True)
'@
Write-File (Join-Path $Root "tests/unit/test_guardrails.py") $testGuardrails

$testEvals = @'
from src.harness.evals import pass_hat_k, pass_at_k


def test_pass_hat_k_all():
    assert pass_hat_k([True, True, True])
    assert not pass_hat_k([True, False, True])


def test_pass_at_k_any():
    assert pass_at_k([False, True, False])
'@
Write-File (Join-Path $Root "tests/unit/test_evals.py") $testEvals

# ---------------------------------------------------------------------------
# 14. pyproject.toml, requirements, Makefile
# ---------------------------------------------------------------------------
Write-Step "Python-конфигурация"

$pyproject = @'
[project]
name = "zettacalc"
version = "0.1.0"
description = "Fresh House CPR module (ZettaCalc)"
requires-python = ">=3.11"
dependencies = [
  "fastapi>=0.115",
  "uvicorn[standard]>=0.30",
  "pydantic>=2.7",
  "sqlalchemy>=2.0",
  "asyncpg>=0.29",
  "boto3>=1.34",
]

[project.optional-dependencies]
dev = [
  "pytest>=8",
  "pytest-asyncio>=0.23",
  "ruff>=0.6",
  "mypy>=1.10",
  "jsonschema>=4.22",
]

[tool.ruff]
line-length = 100
target-version = "py311"

[tool.ruff.lint]
select = ["E", "F", "I", "B", "UP", "SIM"]

[tool.mypy]
python_version = "3.11"
strict = true
files = ["src"]

[tool.pytest.ini_options]
testpaths = ["tests"]
asyncio_mode = "auto"
'@
Write-File (Join-Path $Root "pyproject.toml") $pyproject

$makefile = @'
.PHONY: install lint test contract evals check

install:
	pip install -e ".[dev]"

lint:
	ruff check src tests
	mypy src

test:
	pytest tests/unit -q

contract:
	pytest tests/contract -q

evals:
	pytest tests/evals -q

check: lint test contract evals
'@
Write-File (Join-Path $Root "Makefile") $makefile

# ---------------------------------------------------------------------------
# 15. Скрипты проверки guardrails и техкарт
# ---------------------------------------------------------------------------
Write-Step "Скрипты проверки"

$checkGuardrails = @'
# Проверка, что guardrails не ослаблены.
$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
Set-Location -LiteralPath $root

$fail = $false

$guardrails = Get-Content -Raw "src/harness/guardrails/__init__.py"
if ($guardrails -notmatch "def check_acl") { Write-Host "FAIL: check_acl отсутствует"; $fail = $true }
if ($guardrails -notmatch "def check_write_allowed") { Write-Host "FAIL: check_write_allowed отсутствует"; $fail = $true }

$agents = Get-Content -Raw "AGENTS.md"
if ($agents -notmatch "code-based evals") { Write-Host "FAIL: AGENTS.md не фиксирует code-based ACL"; $fail = $true }

if ($fail) { exit 1 } else { Write-Host "guardrails ok"; exit 0 }
'@
Write-File (Join-Path $Root "scripts/check_guardrails.ps1") $checkGuardrails

$checkTechcards = @'
# Проверка техкарт: версия, ответственный, тестовый пример.
$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
Set-Location -LiteralPath $root

$fail = $false
Get-ChildItem "src/rules/techcards" -Filter *.json | ForEach-Object {
    $tc = Get-Content -Raw $_.FullName | ConvertFrom-Json
    if (-not $tc.version) { Write-Host "FAIL: $($_.Name) без version"; $fail = $true }
    if (-not $tc.owner) { Write-Host "FAIL: $($_.Name) без owner"; $fail = $true }
    if (-not $tc.test_examples -or $tc.test_examples.Count -eq 0) {
        Write-Host "FAIL: $($_.Name) без test_examples"; $fail = $true
    }
}
if ($fail) { exit 1 } else { Write-Host "techcards ok"; exit 0 }
'@
Write-File (Join-Path $Root "scripts/check_techcards.ps1") $checkTechcards

# ---------------------------------------------------------------------------
# 16. GitHub Actions
# ---------------------------------------------------------------------------
Write-Step "CI (GitHub Actions)"

$ci = @'
name: ci

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-python@v5
        with:
          python-version: "3.11"
      - name: Install
        run: pip install -e ".[dev]"
      - name: Lint
        run: ruff check src tests
      - name: Types
        run: mypy src
      - name: Unit tests
        run: pytest tests/unit -q
      - name: Contract tests
        run: pytest tests/contract -q
      - name: Guardrails check
        run: pwsh -File scripts/check_guardrails.ps1
      - name: Techcards check
        run: pwsh -File scripts/check_techcards.ps1
'@
Write-File (Join-Path $Root ".github/workflows/ci.yml") $ci

# ---------------------------------------------------------------------------
# 17. README
# ---------------------------------------------------------------------------
Write-Step "README"

# ---------------------------------------------------------------------------
# 17. README
# ---------------------------------------------------------------------------
Write-Step "README"

$readme = @"
# ZettaCalc — Fresh House CPR module

Модуль маркетплейса ремонта: анкета и геометрия помещения -> предварительный
расчёт -> ЦПР (цифровой паспорт ремонта) -> аукцион подрядчиков -> заказ.

## Структура

- docs/ — архитектура и ADR.
- schemas/ — OpenAPI и JSON Schema ЦПР.
- src/domain/ — детерминированное ядро (catalog, geometry, cpr, estimate, tender, orders).
- src/rules/techcards/ — техкарты (версии, ответственные, тесты).
- src/harness/ — LLM-слой: runtime, guardrails, evals.
- tests/ — unit, contract, evals.
- .ai/ — контекст и промпты для AI-агентов.
- AGENTS.md — правила для AI-агентов (читать первым).

## Быстрый старт

    cd D:\soft\ZettaCalc
    python -m venv .venv
    .\.venv\Scripts\Activate.ps1
    pip install -e ".[dev]"
    pytest tests/unit -q

## Принципы

См. AGENTS.md и docs/architecture/overview.md.

## Прототипы

Исходные файлы (Прототип А.html, Прототип В.html,
Стратегия и задание для CPTO.docx) хранятся в корне и не изменяются
скриптом инициализации.
"@
Write-File (Join-Path $Root "README.md") $readme

# ---------------------------------------------------------------------------
# 18. Git: init / remote / first commit
# ---------------------------------------------------------------------------
Write-Step "Git: init, remote, первый коммит"

$gitDir = Join-Path $Root ".git"
if (-not (Test-Path -LiteralPath $gitDir)) {
    git init | Out-Null
    Write-Ok "git init"
} else {
    Write-Ok "git уже инициализирован"
}

$currentRemote = $null
try {
    $currentRemote = git remote get-url origin 2>$null
} catch {
    $currentRemote = $null
}

if (-not $currentRemote) {
    git remote add origin $RemoteUrl
    Write-Ok "remote origin = $RemoteUrl"
} elseif ($currentRemote -ne $RemoteUrl) {
    git remote set-url origin $RemoteUrl
    Write-Ok "remote origin обновлён = $RemoteUrl"
} else {
    Write-Ok "remote origin уже = $RemoteUrl"
}

git add -A
$status = git status --porcelain
if ($status) {
    git -c user.name="ZettaCalc Bot" -c user.email="bot@fresh-house.local" `
        commit -m "chore: init AI-first project structure" | Out-Null
    Write-Ok "первый коммит создан"
} else {
    Write-Skip "нет изменений для коммита"
}

Write-Step "Готово"
Write-Host @"
Проект инициализирован в: $Root
Remote: $RemoteUrl

Дальше:
  1) Создайте ветку: git checkout -b feat/cpr-skeleton
  2) Откройте AGENTS.md и docs/architecture/overview.md
  3) Запустите: make install && make check
  4) Отправьте: git push -u origin main
"@ -ForegroundColor Yellow