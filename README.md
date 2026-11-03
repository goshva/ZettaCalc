# ZettaCalc — Fresh House CPR module

Модуль маркетплейса ремонта: анкета и геометрия помещения -> предварительный
расчёт -> ЦПР (цифровой паспорт ремонта) -> аукцион подрядчиков -> заказ.

**Начните с [ANSWER.md](ANSWER.md)** (ответ на задание CPTO), затем [todo.md](todo.md)
(статус и процент выполнения).

## Документы

- [docs/product-scope.md](docs/product-scope.md) — v1, отложенное, метрики
- [docs/roadmap.md](docs/roadmap.md) — волны, RACI, DoD
- [docs/risks.md](docs/risks.md) — реестр рисков
- [docs/expert-opinion.md](docs/expert-opinion.md) — что пересмотреть, контраргумент
- [docs/prototypes-comparison.md](docs/prototypes-comparison.md) — прототипы А и В
- [docs/economics.md](docs/economics.md) — гипотезы дохода
- [docs/architecture/](docs/architecture/) — данные, сервисы, harness, ER, S2T, матрица ИИ
- [docs/adr/](docs/adr/) — архитектурные решения

## Структура

- `schemas/` — OpenAPI и JSON Schema ЦПР; `examples/` — примеры ЦПР.
- `src/domain/` — детерминированное ядро (catalog, geometry, estimate, cpr).
- `src/rules/` — rule engine и техкарты (версии, владельцы, тесты).
- `src/harness/` — LLM-слой: runtime, llm (мок), tools, guardrails, evals.
- `tests/` — unit, contract, evals, fixtures.
- `.ai/` — контекст, промпты, описание evals. `AGENTS.md` — правила для агентов (читать первым).

## Быстрый старт

    python -m venv .venv
    source .venv/bin/activate          # Windows: .\.venv\Scripts\Activate.ps1
    pip install -e ".[dev]"
    make check                         # ruff, mypy, unit, contract, evals

## Прототипы

Исходные файлы (Прототип А.html, Прототип В.html, Стратегия и задание для CPTO.docx)
хранятся в корне и не изменяются скриптом инициализации.
