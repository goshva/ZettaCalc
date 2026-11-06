"""Сверка поведения Python-mock и Go-сервиса на нечисловых литералах JSON.

`json.loads` принимает литералы Infinity, -Infinity и NaN, хотя стандарт JSON
их не разрешает, а `json.dumps` без allow_nan на них падает. Из-за этого mock
API часть таких запросов принимает, а часть — отвечает 500 уже на сериализации
ответа. Скрипт проверяет, что Go-сервис повторяет это поведение.

Запуск: python backend/scripts/parity_nonfinite.py
"""

from __future__ import annotations

import json
import sys
import urllib.error
import urllib.parse
import urllib.request

MOCK = "http://127.0.0.1:8010"
GO = "http://127.0.0.1:8080"

CREDENTIALS = "username=test@zetta.ru&password=test123"

# ── Тела запросов с non-finite литералами ─────────────────────────────────
# Формат: (метод, путь, тело, нужен_токен, подпись)
CASES: list[tuple[str, str, bytes, bool, str]] = [
    ("POST", "/api/center/appointment/dry-run",
     b'{"name":"\xd0\x98","phone":"1234567890","consent":true,"tasks":[Infinity]}', False,
     "dry-run: Infinity в tasks"),
    ("POST", "/api/center/appointment/dry-run",
     b'{"name":"\xd0\x98","phone":"1234567890","consent":Infinity}', False,
     "dry-run: Infinity в consent"),
    ("POST", "/api/center/appointment/dry-run",
     b'{"name":"\xd0\x98","phone":"1234567890","consent":true,"x":Infinity}', False,
     "dry-run: Infinity в лишнем поле"),
    ("POST", "/api/center/appointment/dry-run",
     b'{"name":"\xd0\x98","phone":"1234567890","consent":true,"x":[NaN]}', False,
     "dry-run: NaN во вложенном массиве"),
    ("POST", "/api/center/appointment/dry-run",
     b'{"name":"\xd0\x98","phone":"1234567890","consent":true,"tasks":[1e400]}', False,
     "dry-run: переполнение 1e400 в tasks"),
    ("POST", "/api/center/appointment/dry-run",
     b'{"name":"\xd0\x98","phone":"1234567890","consent":true,"wish":1e400}', False,
     "dry-run: 1e400 в строковом поле"),
    ("POST", "/api/center/appointment/dry-run",
     b'{"name":"\xd0\x98","phone":"1234567890","consent":true,"preferred_date":Infinity}', False,
     "dry-run: Infinity в preferred_date"),
    ("POST", "/api/center/appointment/dry-run",
     b'{"name":"\xd0\x98","phone":"1234567890","consent":true,"x":-Infinity}', False,
     "dry-run: -Infinity в лишнем поле"),
    ("POST", "/api/center/appointment/dry-run", b'Infinity', False,
     "dry-run: тело целиком Infinity"),
    ("POST", "/api/center/appointment/dry-run", b'[NaN]', False,
     "dry-run: тело целиком массив с NaN"),
    ("POST", "/api/center/appointment/dry-run",
     b'{"name":"\xd0\x98","phone":"1234567890","consent":true,"x":nan}', False,
     "dry-run: строчная nan — ошибка разбора"),
    ("POST", "/api/center/appointment/dry-run",
     b'{"name":"\xd0\x98","phone":"1234567890","consent":true,"x":Infinity1}', False,
     "dry-run: Infinity1 — ошибка разбора"),
    ("POST", "/api/center/appointment/dry-run",
     b'{"name":"","phone":"1234567890","consent":true,"x":Infinity}', False,
     "dry-run: доменная ошибка важнее non-finite"),
    ("POST", "/api/center/appointment/dry-run",
     b'{"name":"\xd0\x98","phone":"1234567890","consent":1.5,"x":Infinity}', False,
     "dry-run: bool_type для 1.5"),
    ("POST", "/api/center/appointment/dry-run",
     b'{"name":"\xd0\x98","phone":"1234567890","consent":1e400}', False,
     "dry-run: bool_type для 1e400"),
    ("POST", "/api/center/appointment/dry-run",
     b'{"name":"\xd0\x98","phone":"1234567890","consent":0.5}', False,
     "dry-run: bool_type для 0.5"),
    ("POST", "/api/center/appointment/dry-run",
     b'{"name":"\xd0\x98","phone":"1234567890","consent":2}', False,
     "dry-run: bool_parsing для 2"),
    ("POST", "/api/center/appointment/dry-run",
     b'{"name":"\xd0\x98","phone":"1234567890","consent":9223372036854775808}', False,
     "dry-run: bool_type для int вне int64"),
    ("POST", "/api/center/appointment/dry-run",
     b'{"name":"\xd0\x98","phone":"1234567890","consent":9.223372036854776e18}', False,
     "dry-run: bool_type для float вне int64"),
    ("POST", "/api/cpr/generate", b'{"project_id":Infinity}', True,
     "cpr: Infinity в project_id"),
    ("POST", "/api/cpr/generate", b'{"project_id":1e400}', True,
     "cpr: 1e400 в project_id"),
    ("POST", "/api/cpr/generate", b'{"project_id":[Infinity]}', True,
     "cpr: Infinity внутри project_id"),
    ("POST", "/api/cpr/generate", b'{"project_id":"p","extra":Infinity}', True,
     "cpr: Infinity в лишнем поле"),
    ("POST", "/api/cart/items", b'{"price":Infinity}', True,
     "cart: Infinity в price"),
    ("POST", "/api/cart/items", b'{"price":1,"qty":NaN}', True,
     "cart: NaN в qty"),
    ("POST", "/api/cart/items", b'{"extra":Infinity}', True,
     "cart: Infinity в лишнем поле"),
    ("POST", "/api/cart/items", b'{"qty":1e400}', True,
     "cart: 1e400 в qty"),
]

