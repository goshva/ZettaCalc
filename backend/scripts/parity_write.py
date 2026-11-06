"""Сравнение write-эндпоинтов, авторизации и валидации Python-mock и Go-сервиса.

Токены, идентификаторы записей и время создания различаются, поэтому
сравниваются коды, тексты ошибок и структура ответа. Всё остальное (включая
порядок ключей) должно совпадать байт-в-байт: паритет — цель миграции.

Запуск: python backend/scripts/parity_write.py
"""

from __future__ import annotations

import hashlib
import json
import re
import sys
import urllib.error
import urllib.request
import uuid

MOCK = "http://127.0.0.1:8010"
GO = "http://127.0.0.1:8080"

CREDENTIALS = "username=test@zetta.ru&password=test123"
BAD_CREDENTIALS = "username=test@zetta.ru&password=wrong"

APPOINTMENT = {
    "name": "Иван",
    "phone": "+7 (999) 123-45-67",
    "tasks": ["Подбор интерьера"],
    "preferred_time": "Уточнить по телефону",
    "ideas": [],
    "files": [],
    "wish": "Хочу светлую кухню",
    "consent": True,
}

# Идентификатор записи, время и токен зависят от запуска.
VOLATILE_ID = re.compile(rb'"id":"appt_[0-9a-f]{10}"')
VOLATILE_TIME = re.compile(rb'"created_at":"[^"]*"')
VOLATILE_TOKEN = re.compile(rb'"access_token":"[^"]*"')
VOLATILE_CPR = re.compile(rb'"id":"cpr_[0-9a-f]{8}"')


failures: list[str] = []
checks = 0


def normalize(body: bytes, path: str) -> bytes:
    """Убирает изменчивые части ответа, сохраняя всё остальное байт-в-байт."""
    body = VOLATILE_ID.sub(b'"id":"appt_XXX"', body)
    body = VOLATILE_TIME.sub(b'"created_at":"TIME"', body)
    body = VOLATILE_TOKEN.sub(b'"access_token":"TOKEN"', body)
    body = VOLATILE_CPR.sub(b'"id":"cpr_XXX"', body)
    return body


def call(base, path, method="GET", body=None, headers=None, form=False, raw=None,
         content_type="application/json"):
    data = raw
    hdrs = dict(headers or {})
    if body is not None:
        if form:
            data = body.encode()
            hdrs["Content-Type"] = "application/x-www-form-urlencoded"
        else:
            data = json.dumps(body, ensure_ascii=False).encode()
            hdrs["Content-Type"] = content_type
    elif raw is not None and content_type:
        hdrs["Content-Type"] = content_type
    req = urllib.request.Request(base + path, data=data, headers=hdrs, method=method)
    try:
        with urllib.request.urlopen(req, timeout=30) as resp:
            return resp.status, resp.read()
    except urllib.error.HTTPError as e:
        return e.code, e.read()
    except Exception as e:  # noqa: BLE001
        return -1, str(e).encode()


def token(base, credentials=CREDENTIALS):
    status, body = call(base, "/api/auth/login", "POST", credentials, form=True)
    if status != 200:
        raise SystemExit(f"логин не удался: {status} {body[:200]!r}")
    return json.loads(body)["access_token"]


def compare(label, path, **kwargs):
    """Сравнивает ответы mock и Go по коду и нормализованному телу."""
    global checks
    checks += 1
    s1, b1 = call(MOCK, path, **kwargs)
    s2, b2 = call(GO, path, **kwargs)
    n1, n2 = normalize(b1, path), normalize(b2, path)
    if s1 == s2 and n1 == n2:
        return True
    failures.append(label)
    print(f"DIFF {label} ({path})")
    print(f"  mock: {s1} {n1[:400]!r}")
    print(f"   go : {s2} {n2[:400]!r}")
    return False


