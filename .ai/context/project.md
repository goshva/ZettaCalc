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
