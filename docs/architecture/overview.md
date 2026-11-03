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