def compare_shape(label, path, **kwargs):
    """Сравнивает код и набор ключей верхнего уровня (для записей с id)."""
    global checks
    checks += 1
    s1, b1 = call(MOCK, path, **kwargs)
    s2, b2 = call(GO, path, **kwargs)
    j1 = json.loads(b1) if b1 else {}
    j2 = json.loads(b2) if b2 else {}
    if s1 == s2 and sorted(j1) == sorted(j2):
        return True
    failures.append(label)
    print(f"DIFF {label} ({path})")
    print(f"  mock: {s1} {json.dumps(j1, ensure_ascii=False)[:300]}")
    print(f"   go : {s2} {json.dumps(j2, ensure_ascii=False)[:300]}")
    return False


def login_cases():
    print("== авторизация ==")
    compare("login: верный пароль", "/api/auth/login", method="POST",
            body=CREDENTIALS, form=True)
    compare("login: неверный пароль", "/api/auth/login", method="POST",
            body=BAD_CREDENTIALS, form=True)
    compare("login: пустая форма", "/api/auth/login", method="POST",
            body="", form=True)
    compare("login: только username", "/api/auth/login", method="POST",
            body="username=test@zetta.ru", form=True)
    compare("login: пустой username", "/api/auth/login", method="POST",
            body="username=&password=test123", form=True)
    compare("login: GET", "/api/auth/login", method="GET")

    for base, label in ((MOCK, "mock"), (GO, "go")):
        t = token(base)
        s, b = call(base, "/api/auth/me", headers={"Authorization": f"Bearer {t}"})
        print(f"me ({label}): {s} {sorted(json.loads(b).keys())}")

    compare("me: без токена", "/api/auth/me")
    compare("me: неверный токен", "/api/auth/me",
            headers={"Authorization": "Bearer bad"})
    compare("cart: без токена", "/api/cart")
    compare("cart: неверный токен", "/api/cart",
            headers={"Authorization": "Bearer bad"})


def cart_cases():
    print("== корзина ==")
    tm, tg = token(MOCK), token(GO)
    compare("cart: тело", "/api/cart", headers={"Authorization": f"Bearer {tm}"})

    items = [
        ("обычный", {"sku": "X", "price": 100, "qty": 2}),
        ("без qty", {"sku": "X", "price": 100}),
        ("без price", {"sku": "X"}),
        ("с brand/title/image", {"sku": "X", "brand": "B", "title": "T",
                                 "price": 10, "qty": 3, "image": "i"}),
        ("пустой объект", {}),
        ("price строкой", {"sku": "X", "price": "100", "qty": "2"}),
        ("price с подчёркиванием", {"sku": "X", "price": "1_0", "qty": 2}),
        ("price дробный", {"sku": "X", "price": 2.7, "qty": 3}),
        ("price bool", {"sku": "X", "price": True, "qty": 3}),
        ("qty ноль", {"sku": "X", "price": 100, "qty": 0}),
        ("отрицательная цена", {"sku": "X", "price": -5, "qty": 2}),
        ("большое число", {"sku": "X", "price": 10 ** 20, "qty": 2}),
        ("лишнее поле", {"price": 3, "qty": 2, "z": None}),
    ]
    for label, body in items:
        compare(f"cart/items: {label}", "/api/cart/items", method="POST", body=body,
                headers={"Authorization": f"Bearer {tm}"})

    # mock API падает на int("abc") и отдаёт 500 с текстом uvicorn.
    for label, body in (("price не число", {"price": "abc", "qty": 1}),
                        ("price null", {"price": None, "qty": 2}),
                        ("qty не число", {"price": 1, "qty": "x"})):
        global checks
        checks += 1
        s1, b1 = call(MOCK, "/api/cart/items", "POST", body,
                      {"Authorization": f"Bearer {tm}"})
        s2, b2 = call(GO, "/api/cart/items", "POST", body,
                      {"Authorization": f"Bearer {tg}"})
        if s1 == s2 == 500 and b1 == b2:
            continue
        failures.append(f"cart/items: {label}")
        print(f"DIFF cart/items: {label}\n  mock: {s1} {b1[:200]!r}\n   go : {s2} {b2[:200]!r}")

    compare("cart/items: массив", "/api/cart/items", method="POST", raw=b"[]",
            headers={"Authorization": f"Bearer {tm}"})
    compare("cart/items: строка", "/api/cart/items", method="POST", raw=b'"abc"',
            headers={"Authorization": f"Bearer {tm}"})
    compare("cart/items: пустое тело", "/api/cart/items", method="POST", raw=b"",
            headers={"Authorization": f"Bearer {tm}"})
    compare("cart/items: битый json", "/api/cart/items", method="POST", raw=b"{oops",
            headers={"Authorization": f"Bearer {tm}"})
    compare("cart/items: без content-type", "/api/cart/items", method="POST",
            raw=b'{"sku":"X"}', content_type=None,
            headers={"Authorization": f"Bearer {tm}"})
    compare("cart/items: GET", "/api/cart/items", method="GET",
            headers={"Authorization": f"Bearer {tm}"})
    compare("cart/items: без токена", "/api/cart/items", method="POST",
            body={"sku": "X"})
    compare("cart/items: битый json без токена", "/api/cart/items", method="POST",
            raw=b"{oops")
    compare("cart/items: не-json без токена", "/api/cart/items", method="POST",
            raw=b"sku=X", content_type="application/x-www-form-urlencoded")


