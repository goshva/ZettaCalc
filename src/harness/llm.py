"""LLM-абстракция. Единственная точка, через которую harness говорит с моделью.

Реальный провайдер подключается позже (LLM_PROVIDER в .env). Тесты и evals
идут на ScriptedLLM, без сети и без ключей.
"""
from __future__ import annotations

from dataclasses import dataclass, field
from typing import Any, Protocol


@dataclass(frozen=True)
class LLMStep:
    """Ответ модели: либо вызов инструмента, либо финальный результат."""

    tool: str | None = None
    args: dict[str, Any] = field(default_factory=dict)
    final: dict[str, Any] | None = None
    tokens: int = 0
    cost_rub: float = 0.0


class LLMProvider(Protocol):
    def next_step(self, task: str, history: list[dict[str, Any]]) -> LLMStep: ...


class ScriptedLLM:
    """Детерминированный мок: отдаёт заранее заданные шаги по порядку."""

    def __init__(self, steps: list[LLMStep]) -> None:
        self._steps = list(steps)
        self._i = 0

    def next_step(self, task: str, history: list[dict[str, Any]]) -> LLMStep:
        if self._i >= len(self._steps):
            raise RuntimeError("ScriptedLLM: сценарий закончился без final")
        step = self._steps[self._i]
        self._i += 1
        return step
