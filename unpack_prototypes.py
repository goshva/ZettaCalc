#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
unpack_prototypes.py
────────────────────
Разбивает монолитные HTML-прототипы на отдельные статические файлы:
  · CSS   →  css/
  · JS    →  js/
  · IMG   →  img/  (все изображения → .jpg сжатием 85%)
  · JSON  →  json/  (данные из <script type="application/json" id="...">)
  · HTML  →  index.html  (ссылается на css/, js/, json/, img/)

Использование:
    python unpack_prototypes.py                     # оба прототипа
    python unpack_prototypes.py "Прототип А.html"   # один файл

Затем открыть index.html через Live Server (VS Code) или:
    python -m http.server 8080 --directory unpacked/Прототип А
"""

from __future__ import annotations

import base64
import io
import json
import logging
import re
import shutil
import sys
from pathlib import Path

import PIL.Image as Image

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s  %(levelname)-7s  %(message)s",
)
log = logging.getLogger("unpack")

PROJECT_ROOT = Path(__file__).resolve().parent
UNPACKED_DIR = PROJECT_ROOT / "unpacked"
JPEG_QUALITY = 85

# ── Картинки ────────────────────────────────────────────────────────────────

RE_DATA_IMG = re.compile(
    r'data:image/(\w+);base64,([A-Za-z0-9+/=]+)'
)


def _to_jpeg(b64: str) -> bytes:
    """Декодирует base64, конвертирует в JPEG (RGB)."""
    raw = base64.b64decode(b64)
    ext = RE_DATA_IMG.match(b64).group(1).lower() if RE_DATA_IMG.match(b64) else "png"
    if ext == "jpeg":
        return raw
    try:
        img = Image.open(io.BytesIO(raw)).convert("RGB")
        buf = io.BytesIO()
        img.save(buf, format="JPEG", quality=JPEG_QUALITY)
        return buf.getvalue()
    except Exception:
        return raw  # fallback — оригинал


# ── Прототип А: SPA с JSON-данными и динамическим style52 ──────────────────

def unpack_prototype_a(src: Path, dst: Path) -> None:
    """
    Архитектура Прототипа А:
      3 × <style>        (один без id, два с id="appstyle52" и id="about-shell68")
      8 × <script type="application/json" id="...">  (JSON-данные)
      1 × <script>       (SPA-роутер, ссылается на style52, apps52 и другие id)
      489 × data:image   (inline в HTML и JS)
      4 × экранированных </script> внутри JSON

    Стратегия:
      1. Извлечь все <script type="application/json"> → json/*.json
      2. Извлечь <script> → js/app.js
      3. Извлечь все <style> → css/styles.css (объединить)
      4. Извлечь data:image → img/*.jpg
      5. Переписать HTML: заменить теги на ссылки на css/, js/, json/, img/
    """
    html = src.read_text(encoding="utf-8")
    css_dir = dst / "css"
    js_dir = dst / "js"
    img_dir = dst / "img"
    json_dir = dst / "json"
    for d in (css_dir, js_dir, json_dir, img_dir):
        d.mkdir(parents=True, exist_ok=True)

    # ── 1. JSON-данные ─────────────────────────────────────────────────
    json_blocks: dict[str, str] = {}
    for m in re.finditer(
        r'<script\b[^>]*type\s*=\s*["\']?application/json["\']?[^>]*\bid\s*=\s*["\']([^\s"\']+)["\'][^>]*>(.*?)</script>',
        html, re.S | re.I,
    ):
        ident = m.group(1)
        content = m.group(2).strip()
        json_blocks[ident] = content
        (json_dir / f"{ident}.json").write_text(content, encoding="utf-8")

    # Заменяем JSON-скрипты на <link type="json"> для SPA-роутера
    for ident in json_blocks:
        # Паттерн: <script ... id="ident" type="application/json" ...>...</script>
        pat = re.compile(
            re.escape('<script') + r'[^>]*\bid\s*=\s*["\']' + re.escape(ident) +
            r'["\'][^>]*type\s*=\s*["\']?application/json["\']?[^>]*>(.*?)</script>',
            re.S | re.I,
        )
        # Не заменяем — SPA-роутер делает JSON.parse(document.getElementById(ident).textContent)
        # Нужно оставить как <script type="application/json" id="...">...</script>

    # ── 2. Изображения (data:image URIs) ───────────────────────────────
    img_map: dict[str, str] = {}
    img_counter = [0]

    def _save_img(m: re.Match) -> str:
        ext = m.group(1).lower()
        b64 = m.group(2)
        img_counter[0] += 1
        fname = f"img_{img_counter[0]:04d}.jpg"
        try:
            data = _to_jpeg(b64)
        except Exception as e:
            log.warning("IMG #%d: %s", img_counter[0], e)
            data = base64.b64decode(b64)
        (img_dir / fname).write_bytes(data)
        img_map[b64[:48]] = fname
        return fname

    # data:image в HTML и JS
    html = RE_DATA_IMG.sub(_save_img, html)

    # ── 3. CSS ─────────────────────────────────────────────────────────
    styles: list[str] = []
    for m in re.finditer(r'<style([^>]*)>(.*?)</style>', html, re.S | re.I):
        styles.append(m.group(2))

    combined_css = "\n".join(styles)
    (css_dir / "styles.css").write_text(combined_css, encoding="utf-8")

    # Заменяем все <style>...</style> на один <link>
    html = re.sub(
        r'<style[^>]*>.*?</style>',
        '<link rel="stylesheet" href="css/styles.css">',
        html,
        flags=re.S,
    )

    # ── 4. JS ──────────────────────────────────────────────────────────
    js_blocks: list[str] = []
    for m in re.finditer(r'<script>(.*?)</script>', html, re.S):
        content = m.group(1).strip()
        if content and not re.match(r'^\s*type\s*=\s*["\']?application/json', content, re.S):
            js_blocks.append(content)

    combined_js = ";\n".join(js_blocks)
    (js_dir / "app.js").write_text(combined_js, encoding="utf-8")

    # Заменяем <script>...</script> на <script src="...">
    html = re.sub(
        r'<script>(.*?)</script>',
        '<script src="js/app.js"></script>',
        html,
        flags=re.S,
    )

    # ── 5. Чистка и запись ─────────────────────────────────────────────
    html = re.sub(r'\n{3,}', '\n\n', html)
    html = re.sub(r'\s{2,}', ' ', html)
    (dst / "index.html").write_text(html, encoding="utf-8")

    log.info(
        "✓ Прототип А: %d JSON, %d CSS, %d JS, %d IMG → %s",
        len(json_blocks), len(styles), len(js_blocks), img_counter[0], dst,
    )


# ── Прототип В: классический SPA (1 style + 1 script) ─────────────────────

def unpack_prototype_b(src: Path, dst: Path) -> None:
    """
    Архитектура Прототипа В:
      1 × <style>       (inline CSS, ~1690 строк)
      1 × <script>      (SPA-роутер + IMG словарь с base64)
      144 × data:image  (встроенные в HTML img src и JS IMG словарь)

    Стратегия:
      1. Извлечь <style> → css/styles.css
      2. Извлечь <script> → js/app.js
      3. Извлечь data:image → img/*.jpg
      4. Переписать HTML
    """
    html = src.read_text(encoding="utf-8")
    css_dir = dst / "css"
    js_dir = dst / "js"
    img_dir = dst / "img"
    for d in (css_dir, js_dir, img_dir):
        d.mkdir(parents=True, exist_ok=True)

    # ── 1. Изображения ─────────────────────────────────────────────────
    img_counter = [0]

    def _save_img(m: re.Match) -> str:
        ext = m.group(1).lower()
        b64 = m.group(2)
        img_counter[0] += 1
        fname = f"img_{img_counter[0]:04d}.jpg"
        try:
            data = _to_jpeg(b64)
        except Exception as e:
            log.warning("IMG #%d: %s", img_counter[0], e)
            data = base64.b64decode(b64)
        (img_dir / fname).write_bytes(data)
        return fname

    html = RE_DATA_IMG.sub(_save_img, html)

    # ── 2. CSS ─────────────────────────────────────────────────────────
    styles: list[str] = []
    for m in re.finditer(r'<style>(.*?)</style>', html, re.S):
        styles.append(m.group(1))

    (css_dir / "styles.css").write_text("\n".join(styles), encoding="utf-8")
    html = re.sub(
        r'<style>.*?</style>',
        '<link rel="stylesheet" href="css/styles.css">',
        html,
        flags=re.S,
    )

    # ── 3. JS ──────────────────────────────────────────────────────────
    js_blocks: list[str] = []
    for m in re.finditer(r'<script>(.*?)</script>', html, re.S):
        content = m.group(1).strip()
        if content:
            js_blocks.append(content)

    (js_dir / "app.js").write_text(";\n".join(js_blocks), encoding="utf-8")
    html = re.sub(
        r'<script>.*?</script>',
        '<script src="js/app.js"></script>',
        html,
        flags=re.S,
    )

    # ── 4. Чистка и запись ─────────────────────────────────────────────
    html = re.sub(r'\n{3,}', '\n\n', html)
    html = re.sub(r'\s{2,}', ' ', html)
    (dst / "index.html").write_text(html, encoding="utf-8")

    log.info(
        "✓ Прототип В: %d CSS, %d JS, %d IMG → %s",
        len(styles), len(js_blocks), img_counter[0], dst,
    )


# ── Автоопределение типа ───────────────────────────────────────────────────

def detect_type(src: Path) -> str:
    """Определяет тип прототипа по структуре."""
    raw = src.read_text(encoding="utf-8", errors="replace")
    json_scripts = len(
        re.findall(r'type\s*=\s*["\']?application/json["\']?\s+id\s*=', raw)
    )
    if json_scripts >= 3:
        return "a"
    return "b"


# ── Точка входа ────────────────────────────────────────────────────────────

UNPACKERS = {
    "a": unpack_prototype_a,
    "b": unpack_prototype_b,
}

DEFAULT_FILES = [
    PROJECT_ROOT / "Прототип А.html",
    PROJECT_ROOT / "Прототип В.html",
]


def main() -> None:
    files = [PROJECT_ROOT / f for f in sys.argv[1:]] if len(sys.argv) > 1 else DEFAULT_FILES

    if UNPACKED_DIR.exists():
        shutil.rmtree(UNPACKED_DIR)
    UNPACKED_DIR.mkdir(parents=True)

    for fpath in files:
        if not fpath.exists():
            log.error("Не найден: %s", fpath)
            continue

        ptype = detect_type(fpath)
        name = fpath.stem
        dst = UNPACKED_DIR / name

        log.info("=" * 60)
        log.info("Тип %s: %s → %s", ptype, fpath.name, dst)
        UNPACKERS[ptype](fpath, dst)

    # ── Итог ───────────────────────────────────────────────────────────
    log.info("=" * 60)
    log.info("ГОТОВО!")
    log.info("")
    for d in sorted(UNPACKED_DIR.glob("*")):
        if not d.is_dir():
            continue
        total = sum(1 for _ in d.rglob("*"))
        total_size = sum(f.stat().st_size for f in d.rglob("*") if f.is_file())
        log.info("  %s/  (%d файлов, %.1f МБ)", d.name, total, total_size / 1_048_576)
    log.info("")
    log.info("Live Server (VS Code):")
    log.info("  1. Откройте unpacked/Прототип А/ или unpacked/Прототип В/")
    log.info("  2. Правый клик → index.html → Open with Live Server")
    log.info("")
    log.info("Или:")
    log.info("  cd %s/Прототип А && python -m http.server 8080", UNPACKED_DIR)


if __name__ == "__main__":
    main()