def appointment_cases():
    print("== запись в Интерьерный центр ==")
    compare("dry-run: ок", "/api/center/appointment/dry-run", method="POST",
            body=APPOINTMENT)
    compare("dry-run: без согласия", "/api/center/appointment/dry-run", method="POST",
            body={**APPOINTMENT, "consent": False})
    compare("dry-run: лишний query", "/api/center/appointment/dry-run?x=1",
            method="POST", body=APPOINTMENT)
    compare("dry-run: GET", "/api/center/appointment/dry-run", method="GET")
    compare("create: без ключа", "/api/center/appointment", method="POST",
            body=APPOINTMENT)
    compare("create: без approved", "/api/center/appointment", method="POST",
            body=APPOINTMENT, headers={"Idempotency-Key": "parity-noprove"})
    compare("create: approved=false", "/api/center/appointment?approved=false",
            method="POST", body=APPOINTMENT,
            headers={"Idempotency-Key": "parity-false"})
    compare("create: approved=2", "/api/center/appointment?approved=2", method="POST",
            body=APPOINTMENT, headers={"Idempotency-Key": "parity-bad2"})
    compare("create: approved=", "/api/center/appointment?approved=", method="POST",
            body=APPOINTMENT, headers={"Idempotency-Key": "parity-empty"})
    compare("create: пустой ключ", "/api/center/appointment", method="POST",
            body=APPOINTMENT, headers={"Idempotency-Key": ""})
    compare("create: ключ из пробелов", "/api/center/appointment?approved=true",
            method="POST", body=APPOINTMENT, headers={"Idempotency-Key": "  "})
    compare("create: неверная форма", "/api/center/appointment", method="POST",
            body={"name": "", "phone": "1"},
            headers={"Idempotency-Key": "parity-invalid"})
    compare("create: неверная форма + approved=2", "/api/center/appointment?approved=2",
            method="POST", body={"name": "", "phone": "1"})

    key = "parity-" + uuid.uuid4().hex[:8]
    compare_shape("create: approved", "/api/center/appointment?approved=true",
                  method="POST", body=APPOINTMENT, headers={"Idempotency-Key": key})
    compare_shape("create: повтор ключа", "/api/center/appointment", method="POST",
                  body=APPOINTMENT, headers={"Idempotency-Key": key})
    compare_shape("create: повтор с невалидным телом", "/api/center/appointment",
                  method="POST", body={**APPOINTMENT, "consent": False},
                  headers={"Idempotency-Key": key})
    compare_shape("create: повтор без approved", "/api/center/appointment?approved=false",
                  method="POST", body=APPOINTMENT, headers={"Idempotency-Key": key})

    for base, label in ((MOCK, "mock"), (GO, "go")):
        s, b = call(base, "/api/center/appointment?approved=true", "POST", APPOINTMENT,
                    {"Idempotency-Key": key})
        j = json.loads(b)
        print(f"повтор ({label}): {s} idempotent_replay={j.get('idempotent_replay')}")

    compare("create: битый json", "/api/center/appointment", method="POST",
            raw=b"{oops", headers={"Idempotency-Key": "parity-badjson"})
    compare("create: битый json + approved=2", "/api/center/appointment?approved=2",
            method="POST", raw=b"{oops")
    compare("create: пустое тело", "/api/center/appointment", method="POST", raw=b"",
            headers={"Idempotency-Key": "parity-emptybody"})
    compare("create: массив", "/api/center/appointment", method="POST", raw=b"[]",
            headers={"Idempotency-Key": "parity-array"})
    compare("create: не-json", "/api/center/appointment", method="POST", raw=b"x=1",
            content_type="application/x-www-form-urlencoded",
            headers={"Idempotency-Key": "parity-form"})
    compare("create: битая кодировка", "/api/center/appointment", method="POST",
            raw=b"\x80{}", headers={"Idempotency-Key": "parity-utf8"})
    compare("appointments: не найдено", "/api/center/appointments/nope")
    compare("options: тело", "/api/center/appointment/options")


