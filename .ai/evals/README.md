# .ai/evals

Evals живут в `tests/evals/` (code-based, запуск: `pytest tests/evals -q`).

| Eval | Что проверяет | Тип |
|---|---|---|
| `test_cpr_evals.py` | схема ЦПР, площади из контура, SKU только из PIM, статус, `pass^k` (k=10) | code-based |
| `test_guardrails_acl.py` | ACL и write-safety нельзя обойти | code-based |
| `test_prompts_contract.py` | в промптах остались запреты (не придумывать, `assumptions`) | code-based |

Следующий шаг (волна 2): LLM-as-judge с калибровкой на размеченных кейсах и human evaluation
для edge-кейсов. До калибровки judge не используется как гейт.