failures: list[str] = []
checks = 0


def token(base: str) -> str:
    """Получает токен демо-пользователя у указанного сервиса."""
    data = urllib.parse.urlencode({"username": "test@zetta.ru", "password": "test123"}).encode()
    req = urllib.request.Request(base + "/api/auth/login", data=data,
                                 headers={"Content-Type": "application/x-www-form-urlencoded"},
                                 method="POST")
    with urllib.request.urlopen(req, timeout=30) as resp:
        return json.loads(resp.read())["access_token"]


def call(base: str, path: str, body: bytes, bearer: str | None):
    """Отправляет запрос и возвращает (код, content-type, тело)."""
    headers = {"Content-Type": "application/json"}
    if bearer:
        headers["Authorization"] = "Bearer " + bearer
    req = urllib.request.Request(base + path, data=body, headers=headers, method="POST")
    try:
        with urllib.request.urlopen(req, timeout=30) as resp:
            return resp.status, resp.headers.get("Content-Type", ""), resp.read()
    except urllib.error.HTTPError as exc:
        return exc.code, exc.headers.get("Content-Type", ""), exc.read()
    except Exception as exc:  # noqa: BLE001
        return -1, "", str(exc).encode()


def main() -> int:
    global checks
    mock_token = token(MOCK)
    go_token = token(GO)

    for _method, path, body, need_token, label in CASES:
        checks += 1
        m = call(MOCK, path, body, mock_token if need_token else None)
        g = call(GO, path, body, go_token if need_token else None)
        if m != g:
            failures.append(label)
            print(f"DIFF {label}")
            print(f"  mock: {m[0]} {m[1]} {m[2][:160]!r}")
            print(f"   go : {g[0]} {g[1]} {g[2][:160]!r}")

    print(f"\nпроверок: {checks}, расхождений: {len(failures)}")
    return 1 if failures else 0


if __name__ == "__main__":
    sys.exit(main())
