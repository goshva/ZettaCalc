"""Code-based evals автосборки ЦПР: схема, геометрия, SKU, pass^k."""
import json
from pathlib import Path

import jsonschema
import pytest

from src.domain.cpr import RoomInput, build_cpr
from src.domain.geometry import Contour
from src.harness.evals import pass_hat_k, run_k
from src.rules.engine import load_techcards
from tests.conftest import FIXTURES, NOW, load_json

SCHEMA = load_json(Path(__file__).parents[2] / "schemas" / "json" / "cpr.schema.json")
CASES = sorted((FIXTURES / "cpr").glob("*.json"))
K = 10


def build(case: dict, catalog):
    room = RoomInput(
        room_id=case["room_id"],
        type=case["type"],
        contour_mm=[tuple(p) for p in case["contour_mm"]],
        ceiling_height_m=case["ceiling_height_m"],
        floor_material=case["floor_material"],
        floor_unevenness_mm_per_m=case["floor_unevenness_mm_per_m"],
        completeness_checklist=case["completeness_checklist"],
    )
    return build_cpr(
        cpr_seq=1,
        year=2026,
        project_id=case["project_id"],
        room=room,
        catalog=catalog,
        cards=load_techcards(),
        now=NOW,
    )


@pytest.mark.parametrize("path", CASES, ids=lambda p: p.stem)
def test_cpr_matches_schema(path, catalog):
    jsonschema.validate(build(load_json(path), catalog), SCHEMA)


@pytest.mark.parametrize("path", CASES, ids=lambda p: p.stem)
def test_geometry_consistency(path, catalog):
    case = load_json(path)
    cpr = build(case, catalog)
    expected = Contour([tuple(p) for p in case["contour_mm"]]).area_m2()
    assert cpr["rooms"][0]["area_m2"] == pytest.approx(expected, abs=0.01)
    assert cpr["rooms"][0]["area_m2"] == case["expect"]["area_m2"]


@pytest.mark.parametrize("path", CASES, ids=lambda p: p.stem)
def test_every_sku_is_from_catalog(path, catalog):
    cpr = build(load_json(path), catalog)
    for m in cpr["specification"]["materials"]:
        assert m["sku"] in catalog.known_skus()


@pytest.mark.parametrize("path", CASES, ids=lambda p: p.stem)
def test_status_and_assumptions(path, catalog):
    case = load_json(path)
    cpr = build(case, catalog)
    assert cpr["status"] == case["expect"]["status"]
    if case["expect"].get("has_assumption"):
        assert cpr["client"]["assumptions"]


@pytest.mark.parametrize("path", CASES, ids=lambda p: p.stem)
def test_pass_hat_k_deterministic(path, catalog):
    case = load_json(path)
    reference = json.dumps(build(case, catalog), sort_keys=True)
    runs = run_k(lambda: json.dumps(build(case, catalog), sort_keys=True) == reference, K)
    assert len(runs) == K and pass_hat_k(runs)


def test_uneven_base_keeps_mandatory_operations_out_not_silently(catalog):
    cpr = build(load_json(FIXTURES / "cpr" / "kitchen_uneven.json"), catalog)
    assert cpr["rooms"][0]["works"] == []
    assert any("неровность" in a for a in cpr["client"]["assumptions"])


def test_techcard_example_matches_engine(catalog):
    ex = load_json(FIXTURES / "techcards" / "floor_ker_600.json")
    case = load_json(FIXTURES / "cpr" / "kitchen_small.json")
    cpr = build(case, catalog)
    ops = [w["code"] for w in cpr["rooms"][0]["works"]]
    assert ops == ex["expect"]["operations"]
    qty = {m["sku"]: m["qty"] for m in cpr["specification"]["materials"]}
    assert qty["SKU-GLUE-C2-25"] == ex["expect"]["glue_kg"]
    assert qty["SKU-GROUT-5"] == ex["expect"]["grout_kg"]
