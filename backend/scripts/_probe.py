"""Черновой зонд: показывает сырые ответы mock и Go на один и тот же запрос.

Запуск: python backend/scripts/_probe.py
"""
from __future__ import annotations

import json
import sys
import urllib.error
import urllib.request

MOCK = "http://127.0.0.1:8010"
GO = "http://127.0.0.1:8080"


def call(base, path, method="GET", raw=None, headers=None, ct="application/json"):
    hdrs = dict(headers or {})
    if raw is not None and ct:
        hdrs["Content-Type"] = ct
    req = urllib.request.Request(base + path, data=raw, headers=hdrs, method=method)
    try:
        with urllib.request.urlopen(req, timeout=30) as resp:
            return resp.status, resp.read()
    except urllib.error.HTTPError as e:
        return e.code, e.read()
    except Exception as e:  # noqa: BLE001
        return -1, str(e).encode()


def show(label, path, raw=None, method="POST", headers=None, ct="application/json"):
    s1, b1 = call(MOCK, path, method, raw, headers, ct)
    s2, b2 = call(GO, path, method, raw, headers, ct)
    mark = "OK  " if (s1, b1) == (s2, b2) else "DIFF"
    print(f"[{mark}] {label}")
    print(f"   mock {s1}: {b1.decode('utf-8', 'replace')}")
    print(f"     go {s2}: {b2.decode('utf-8', 'replace')}")


BASE = {"name": "Иван", "phone": "1234567890", "consent": True}


def dry(body):
    return json.dumps({**BASE, **body}, ensure_ascii=False).encode()


if __name__ == "__main__":
    which = sys.argv[1] if len(sys.argv) > 1 else "dates"
    if which == "dates":
        for value in ["", "1", "202", "2026", "2026-", "2026-1", "2026-11-",
                      "2026-11-0", "не дата", "20261106", "2026-11-06X", "x",
                      "12345678", "2026-11-06T25:00", "0", "-1"]:
            show(f"date={value!r}", "/api/center/appointment/dry-run",
                 dry({"preferred_date": value}))
    elif which == "types":
        for value in [None, 5, "abc", True, [], {"a": 1}]:
            show(f"preferred_time={value!r}", "/api/center/appointment/dry-run",
                 dry({"preferred_time": value}))
            show(f"wish={value!r}", "/api/center/appointment/dry-run",
                 dry({"wish": value}))
    elif which == "login":
        show("login ok", "/api/auth/login",
             b"username=test@zetta.ru&password=test123", ct="application/x-www-form-urlencoded")
    elif which == "cart":
        raw = b"username=test@zetta.ru&password=test123"
        _, body = call(MOCK, "/api/auth/login", "POST", raw,
                       ct="application/x-www-form-urlencoded")
        tok = json.loads(body)["access_token"]
        _, body = call(GO, "/api/auth/login", "POST", raw,
                       ct="application/x-www-form-urlencoded")
        tok2 = json.loads(body)["access_token"]
        show("cart", "/api/cart", None, "GET", {"Authorization": f"Bearer {tok}"})
        show("cart/items null", "/api/cart/items", b'{"price":null,"qty":2}',
             headers={"Authorization": f"Bearer {tok}"})
        s1, b1 = call(MOCK, "/api/cart/items", "POST", b'{"price":null,"qty":2}',
                      {"Authorization": f"Bearer {tok}"})
        s2, b2 = call(GO, "/api/cart/items", "POST", b'{"price":null,"qty":2}',
                      {"Authorization": f"Bearer {tok2}"})
        print(f"   mock {s1}: {b1!r}\n     go {s2}: {b2!r}")
    elif which == "create":
        for path, raw, hdrs in [
            ("/api/center/appointment?approved=2", b"{oops", None),
            ("/api/center/appointment?approved=2", b"[]", None),
            ("/api/center/appointment?approved=2", b"", None),
        ]:
            show(f"{path} {raw!r}", path, raw, headers=hdrs)
        key = "probe-echo-1"
        body = dry({})
        show("create ok", f"/api/center/appointment?approved=true", body,
             headers={"Idempotency-Key": key})
        show("create replay другой телом", "/api/center/appointment?approved=true",
             dry({"consent": False}), headers={"Idempotency-Key": key})
