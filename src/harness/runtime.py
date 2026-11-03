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
