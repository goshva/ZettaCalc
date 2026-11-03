"""Evals: pass^k для автосборки ЦПР.

Правило: метрика — pass^k (все k прогонов успешны), не pass@k.
"""
from __future__ import annotations


def pass_hat_k(successes: list[bool]) -> bool:
    """Все прогоны успешны."""
    return all(successes) if successes else False


def pass_at_k(successes: list[bool]) -> bool:
    """Хотя бы один успешен (для сравнения, не для продакшна)."""
    return any(successes) if successes else False
