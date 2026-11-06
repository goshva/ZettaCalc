"""Эталон поведения pydantic v2 для полей формы записи (mock API).

Значения прогоняются через ту же модель, что в mock API, поэтому результат
служит эталоном для Go-реализации: pydantic-core у mock и у скрипта один.
Режим `dump` пишет JSON-фикстуру для тестов Go.

Запуск:
    python backend/scripts/_ref.py dates
    python backend/scripts/_ref.py dump <путь>
"""
from __future__ import annotations

import json
import sys
from datetime import date
from typing import Any

from pydantic import BaseModel, Field, ValidationError

TIME_OPTIONS = ["Уточнить по телефону", "Первая половина дня", "Вторая половина дня"]
MAX_PICKED_IDEAS = 5
MAX_FILES = 10


class AppointmentRequest(BaseModel):
    name: str = Field(min_length=1, max_length=80)
    phone: str = Field(min_length=10, max_length=22)
    tasks: list[str] = Field(default_factory=list)
    preferred_date: date | None = None
    preferred_time: str = TIME_OPTIONS[0]
    ideas: list[str] = Field(default_factory=list, max_length=MAX_PICKED_IDEAS)
    files: list[str] = Field(default_factory=list, max_length=MAX_FILES)
    wish: str = Field(default="", max_length=1200)
    consent: bool = False


BASE: dict[str, Any] = {"name": "И", "phone": "1234567890", "consent": True}

# Набор значений для поля preferred_date: строки, числа, «мусор», границы.
DATE_CASES: list[Any] = [
    "", " ", "0", "-0", "+0", "1", "-1", "1.0", "0.0", "-0.0", "0.", ".5", "5.",
    "+1.5", "1,5", ".", "+", "-", "1..2", "1.5.5", "007", "0000", "00000",
    "0000000000", "1_0", "1 000", " 5 ", "5 ", "86400", "-86400", "-86401",
    "172800", "-172800", "4102444800", "253402214400", "253402300799",
    "253402300800", "2147483648", "20000000000", "20000000001", "-20000000000",
    "-20000000001", "99999999999", "100000000000", "999999999999",
    "999999999999999999", "1000000000000000000", "9223372036854775806",
    "9223372036854775807", "9223372036854775808", "-922337203685477580",
    "-9223372036854775807", "-9223372036854775808", "-9223372036854775809",
    "1e3", "1E3", "1e0", "0e0", "1.5e3", ".5e3", "5.e3", "1.5E3", "1.5e-3",
    "1.5e+3", "0.5e1", "1.5e", "1.5e3.5", "1e18", "1.0e18", "253402300800.0",
    "20261106", "20261106.5", "не дата", "н", "нннн", "ннннн", "не да",
    "не дат", "не ", "éééééé", "x", "abcd", "abcdefghi", "abcdefghij",
    "abcd-11-06", "a234-11-06", "abcd-", "2026-", "2026-1", "2026-11-",
    "2026-11-0", "2026-11-6", "2026-1-1", "2026/11/06", "2026.11.06",
    "2026-11-06", "2026-11-06 ", "2026-11-06X", "2026-11-06T", "2026-11-06T25:00",
    "2026-11-06T25:00:00", "2026-11-06T23:60", "2026-11-06T23:59:60",
    "2026-11-06T00:00", "2026-11-06T00:00:00", "2026-11-06T00:00:00Z",
    "2026-11-06T00:00:00z", "2026-11-06T00:00:00+03:00",
    "2026-11-06T00:00:00+0300", "2026-11-06T00:00:00-03:00",
    "2026-11-06T00:00:00+3:00", "2026-11-06T00:00:00+03", "2026-11-06T00:00:00+",
    "2026-11-06T00:00:00+03:60", "2026-11-06T00:00:00+24:00",
    "2026-11-06T00:00:00+23:59", "2026-11-06T00:00:00.000000",
    "2026-11-06T00:00:00.0000001", "2026-11-06T00:00:00.1",
    "2026-11-06T00:00:00,5", "2026-11-06T00:00:00.", "2026-11-06T00:00:00é",
    "2026-11-06_00:00:00", "2026-11-06 00:00:00", "2026-11-06t00:00:00",
    "2026-11-06T12:00:00", "2026-11-06T00:00:01", "0000-01-01", "0001-01-01",
    "9999-12-31", "9999-12-31T00:00:00", "2026-13-01", "2026-00-01", "2026-01-00",
    "2026-02-30", "2026-02-29", "2024-02-29", "2026-04-31", "2026-11-31",
    5, 5.0, 0, 0.0, -0.0, 86400, 86400.0, 253402300800, 253402300799,
    True, False, None, [], {}, {"a": 1}, [1], 1e18, -1e18, float("inf"),
    float("-inf"),
]

