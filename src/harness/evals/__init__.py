"""Evals: pass^k для автосборки ЦПР.

Правило: метрика — pass^k (все k прогонов успешны), не pass@k.
"""
from __future__ import annotations

from collections.abc import Callable


def pass_hat_k(successes: list[bool]) -> bool:
    """Все прогоны успешны."""
    return all(successes) if successes else False


def pass_at_k(successes: list[bool]) -> bool:
    """Хотя бы один успешен (для сравнения, не для продакшна)."""
    return any(successes) if successes else False


def run_k(fn: Callable[[], object], k: int) -> list[bool]:
    """Прогнать проверку k раз; исключение считается провалом."""
    results: list[bool] = []
    for _ in range(k):
        try:
            results.append(bool(fn()))
        except Exception:
            results.append(False)
    return results
