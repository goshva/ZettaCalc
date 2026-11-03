from src.harness.guardrails import check_acl, check_write_allowed


def test_acl_denied():
    assert not check_acl(["client"], "expert")


def test_write_requires_approval():
    assert not check_write_allowed(allow_write=True, approved=False)
    assert check_write_allowed(allow_write=True, approved=True)
