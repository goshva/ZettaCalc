"""Harness runtime: бюджеты, ретраи, anti-loop, трейсы.

Правило: детерминированное ядро не вызывает LLM. Harness — единственная
точка вызова LLM.
"""
from __future__ import annotations

from collections.abc import Callable
from dataclasses import dataclass, field
from typing import Any

from src.harness.guardrails import check_write_allowed
from src.harness.llm import LLMProvider

ANTI_LOOP_REPEATS = 3


@dataclass
class Budget:
    max_steps: int = 20
    max_tokens: int = 50_000
    max_cost_rub: float = 50.0


@dataclass
class Trace:
    steps: list[dict[str, Any]] = field(default_factory=list)
    stop_reason: str = "running"
    result: dict[str, Any] | None = None
    tokens: int = 0
    cost_rub: float = 0.0

    def log(
        self,
        tool: str,
        args: dict[str, Any],
        result: dict[str, Any],
        cost_rub: float,
    ) -> None:
        self.steps.append(
            {"tool": tool, "args": args, "result": result, "cost_rub": cost_rub}
        )


@dataclass(frozen=True)
class Tool:
    fn: Callable[..., dict[str, Any]]
    writes: bool = False


@dataclass
class WritePolicy:
    allow_write: bool = False
    approved: bool = False


def run_agent(
    task: str,
    budget: Budget,
    llm: LLMProvider,
    tools: dict[str, Tool],
    policy: WritePolicy | None = None,
) -> Trace:
    """Цикл агента. Останавливается по final, бюджету или anti-loop."""
    policy = policy or WritePolicy()
    trace = Trace()
    seen: dict[str, int] = {}

    while True:
        if len(trace.steps) >= budget.max_steps:
            trace.stop_reason = "budget_steps"
            return trace

        step = llm.next_step(task, trace.steps)
        trace.tokens += step.tokens
        trace.cost_rub += step.cost_rub
        if trace.tokens > budget.max_tokens:
            trace.stop_reason = "budget_tokens"
            return trace
        if trace.cost_rub > budget.max_cost_rub:
            trace.stop_reason = "budget_cost"
            return trace

        if step.final is not None:
            trace.result = step.final
            trace.stop_reason = "final"
            return trace

        name = step.tool or ""
        key = f"{name}:{sorted(step.args.items())}"
        seen[key] = seen.get(key, 0) + 1
        if seen[key] >= ANTI_LOOP_REPEATS:
            trace.stop_reason = "anti_loop"
            return trace

        tool = tools.get(name)
        if tool is None:
            trace.log(name, step.args, {"error": "unknown_tool"}, step.cost_rub)
            continue
        if tool.writes and not check_write_allowed(policy.allow_write, policy.approved):
            trace.log(name, step.args, {"error": "write_denied"}, step.cost_rub)
            continue
        trace.log(name, step.args, tool.fn(**step.args), step.cost_rub)
