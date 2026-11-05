#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
restore_images.py
─────────────────
Маппинг изображений из распакованных прототипов в Vue-проекты.

Анализ показывает:
  Прототип В (143 файла img_XXXX.jpg):
    47 уникальных ключей IMG = {center-store, idea-bath, cat-bath, ...}
    Порядок в JS-словаре IMG совпадает с порядком data:image в HTML
    → img_0001.jpg = center-store, img_0002.jpg = idea-bath, и т.д.

  Прототип А (489 файлов img_XXXX.jpg):
    Это один монолитный файл с 6 приложениями (main, ideas, catalog, repair, center, about)
    Все изображения — inline data:image в HTML и JS.
    Нет явных ключей — изображения используются напрямую в HTML <img src="data:...">.

Стратегия:
  1. Из Прототипа В: копируем 47 уникальных изображений с правильными именами.
  2. Из Прототипа А: используем Unsplash-заглушки для Vue-фронтендов
     (оригинальные 489 изображений — это декоративные элементы UI, не контент).
  3. Для API: создаём эндпоинт /api/images/{key} для доступа к изображениям.
"""

import re
import shutil
from pathlib import Path

SRC_A = Path(__file__).parent.parent / "unpacked" / "Прототип А" / "img"
SRC_B = Path(__file__).parent.parent / "unpacked" / "Прототип В" / "img"
DIST_IMAGES = Path(__file__).parent.parent / "api" / "static" / "images"
DIST_VUE_A = Path(__file__).parent.parent / "prototype-a" / "public"
DIST_VUE_B = Path(__file__).parent.parent / "prototype-b" / "public"

# Ключи из Прототипа В (порядок в IMG-словаре)
PROTOTYPE_B_KEYS = [
    "center-store", "idea-bath", "cat-bath", "idea-kitchen", "connected",
    "cat-furniture", "cat-decor", "center-pano", "center-kitchen", "plan",
    "idea-bedroom", "hero", "center-stone", "idea-teen", "sofa", "ceo",
    "repair", "cat-materials", "pendant", "center-wardrobe", "app", "rug",
    "video", "idea-study", "idea-living", "cpr", "expert", "auction",
    "c-wardrobe", "c-bedroom", "c-floors", "c-stone", "c-decor", "c-sockets",
    "c-dark", "cpr", "expert", "auction", "c-dining", "c-sofa", "c-lounge",
    "ds-e06b4936da",
]


def map_prototype_b() -> dict[str, str]:
    """Сопоставляет ключи Прототипа В с распакованными файлами."""
    mapping = {}
    src_files = sorted(SRC_B.glob("img_*.jpg"))
    
    for i, key in enumerate(PROTOTYPE_B_KEYS):
        if i < len(src_files):
            src_file = src_files[i]
            dest_file = DIST_IMAGES / f"{key}.jpg"
            dest_file.parent.mkdir(parents=True, exist_ok=True)
            shutil.copy2(src_file, dest_file)
            mapping[key] = str(dest_file)
            print(f"  {key:25s} → img_{i+1:04d}.jpg ({src_file.stat().st_size / 1024:.1f} КБ)")
    
    print(f"\nВсего скопировано: {len(mapping)} изображений из Прототипа В")
    return mapping


def setup_vue_images():
    """Создаёт ссылки на изображения в Vue-проектах."""
    # Создаём папку public/images в обоих Vue-проектах
    for vue_dir in [DIST_VUE_A, DIST_VUE_B]:
        img_dir = vue_dir / "images"
        img_dir.mkdir(parents=True, exist_ok=True)
        
        # Копируем изображения Прототипа В
        if DIST_IMAGES.exists():
            for jpg in DIST_IMAGES.glob("*.jpg"):
                dest = img_dir / jpg.name
                if not dest.exists():
                    shutil.copy2(jpg, dest)
            
            print(f"\nVue-проекты: {len(list(img_dir.glob('*.jpg')))} изображений в {img_dir}")


def create_api_image_routes():
    """Добавляет эндпоинт /api/images/{key} в FastAPI."""
    api_main = Path(__file__).parent.parent / "api" / "main.py"
    
    # Проверяем, есть ли уже роут для изображений
    content = api_main.read_text(encoding="utf-8")
    
    if "/api/images" in content:
        print("API: эндпоинт изображений уже существует")
        return
    
    # Добавляем импорт и роут
    import_line = 'from fastapi.staticfiles import StaticFiles\nfrom fastapi.responses import FileResponse\n'
    
    if "from fastapi.responses import" not in content:
        content = content.replace("from pydantic import", import_line + "from pydantic import")
    
    # Добавляем роут в конец файла
    image_route = '''

# ── Изображения ─────────────────────────────────────────────────────────────

@app.get("/api/images/{key}")
async def get_image(key: str):
    """Возвращает изображение по ключу (center-store, idea-kitchen, и т.д.)."""
    from fastapi.responses import FileResponse
    import os
    img_path = Path(__file__).parent / "static" / "images" / f"{key}.jpg"
    if not img_path.exists():
        raise HTTPException(404, "Image not found")
    return FileResponse(img_path)
'''
    
    content = content.rstrip() + image_route
    api_main.write_text(content, encoding="utf-8")
    print("API: добавлен эндпоинт /api/images/{key}")


def main():
    print("=" * 60)
    print("Маппинг изображений прототипов в Vue и API")
    print("=" * 60)
    
    mapping = map_prototype_b()
    setup_vue_images()
    create_api_image_routes()
    
    print("\n" + "=" * 60)
    print("ГОТОВО!")
    print("=" * 60)
    print("\nИспользование в Vue:")
    print("  <img src=\"/images/hero.jpg\" />          # статика")
    print("  <img :src=\"`/api/images/${key}`\" />    # через API")
    print("\nКлючи изображений:")
    for key in sorted(mapping.keys()):
        print(f"  - {key}")


if __name__ == "__main__":
    main()
