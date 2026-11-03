import pytest

from src.harness.guardrails import check_acl, check_sku_exists, check_write_allowed


@pytest.mark.parametrize("roles", [[], ["client"], ["contractor"], ["Expert"]])
def test_acl_cannot_be_bypassed(roles):
    assert not check_acl(roles, "expert")


def test_acl_allows_expert():
    assert check_acl(["client", "expert"], "expert")


def test_sku_must_exist():
    assert not check_sku_exists("MADE-UP", {"A"})


@pytest.mark.parametrize("allow,approved", [(False, False), (False, True), (True, False)])
def test_write_needs_both(allow, approved):
    assert not check_write_allowed(allow, approved)
