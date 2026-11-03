from src.harness.evals import pass_hat_k, pass_at_k


def test_pass_hat_k_all():
    assert pass_hat_k([True, True, True])
    assert not pass_hat_k([True, False, True])


def test_pass_at_k_any():
    assert pass_at_k([False, True, False])
