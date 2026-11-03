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
