# Harness и guardrails

```mermaid
flowchart LR
  T[Задача] --> RT[Runtime: бюджет шагов, токенов, рублей]
  RT --> LLM[LLMProvider]
  LLM -- tool call --> TL[Tool layer: JSON Schema]
  TL --> GR{Guardrails: ACL, SKU, write-safety}
  GR -- read --> CORE[Детерминированное ядро]
  GR -- write --> AP{dry-run, Idempotency-Key, approval человека}
  AP --> CORE
  RT --> TR[Trace: шаги, стоимость, stop_reason]
  TR --> EV[Evals: code-based, pass^k]
```

Реализовано в `src/harness/`:

| Механизм | Где | Проверяет |
|---|---|---|
| Бюджет шагов / токенов / рублей | `runtime.py` | `tests/unit/test_runtime.py` |
| Anti-loop (3 одинаковых вызова) | `runtime.py` | то же |
| Write только с `allow_write` и `approved` | `runtime.py`, `guardrails` | `test_runtime.py`, `tests/evals/test_guardrails_acl.py` |
| LLM только через `LLMProvider` | `llm.py` | мок `ScriptedLLM` для evals |
| Tool layer (read-only) | `tools/` | `get_catalog_item`, `validate_geometry` |

Не реализовано (волна 2): реальный провайдер LLM, MCP-слой к dbt/Airflow/Greenplum,
Text2SQL по семантическому слою, песочница (NemoClaw-подход), LLM-as-judge с калибровкой.
