"""Печатает вырезанные функции в читаемом виде.

Оригинальный app.js минифицирован: почти весь код в одну строку.
Скрипт расставляет переносы, чтобы разметку можно было читать и переносить в Vue.
"""

from __future__ import annotations

import re
import sys
from pathlib import Path

SRC = Path(__file__).parent.parent / "unpacked" / "Прототип А" / "js" / "app.js"
DECL = r"(?:^|[\n;])(?:function\s+{n}\s*\(|(?:const|let|var)?\s*{n}\s*=\s*(?:async\s*)?function\s*\()"


def extract(src: str, name: str) -> str | None:
    m = re.search(DECL.format(n=re.escape(name)), src, re.M)
    if not m:
        return None
    brace = src.index("{", m.start())
    depth = 0
    for i in range(brace, len(src)):
        if src[i] == "{":
            depth += 1
        elif src[i] == "}":
            depth -= 1
            if depth == 0:
                return src[m.start():i + 1]
    return None


def pretty(code: str) -> str:
    # переносы перед закрывающими тегами и после `;` внутри шаблонов
    out = re.sub(r">(?=<)", ">\n", code)
    out = re.sub(r"`(?=.)", "`\n", out)
    out = re.sub(r"(?<=`)(?!\s*$)", "", out)
    return out


OUT = Path(__file__).parent / "_pretty"


def extract_nth(src: str, name: str, n: int) -> str | None:
    """Возвращает n-е (1-based) объявление функции — в оригинале они переопределяются."""
    matches = list(re.finditer(DECL.format(n=re.escape(name)), src, re.M))
    if len(matches) < n:
        return None
    m = matches[n - 1]
    brace = src.index("{", m.start())
    depth = 0
    for i in range(brace, len(src)):
        if src[i] == "{":
            depth += 1
        elif src[i] == "}":
            depth -= 1
            if depth == 0:
                return src[m.start():i + 1]
    return None


def main() -> None:
    names = sys.argv[1:]
    if not names:
        print("укажите имена функций (можно с номером версии: listing.3)")
        return
    src = SRC.read_text(encoding="utf-8")
    OUT.mkdir(exist_ok=True)
    for spec in names:
        name, _, ver = spec.partition(".")
        code = extract_nth(src, name, int(ver) if ver else 1)
        if code is None:
            print(f"не найдено: {spec}")
            continue
        # разбиваем шаблонные литералы на строки по тегам
        text = re.sub(r"(?<=>)(?=<)", ">\n", code)
        text = re.sub(r"(?<=[;}])(?=(?:const|let|let|\$\(|document|if|return|for|const)\s)", ";\n", text)
        (OUT / f"{spec}.txt").write_text(text, encoding="utf-8")
        print(f"{spec}: {len(text)} символов")


if __name__ == "__main__":
    main()
