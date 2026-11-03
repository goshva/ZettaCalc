"""Контракты: OpenAPI валиден и согласован с JSON Schema ЦПР."""
from pathlib import Path

import jsonschema
import pytest
import yaml
from openapi_spec_validator import validate

from src.domain.cpr import RoomInput, build_cpr
from src.rules.engine import load_techcards
from tests.conftest import FIXTURES, NOW, load_json

ROOT = Path(__file__).parents[2]
API = yaml.safe_load((ROOT / "schemas" / "openapi" / "cpr.yaml").read_text(encoding="utf-8"))
CPR_SCHEMA = load_json(ROOT / "schemas" / "json" / "cpr.schema.json")
WRITE_METHODS = {"post", "put", "patch", "delete"}


def test_openapi_is_valid():
    validate(API)


def test_has_real_paths():
    assert len(API["paths"]) >= 10


def test_all_writes_have_idempotency_key():
    exempt = {"/files"}  # загрузка файла: ключ — hash содержимого
    for path, item in API["paths"].items():
        for method, op in item.items():
            if method in WRITE_METHODS and path not in exempt:
                refs = [p.get("$ref") for p in op.get("parameters", [])]
                assert "#/components/parameters/IdempotencyKey" in refs, f"{method} {path}"


def test_cpr_status_enum_matches_json_schema():
    api_enum = API["components"]["schemas"]["Cpr"]["properties"]["status"]["enum"]
    assert api_enum == CPR_SCHEMA["properties"]["status"]["enum"]


def test_generated_cpr_satisfies_api_cpr_component(catalog):
    case = load_json(FIXTURES / "cpr" / "kitchen_small.json")
    room = RoomInput(
        case["room_id"], case["type"], [tuple(p) for p in case["contour_mm"]],
        case["ceiling_height_m"], case["floor_material"],
        case["floor_unevenness_mm_per_m"], case["completeness_checklist"],
    )
    cpr = build_cpr(cpr_seq=1, year=2026, project_id="P", room=room, catalog=catalog,
                    cards=load_techcards(), now=NOW)
    jsonschema.validate(cpr, API["components"]["schemas"]["Cpr"])


@pytest.mark.parametrize("name", ["kitchen_full", "bathroom_minimal"])
def test_examples_validate(name):
    jsonschema.validate(load_json(ROOT / "examples" / f"cpr_{name}.json"), CPR_SCHEMA)
