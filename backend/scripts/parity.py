"""Сравнение ответов Python-mock и Go-сервиса по всем маршрутам.

Запуск:  python backend/scripts/parity.py
Отчёт:   расхождения по коду ответа, content-type и телу (байт-в-байт).
"""

from __future__ import annotations

import hashlib
import json
import sys
import urllib.error
import urllib.request

MOCK = "http://127.0.0.1:8010"
GO = "http://127.0.0.1:8080"

# Публичные маршруты без параметров.
PATHS = [
    "/health",
    "/api/health",
    "/api/design/variants",
    "/api/design/variants/variant_01/estimate",
    "/api/design/variants/variant_01/cpr",
    "/api/design/variants/variant_01/bids",
    "/api/design/variants/variant_04/bids",
    "/api/interiors",
    "/api/interiors/categories",
    "/api/interiors/descriptions",
    "/api/interiors/idea0",
    "/api/catalog/products",
    "/api/catalog/products?group=1",
    "/api/catalog/rooms",
    "/api/catalog/rooms/room0",
    "/api/center/interiors",
    "/api/center/photos",
    "/api/hotspot/data",
    "/api/hotspot/analogues",
    "/api/main-page/start",
    "/api/main-page/how",
    "/api/main-page/ideas",
    "/api/main-page/connected",
    "/api/references",
    "/api/references/ref_001",
    "/api/techcards",
    "/api/techcards?surface=wall",
    "/api/estimate/est_001",
    "/api/cpr/cpr_001",
    "/api/bids",
    "/api/planners",
    "/api/center/appointment/options",
    "/api/apps",
    "/api/apps/main/css",
    "/api/apps/interiors/css",
    "/api/apps/interiors/html",
    "/api/apps/about/html",
    "/api/page",
    "/api/page/a/header.html",
    "/api/page/b/header.html",
    "/api/css/a",
    "/api/css/b",
    "/api/css/a/shell",
    "/api/css/b/shell",
    "/api/images",
    "/api/images/idea-kitchen",
    "/api/images/hero",
    "/api/images/file/a_main_img_0003.jpg",
    "/api/images/file/b_img_0001.jpg",
    # ошибки
    "/api/images/nope",
    "/api/images/file/nope.jpg",
    "/api/apps/nope/css",
    "/api/interiors/nope",
    "/api/estimate/nope",
    "/api/css/c",
    "/api/page/c/x.html",
]

# Методы с телом: POST (публичные заглушки).
POSTS = [
    ("/api/estimate/calculate", None),
    ("/api/cpr/generate", {"project_id": "proj_001"}),
]


def fetch(base: str, path: str, method: str = "GET", body: dict | None = None):
    url = base + path
    data = None
    headers = {}
    if method == "POST":
        data = json.dumps(body if body is not None else {}, ensure_ascii=False).encode()
        headers["Content-Type"] = "application/json"
    req = urllib.request.Request(url, data=data, headers=headers, method=method)
    try:
        with urllib.request.urlopen(req, timeout=30) as resp:
            return resp.status, resp.headers.get("Content-Type", ""), resp.read()
    except urllib.error.HTTPError as e:
        return e.code, e.headers.get("Content-Type", ""), e.read()
    except Exception as e:  # noqa: BLE001
        return -1, "", str(e).encode()


def discover() -> list[str]:
    """Дополняет список маршрутов всеми фрагментами, ключами и файлами картинок."""
    paths: list[str] = []

    status, _, body = fetch(MOCK, "/api/page")
    if status == 200:
        listing = json.loads(body)
        for proto in ("a", "b"):
            for name in listing[proto]["files"]:
                paths.append(f"/api/page/{proto}/{name}")

    status, _, body = fetch(MOCK, "/api/images")
    if status == 200:
        for key in json.loads(body)["keys"]:
            paths.append(f"/api/images/{key}")

    # все файлы картинок сущностей: берём имена из индекса ключей и данных
    status, _, body = fetch(MOCK, "/api/interiors")
    if status == 200:
        for item in json.loads(body):
            image = item.get("image", "")
            if "/api/images/file/" in image:
                paths.append(image)

    status, _, body = fetch(MOCK, "/api/catalog/rooms")
    if status == 200:
        for room in json.loads(body):
            image = room.get("image", "")
            if "/api/images/file/" in image:
                paths.append(image)

    return paths


def digest(body: bytes) -> str:
    return hashlib.sha256(body).hexdigest()[:12]


# Ожидаемые отличия в /health: Go-сервис не является mock (флаг mock=false), а
# счётчик appointments растёт в mock при каждой созданной записи и равен числу
# строк в БД у Go — сравнивать его между сервисами нельзя.
EXPECTED = {"/health", "/api/health"}

# Поля, зависящие от состояния процесса, а не от контракта API.
VOLATILE_FIELDS = ("mock", "appointments")


def strip_expected(path: str, body: bytes) -> bytes:
    if path not in EXPECTED:
        return body
    try:
        data = json.loads(body)
    except ValueError:
        return body
    for field in VOLATILE_FIELDS:
        data.pop(field, None)
    return json.dumps(data, ensure_ascii=False, separators=(",", ":")).encode()


def main() -> int:
    bad = 0
    total = 0
    for path in PATHS + discover():
        total += 1
        s1, c1, b1 = fetch(MOCK, path)
        s2, c2, b2 = fetch(GO, path)
        same = (s1 == s2) and (strip_expected(path, b1) == strip_expected(path, b2))
        if not same:
            bad += 1
            print(f"DIFF {path}")
            print(f"  mock: {s1} {c1} len={len(b1)} sha={digest(b1)}")
            print(f"   go : {s2} {c2} len={len(b2)} sha={digest(b2)}")
            if len(b1) < 400 and len(b2) < 400:
                print(f"  mock body: {b1[:300]!r}")
                print(f"   go  body: {b2[:300]!r}")
    for path, body in POSTS:
        total += 1
        s1, _, b1 = fetch(MOCK, path, "POST", body)
        s2, _, b2 = fetch(GO, path, "POST", body)
        if s1 != s2 or strip_expected(path, b1) != strip_expected(path, b2):
            bad += 1
            print(f"DIFF POST {path}")
            print(f"  mock: {s1} len={len(b1)} sha={digest(b1)}")
            print(f"   go : {s2} len={len(b2)} sha={digest(b2)}")
            if len(b1) < 400 and len(b2) < 400:
                print(f"  mock body: {b1[:300]!r}")
                print(f"   go  body: {b2[:300]!r}")
    print(f"\nвсего маршрутов: {total}, расхождений: {bad}")
    return 1 if bad else 0


if __name__ == "__main__":
    sys.exit(main())