# Набор значений для строковых полей и списков.
FIELD_CASES: list[tuple[str, Any]] = [
    ("name", None), ("name", 5), ("name", True), ("name", []), ("name", {}),
    ("name", ""), ("name", " "), ("name", "я" * 80), ("name", "я" * 81),
    ("name", "a" * 80), ("name", "a" * 81),
    ("phone", None), ("phone", 5), ("phone", "1" * 22), ("phone", "1" * 23),
    ("phone", "1" * 10), ("phone", "1" * 9),
    ("preferred_time", None), ("preferred_time", 5), ("preferred_time", []),
    ("preferred_time", "Ночью"), ("preferred_time", ""),
    ("wish", None), ("wish", 5), ("wish", []), ("wish", {}), ("wish", ""),
    ("wish", "я" * 1200), ("wish", "я" * 1201),
    ("tasks", None), ("tasks", 5), ("tasks", "Подбор интерьера"), ("tasks", {}),
    ("tasks", []), ("tasks", ["a"]), ("tasks", [1]), ("tasks", ["a", 1]),
    ("tasks", [None]), ("tasks", ["a", "b", "c", "d", "e", "f"]),
    ("ideas", None), ("ideas", 5), ("ideas", []), ("ideas", [1]),
    ("ideas", ["a"] * 5), ("ideas", ["a"] * 6), ("ideas", [1, 2, 3, 4, 5, 6]),
    ("files", None), ("files", 5), ("files", []), ("files", ["a"] * 10),
    ("files", ["a"] * 11), ("files", [1] * 11),
    ("consent", None), ("consent", "abc"), ("consent", 2), ("consent", -1),
    ("consent", 1), ("consent", 0), ("consent", 1.0), ("consent", 0.0),
    ("consent", "yes"), ("consent", "no"), ("consent", "on"), ("consent", "off"),
    ("consent", "t"), ("consent", "f"), ("consent", "TRUE"), ("consent", "y"),
    ("consent", "n"), ("consent", ""), ("consent", []), ("consent", {}),
    ("consent", "2"), ("consent", "0.0"),
]


def errors_of(field: str, value: Any) -> list[dict[str, Any]] | None:
    """Возвращает список ошибок pydantic или None, если значение принято."""
    try:
        AppointmentRequest(**{**BASE, field: value})
        return None
    except ValidationError as exc:
        out = []
        for err in exc.errors():
            item = {
                "type": err["type"],
                "loc": [str(x) for x in err["loc"]],
                "msg": err["msg"],
            }
            if "ctx" in err and err["ctx"] is not None:
                item["ctx"] = {k: str(v) for k, v in err["ctx"].items()}
            out.append(item)
        return out


def check(field: str, value: Any) -> str:
    errs = errors_of(field, value)
    if errs is None:
        return "OK"
    return "; ".join(f"{e['type']} {e['msg']!r} ctx={e.get('ctx')}" for e in errs)


def dump(path: str) -> None:
    payload = {
        "dates": [{"value": v, "errors": errors_of("preferred_date", v)}
                  for v in DATE_CASES],
        "fields": [{"field": f, "value": v, "errors": errors_of(f, v)}
                   for f, v in FIELD_CASES],
    }
    with open(path, "w", encoding="utf-8") as fh:
        json.dump(payload, fh, ensure_ascii=False, indent=1)
    print(f"записано: {path} (дат {len(DATE_CASES)}, полей {len(FIELD_CASES)})")


def main() -> None:
    kind = sys.argv[1] if len(sys.argv) > 1 else "dates"
    if kind == "dump":
        dump(sys.argv[2])
        return
    if kind == "dates":
        for value in DATE_CASES:
            print(f"date={value!r:34} -> {check('preferred_date', value)}")
        return
    for field, value in FIELD_CASES:
        print(f"{field}={value!r:24} -> {check(field, value)}")


if __name__ == "__main__":
    main()
