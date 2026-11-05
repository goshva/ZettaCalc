#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Извлекает данные из JSON-блоков прототипов А и В в структурированный JSON
для использования в mock API.

Извлекает:
- Все top-level JSON-блоки (photos, interiors, dataset, data, adaptdata, hotspotData60, extraAnalogues64)
- apps52: 6 приложений (main, interiors, catalog, center, repair, about) с вложенным JSON
- Все data:image из обоих прототипов (489 + 143 = 632 картинки)

Результат:
- vue-prototypes/api/data/*.json — структурированные данные
- vue-prototypes/api/static/images/*.jpg — все картинки
"""
import re
import json
import base64
import io
from pathlib import Path

# Настройки
PROJECT_ROOT = Path("C:/soft/ZettaCalc")
SRC_A = PROJECT_ROOT / "Прототип А.html"
SRC_B = PROJECT_ROOT / "Прототип В.html"
DATA_DIR = PROJECT_ROOT / "vue-prototypes/api/data"
IMG_DIR = PROJECT_ROOT / "vue-prototypes/api/static/images"

DATA_DIR.mkdir(parents=True, exist_ok=True)
IMG_DIR.mkdir(parents=True, exist_ok=True)

IMG_COUNTER = [0]

def save_image(b64: str) -> str:
    """Сохраняет base64 картинку в файл, возвращает имя."""
    IMG_COUNTER[0] += 1
    fname = f"img_{IMG_COUNTER[0]:04d}.jpg"
    raw = base64.b64decode(b64)
    (IMG_DIR / fname).write_bytes(raw)
    return fname

def extract_json_blocks(html: str) -> dict:
    """Извлекает все <script type="application/json" id="...">...</script>."""
    blocks = {}
    for m in re.finditer(
        r'<script[^>]*\bid\s*=\s*["\']([^\s"\']+)["\'][^>]*type\s*=\s*["\']?application/json["\']?[^>]*>(.*?)</script>',
        html, re.S | re.I
    ):
        ident = m.group(1)
        content = m.group(2).strip()
        try:
            obj = json.loads(content)
            blocks[ident] = obj
        except json.JSONDecodeError:
            # Сохраняем как raw
            blocks[ident] = {"_raw": content}
    return blocks

def extract_data_images(html: str, prefix: str = "") -> dict:
    """Извлекает все data:image из HTML, возвращает {b64_prefix: fname}."""
    img_map = {}
    counter = [0]
    
    def replacer(m):
        ext = m.group(1).lower()
        b64 = m.group(2)
        counter[0] += 1
        fname = f"{prefix}img_{counter[0]:04d}.jpg"
        try:
            raw = base64.b64decode(b64)
            (IMG_DIR / fname).write_bytes(raw)
            img_map[b64[:48]] = fname
            return fname
        except Exception as e:
            print(f"  ! IMG error: {e}")
            return "img/error.jpg"
    
    re_img = re.compile(r'data:image/(\w+);base64,([A-Za-z0-9+/=]+)')
    re_img.sub(replacer, html)
    print(f"  Извлечено {counter[0]} data:image из {prefix or 'MAIN'}")
    return img_map

def extract_nested_json(html: str, img_map: dict) -> dict:
    """Извлекает вложенные JSON из HTML приложений."""
    blocks = {}
    
    # Находим script-теги с id внутри html
    for m in re.finditer(
        r'<script[^>]*\bid\s*=\s*["\']([^\s"\']+)["\'][^>]*>(.*?)</script>',
        html, re.S | re.I
    ):
        ident = m.group(1)
        content = m.group(2).strip()
        
        # Проверяем, это JSON?
        if not content.startswith(('{', '[')):
            continue
        
        try:
            obj = json.loads(content)
            blocks[ident] = obj
        except json.JSONDecodeError:
            blocks[ident] = {"_raw": content[:1000]}
    
    return blocks

def process_prototype_a():
    """Обрабатывает Прототип А."""
    print("=" * 60)
    print("Прототип А")
    print("=" * 60)
    
    html = SRC_A.read_text(encoding="utf-8")
    print(f"Размер HTML: {len(html):,} chars")
    
    # 1. Top-level JSON-блоки
    print("\n--- Top-level JSON blocks ---")
    top_blocks = extract_json_blocks(html)
    print(f"Найдено {len(top_blocks)} блоков:")
    for name in top_blocks:
        print(f"  - {name}")
    
    # Сохраняем каждый блок отдельно
    for name, data in top_blocks.items():
        (DATA_DIR / f"proto_a_{name}.json").write_text(
            json.dumps(data, ensure_ascii=False, indent=2), encoding="utf-8"
        )
        print(f"  OK {name}.json")
    
    # 2. apps52
    print("\n--- apps52 ---")
    apps = top_blocks.get("apps52", {})
    if isinstance(apps, dict):
        app_data = {}
        for app_name, app_content in apps.items():
            print(f"\n  Приложение: {app_name}")
            app_data[app_name] = {}
            
            if "html" in app_content:
                app_html = app_content["html"]
                print(f"    html: {len(app_html):,} chars")
                
                # Извлекаем вложенные JSON
                nested = extract_nested_json(app_html, {})
                print(f"    Вложенные JSON: {list(nested.keys())}")
                app_data[app_name]["nested"] = nested
                
                # Извлекаем картинки
                img_map = extract_data_images(app_html, f"a_{app_name}_")
                print(f"    Картинки: {len(img_map)}")
                
                # Сохраняем HTML
                app_data[app_name]["html_length"] = len(app_html)
            
            if "css" in app_content:
                css = app_content["css"]
                app_data[app_name]["css_length"] = len(css)
                print(f"    css: {len(css):,} chars")
        
        (DATA_DIR / "proto_a_apps52.json").write_text(
            json.dumps(app_data, ensure_ascii=False, indent=2), encoding="utf-8"
        )
        print(f"  ✓ apps52.json")
    
    # 3. Картинки в остальном HTML (вне JSON-блоков)
    print("\n--- Картинки вне JSON ---")
    # Сначала удаляем JSON-блоки из HTML для извлечения оставшихся картинок
    html_no_json = html
    for m in re.finditer(
        r'<script[^>]*\bid\s*=\s*["\']([^\s"\']+)["\'][^>]*type\s*=\s*["\']?application/json["\']?[^>]*>(.*?)</script>',
        html, re.S | re.I
    ):
        start = m.start()
        end = m.end()
        html_no_json = html_no_json[:start] + " " * (end - start) + html_no_json[end:]
    
    img_map = extract_data_images(html_no_json, "a_main_")
    
    # Итого
    total_imgs = IMG_COUNTER[0]
    print(f"\n  Итого картинок из Прототипа А: {total_imgs}")

def process_prototype_b():
    """Обрабатывает Прототип В."""
    print("\n" + "=" * 60)
    print("Прототип В")
    print("=" * 60)
    
    html = SRC_B.read_text(encoding="utf-8")
    print(f"Размер HTML: {len(html):,} chars")
    
    # 1. Top-level JSON-блоки
    print("\n--- Top-level JSON blocks ---")
    top_blocks = extract_json_blocks(html)
    print(f"Найдено {len(top_blocks)} блоков:")
    for name in top_blocks:
        print(f"  - {name}")
    
    for name, data in top_blocks.items():
        (DATA_DIR / f"proto_b_{name}.json").write_text(
            json.dumps(data, ensure_ascii=False, indent=2), encoding="utf-8"
        )
        print(f"  ✓ {name}.json")
    
    # 2. Картинки
    print("\n--- Картинки ---")
    img_map = extract_data_images(html, "b_")
    
    total_imgs = IMG_COUNTER[0]
    print(f"\n  Итого картинок из Прототипа В: {total_imgs}")

if __name__ == "__main__":
    print("Извлечение данных из прототипов А и В...")
    print(f"Цель: {DATA_DIR}")
    print(f"Картинки: {IMG_DIR}")
    
    process_prototype_a()
    process_prototype_b()
    
    print("\n" + "=" * 60)
    print("ГОТОВО!")
    print(f"Всего картинок: {IMG_COUNTER[0]}")
    print(f"Данные: {list(DATA_DIR.glob('*.json'))}")
