from src.rules.engine import Surface, load_techcards, select_techcard

CARDS = load_techcards()


def test_card_selected_for_even_base():
    sel = select_techcard(Surface("floor", "keramogranit", 12.8, 2), CARDS)
    assert sel.techcard is not None and sel.techcard["code"] == "FLOOR-KER-600"
    assert not sel.needs_review


def test_uneven_base_is_not_silently_ignored():
    sel = select_techcard(Surface("floor", "keramogranit", 12.8, 8), CARDS)
    assert sel.techcard is None and sel.needs_review
    assert "неровность" in sel.reasons[0]


def test_unknown_base_becomes_assumption():
    sel = select_techcard(Surface("floor", "keramogranit", 3.4, None), CARDS)
    assert sel.techcard is not None
    assert sel.assumptions and "требует проверки" in sel.assumptions[0]


def test_no_card_means_review():
    sel = select_techcard(Surface("wall", "marble", 10, 1), CARDS)
    assert sel.techcard is None and sel.needs_review
