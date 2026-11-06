# ZettaCalc backend (Go)

Go-сервис заменяет Python-mock (`vue-prototypes/api/main.py`) на том же HTTP-контракте:
те же пути, коды ответов и тела **байт-в-байт**. Смысл — детерминированное ядро
без LLM (`AGENTS.md`), статически типизированный бинарь и данные в PostgreSQL
вместо JSON-файлов в памяти процесса.

## Запуск

    # окружение
    export DATABASE_URL=postgres://postgres:postgres@127.0.0.1:5432/zettacalc
    export ZETTA_REPO_ROOT=C:/soft/ZettaCalc

    # миграции применяются на старте
    cd backend
    go run ./cmd/api -migrations migrations

    # первичная загрузка данных mock API (идемпотентна, пишет seed_meta)
    go run ./cmd/seed -data ../vue-prototypes/api/data -images ../vue-prototypes/api/static/images

Сервис слушает `ZETTA_ADDR` (по умолчанию `127.0.0.1:8080`), статику картинок
читает из `ZETTA_STATIC_IMAGES_DIR` (по умолчанию — картинки mock API).

## Проверка паритета

Три скрипта сравнивают ответы mock и Go-сервиса; оба сервиса должны быть
запущены (mock — на 8010, Go — на 8080).

    python backend/scripts/parity.py           # 173 маршрута, тела байт-в-байт
    python backend/scripts/parity_write.py     # write-эндпоинты, ACL, pydantic-ошибки
    python backend/scripts/parity_nonfinite.py # Infinity/NaN в JSON

Все три должны заканчиваться нулём расхождений. `/health` и `/api/health`
сравниваются без полей `mock` и `appointments`: первое по определению
различается, второе — состояние процесса (в mock счётчик растёт от записей,
в Go равен числу строк в БД).

## Структура

    cmd/api        composition root: конфиг, пул pgx, миграции, HTTP-сервер
    cmd/seed       загрузчик данных mock API в PostgreSQL
    internal/<контекст>/
      domain/      инварианты, сервисы, порты (без HTTP и SQL)
      adapters/pg/ репозитории
      adapters/httpapi/  обработчики
    internal/platform/  общая инфраструктура: auth, httpx, idempotency,
                        pyjson, pybody, pyvalidate, jsonx, config, pgxdb
    migrations/    версионированные SQL-миграции (0001…)
    scripts/       скрипты сверки паритета

## Паритет важнее «правильности»

Сервис воспроизводит поведение FastAPI и pydantic, включая неочевидные места:

- `platform/pyjson` — собственный разбор JSON в семантике Python: позиции
  ошибок в символах, `int` не смешивается с `float`, `repr()` для
  `cpr_id = md5(str(payload))`, литералы `Infinity`/`NaN`/`1e400`;
- `platform/pybody` — порядок FastAPI: разбор тела → 401 по токену → 422 по
  телу (поэтому битый JSON отвечает 422 и без токена);
- `platform/pyvalidate` — формы ошибок pydantic v2 (`type`, `loc`, `msg`,
  `input`, `ctx`), lax-приведение типов и разбор дат по speedate;
- `platform/httpx` — `json_invalid` вытесняет ошибки query, а `NaN`/`±Inf`
  в теле ответа дают 500 «Internal Server Error», как `json.dumps` в Python.

Такое поведение осознанное: цель миграции — чтобы фронтенды прототипов
работали без изменений, а не чтобы получить «более чистый» API.

## Известное расхождение

Служебные пути FastAPI (`/docs`, `/openapi.json`, `/redoc`,
`/docs/oauth2-redirect`) в Go-сервисе не реализованы и отвечают 404, тогда как
mock отдаёт Swagger UI и сгенерированную схему. Ссылка «Swagger UI» в футере
prototype B (проксируется через `/docs`) поэтому ведёт на 404. Это не влияет
на прикладные сценарии: контракт API описан в `schemas/openapi/cpr.yaml`, а
эталонный `openapi.json` всегда можно снять с запущенного mock. Решение
(отдавать артефакты из Go или убрать ссылку из футера) — за владельцем
прототипа.