def validation_cases():
    print("== валидация формы (pydantic) ==")
    cases = [
        ("name пустой", {"name": ""}),
        ("name 81 символ", {"name": "я" * 81}),
        ("name 80 символов", {"name": "я" * 80}),
        ("name число", {"name": 123}),
        ("name null", {"name": None}),
        ("name отсутствует", {"__del__": "name"}),
        ("phone короткий", {"phone": "123"}),
        ("phone 23 символа", {"phone": "1" * 23}),
        ("phone 22 символа", {"phone": "1" * 22}),
        ("phone отсутствует", {"__del__": "phone"}),
        ("tasks строка", {"tasks": "Подбор интерьера"}),
        ("tasks null", {"tasks": None}),
        ("tasks с числом", {"tasks": [1, "a", None]}),
        ("tasks пустой", {"tasks": []}),
        ("tasks неизвестные", {"tasks": ["a", "b"]}),
        ("preferred_time неверное", {"preferred_time": "Ночью"}),
        ("preferred_time число", {"preferred_time": 5}),
        ("preferred_time null", {"preferred_time": None}),
        ("ideas 6", {"ideas": ["a", "b", "c", "d", "e", "f"]}),
        ("ideas 6 чисел", {"ideas": [1, 2, 3, 4, 5, 6]}),
        ("ideas null", {"ideas": None}),
        ("ideas с числом", {"ideas": [1, "a"]}),
        ("files 11", {"files": [str(i) for i in range(11)]}),
        ("wish 1201", {"wish": "я" * 1201}),
        ("wish 1200", {"wish": "я" * 1200}),
        ("wish null", {"wish": None}),
        ("consent строка", {"consent": "abc"}),
        ("consent число 2", {"consent": 2}),
        ("consent null", {"consent": None}),
        ("consent строка yes", {"consent": "yes"}),
        ("consent 1", {"consent": 1}),
        ("consent 0", {"consent": 0}),
        ("consent отсутствует", {"__del__": "consent"}),
        ("лишнее поле", {"extra": 1}),
        ("дата в прошлом", {"preferred_date": "2020-01-01"}),
        ("дата мусор", {"preferred_date": "не дата"}),
        ("дата пустая", {"preferred_date": ""}),
        ("дата число", {"preferred_date": 123}),
        ("дата true", {"preferred_date": True}),
        ("дата массив", {"preferred_date": []}),
        ("дата null", {"preferred_date": None}),
        ("дата 0", {"preferred_date": 0}),
        ("дата 86400", {"preferred_date": 86400}),
        ("дата год 0000", {"preferred_date": "0000-01-01"}),
        ("дата месяц 13", {"preferred_date": "2026-13-01"}),
        ("дата 30 февраля", {"preferred_date": "2026-02-30"}),
        ("дата 29 февраля", {"preferred_date": "2024-02-29"}),
        ("дата с временем", {"preferred_date": "2026-11-06T12:00:00"}),
        ("дата с T00:00", {"preferred_date": "2026-11-06T00:00:00Z"}),
        ("дата с долями", {"preferred_date": "2026-11-06T00:00:00.0000001"}),
        ("дата с таймзоной", {"preferred_date": "2026-11-06T00:00:00+03:00"}),
        ("дата без секунд", {"preferred_date": "2026-11-06T00:00"}),
        ("дата разделитель X", {"preferred_date": "2026-11-06X00:00:00"}),
        ("дата час 25", {"preferred_date": "2026-11-06T25:00:00"}),
        ("дата минута 60", {"preferred_date": "2026-11-06T23:60"}),
        ("дата короткая", {"preferred_date": "2026-11-0"}),
        ("дата компактная", {"preferred_date": "20261106"}),
        ("всё неверно", {"name": "", "phone": "1", "tasks": [1],
                         "preferred_date": "zz", "preferred_time": 5,
                         "ideas": [1, 2, 3, 4, 5, 6],
                         "files": [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11],
                         "wish": "я" * 1201}),
    ]
    for label, patch in cases:
        body = dict(APPOINTMENT)
        for field, value in patch.items():
            if field == "__del__":
                body.pop(value, None)
            else:
                body[field] = value
        compare(f"dry-run: {label}", "/api/center/appointment/dry-run", method="POST",
                body=body)

    print("== разбор тела ==")
    compare("dry-run: пустое тело", "/api/center/appointment/dry-run", method="POST",
            raw=b"")
    compare("dry-run: пробелы", "/api/center/appointment/dry-run", method="POST",
            raw=b"   ")
    compare("dry-run: битый json", "/api/center/appointment/dry-run", method="POST",
            raw=b"{oops")
    compare("dry-run: json + мусор", "/api/center/appointment/dry-run", method="POST",
            raw=b'{"name":"x"} extra')
    compare("dry-run: незакрытый json", "/api/center/appointment/dry-run", method="POST",
            raw=b'{"name":"x"')
    compare("dry-run: null", "/api/center/appointment/dry-run", method="POST",
            raw=b"null")
    compare("dry-run: число", "/api/center/appointment/dry-run", method="POST",
            raw=b"123")
    compare("dry-run: строка", "/api/center/appointment/dry-run", method="POST",
            raw=b'"abc"')
    compare("dry-run: массив", "/api/center/appointment/dry-run", method="POST",
            raw=b"[1]")
    compare("dry-run: BOM + валидное тело", "/api/center/appointment/dry-run",
            method="POST", raw=b"\xef\xbb\xbf" + json.dumps(APPOINTMENT).encode())
    compare("dry-run: BOM + битый json", "/api/center/appointment/dry-run",
            method="POST", raw=b"\xef\xbb\xbf{oops")
    compare("dry-run: utf-16", "/api/center/appointment/dry-run", method="POST",
            raw=b"\xff\xfe{\x00o\x00o\x00p\x00s\x00")
    compare("dry-run: битая utf-8", "/api/center/appointment/dry-run", method="POST",
            raw=b"\x80{oops")
    compare("dry-run: не-json content-type", "/api/center/appointment/dry-run",
            method="POST", raw=json.dumps(APPOINTMENT).encode(),
            content_type="text/plain")
    compare("dry-run: content-type с charset", "/api/center/appointment/dry-run",
            method="POST", raw=b"{}", content_type="application/json; charset=utf-8")
    compare("dry-run: content-type +json", "/api/center/appointment/dry-run",
            method="POST", raw=b"{}", content_type="application/vnd.api+json")
    compare("dry-run: без content-type", "/api/center/appointment/dry-run",
            method="POST", raw=json.dumps(APPOINTMENT).encode(), content_type=None)


