"""Code-based guardrails: ACL, JSON, SKU, площади, write-safety.

Правило: права доступа проверяются только code-based, не LLM.
"""
from __future__ import annotations


def check_acl(actor_roles: list[str], required: str) -> bool:
    return required in actor_roles


def check_sku_exists(sku: str, known_skus: set[str]) -> bool:
    return sku in known_skus


def check_write_allowed(allow_write: bool, approved: bool) -> bool:
    """Write-операция допустима только при allow_write и approval."""
    return allow_write and approved