def cpr_cases():
    print("== генерация ЦПР ==")
    tm, tg = token(MOCK), token(GO)
    compare("cpr/generate: без токена", "/api/cpr/generate", method="POST", body={})
    compare("cpr/generate: без токена, битый json", "/api/cpr/generate", method="POST",
            raw=b"{oops")
    compare("cpr/generate: пустое тело", "/api/cpr/generate", method="POST", raw=b"",
            headers={"Authorization": f"Bearer {tm}"})
    compare("cpr/generate: массив", "/api/cpr/generate", method="POST", raw=b"[]",
            headers={"Authorization": f"Bearer {tm}"})
    compare("cpr/generate: не-json", "/api/cpr/generate", method="POST", raw=b"{}",
            content_type=None, headers={"Authorization": f"Bearer {tm}"})
    compare("cpr/generate: битый json", "/api/cpr/generate", method="POST", raw=b"{oops",
            headers={"Authorization": f"Bearer {tm}"})
    compare("cpr/generate: неверный токен", "/api/cpr/generate", method="POST",
            body={}, headers={"Authorization": "Bearer bad"})
    compare("cpr/generate: GET", "/api/cpr/generate", method="GET")

    # Идентификатор считается как md5(str(payload)) — проверяем формулу.
    for label, payload in (("project_id", {"project_id": "p"}),
                           ("пустой объект", {}),
                           ("project_id null", {"project_id": None}),
                           ("project_id число", {"project_id": 5}),
                           ("project_id массив", {"project_id": []}),
                           ("вложенный объект", {"x": [1, 2, {"y": None}]})):
        compare_shape(f"cpr/generate: {label}", "/api/cpr/generate", method="POST",
                      body=payload, headers={"Authorization": f"Bearer {tm}"})
        raw = json.dumps(payload, ensure_ascii=False).encode()
        s1, b1 = call(MOCK, "/api/cpr/generate", "POST", raw=raw,
                      headers={"Authorization": f"Bearer {tm}"})
        s2, b2 = call(GO, "/api/cpr/generate", "POST", raw=raw,
                      headers={"Authorization": f"Bearer {tg}"})
        expected = hashlib.md5(str(payload).encode()).hexdigest()[:8]
        if s1 == s2 == 200:
            got_mock = json.loads(b1)["id"]
            got_go = json.loads(b2)["id"]
            if got_mock != f"cpr_{expected}" or got_go != got_mock:
                failures.append(f"cpr/generate: id {label}")
                print(f"DIFF cpr/generate: id {label}: mock={got_mock} go={got_go} "
                      f"ожидалось=cpr_{expected}")


def estimate_cases():
    print("== смета ==")
    compare("estimate/calculate: без тела", "/api/estimate/calculate", method="POST",
            raw=b"")
    compare("estimate/calculate: с телом", "/api/estimate/calculate", method="POST",
            body={"a": 1})
    compare("estimate/calculate: GET", "/api/estimate/calculate", method="GET")


def main() -> int:
    login_cases()
    cart_cases()
    appointment_cases()
    validation_cases()
    cpr_cases()
    estimate_cases()

    print(f"\nпроверок: {checks}, расхождений: {len(failures)}")
    if failures:
        print("расхождения:")
        for label in failures:
            print(f"  - {label}")
    return 1 if failures else 0


if __name__ == "__main__":
    sys.exit(main())
