"""Конвертация ANSWER.md в DOCX для отправки владельцу.

Запуск:  .venv/Scripts/python.exe scripts/md_to_docx.py
Результат: ANSWER.docx в корне проекта.

Поддерживается: заголовки, таблицы, списки, нумерация, **жирный**,
`код`, цитаты, горизонтальные разделители, ссылки (текст без URL-ссылок).
"""
from __future__ import annotations

import re
from pathlib import Path

from docx import Document
from docx.enum.table import WD_TABLE_ALIGNMENT
from docx.enum.text import WD_ALIGN_PARAGRAPH
from docx.shared import Pt, RGBColor

ROOT = Path(__file__).resolve().parent.parent
SRC = ROOT / "ANSWER.md"
DST = ROOT / "ANSWER.docx"

INLINE_RE = re.compile(r"(\*\*.+?\*\*|`[^`]+`|\*[^*]+\*)")


def add_inline(par, text: str) -> None:
    """Добавляет текст в абзац с разбором **жирного**, `кода` и *курсива*."""
    text = re.sub(r"\[([^\]]+)\]\([^)]+\)", r"\1", text)
    for part in INLINE_RE.split(text):
        if not part:
            continue
        if part.startswith("**") and part.endswith("**"):
            run = par.add_run(part[2:-2])
            run.bold = True
        elif part.startswith("`") and part.endswith("`"):
            run = par.add_run(part[1:-1])
            run.font.name = "Consolas"
            run.font.size = Pt(9.5)
        elif part.startswith("*") and part.endswith("*") and len(part) > 2:
            run = par.add_run(part[1:-1])
            run.italic = True
        else:
            par.add_run(part)


def flush_table(doc: Document, rows: list[list[str]]) -> None:
    if not rows:
        return
    ncols = max(len(r) for r in rows)
    table = doc.add_table(rows=len(rows), cols=ncols)
    table.style = "Table Grid"
    table.alignment = WD_TABLE_ALIGNMENT.LEFT
    for i, row in enumerate(rows):
        for j in range(ncols):
            cell = table.cell(i, j)
            cell.text = ""
            par = cell.paragraphs[0]
            add_inline(par, row[j] if j < len(row) else "")
            for run in par.runs:
                run.font.size = Pt(9)
                if i == 0:
                    run.bold = True
    doc.add_paragraph()


def split_row(line: str) -> list[str]:
    return [c.strip() for c in line.strip().strip("|").split("|")]


def is_sep(line: str) -> bool:
    return bool(re.match(r"^\s*\|?[\s:|-]+\|[\s:|-]*$", line)) and "-" in line


def main() -> None:
    md = SRC.read_text(encoding="utf-8")
    lines = md.splitlines()

    doc = Document()
    style = doc.styles["Normal"]
    style.font.name = "Calibri"
    style.font.size = Pt(10.5)

    table_rows: list[list[str]] = []
    in_code = False
    code_buf: list[str] = []
    i = 0
    while i < len(lines):
        line = lines[i]

        if line.strip().startswith("```"):
            if in_code:
                par = doc.add_paragraph()
                run = par.add_run("\n".join(code_buf))
                run.font.name = "Consolas"
                run.font.size = Pt(9)
                run.font.color.rgb = RGBColor(0x44, 0x44, 0x44)
                code_buf = []
            in_code = not in_code
            i += 1
            continue
        if in_code:
            code_buf.append(line)
            i += 1
            continue

        if line.strip().startswith("|"):
            if not is_sep(line):
                table_rows.append(split_row(line))
            i += 1
            continue
        if table_rows:
            flush_table(doc, table_rows)
            table_rows = []

        stripped = line.strip()
        if not stripped:
            i += 1
            continue

        m = re.match(r"^(#{1,4})\s+(.*)$", stripped)
        if m:
            level = len(m.group(1))
            text = re.sub(r"\*\*", "", m.group(2))
            doc.add_heading(text, level=min(level, 4))
            i += 1
            continue

        if stripped in ("---", "***", "___"):
            i += 1
            continue

        m = re.match(r"^[-*]\s+(.*)$", stripped)
        if m:
            par = doc.add_paragraph(style="List Bullet")
            add_inline(par, m.group(1))
            i += 1
            continue

        m = re.match(r"^\d+[.)]\s+(.*)$", stripped)
        if m:
            par = doc.add_paragraph(style="List Number")
            add_inline(par, m.group(1))
            i += 1
            continue

        if stripped.startswith(">"):
            par = doc.add_paragraph(style="Intense Quote")
            add_inline(par, stripped.lstrip("> "))
            i += 1
            continue

        par = doc.add_paragraph()
        add_inline(par, stripped)
        par.alignment = WD_ALIGN_PARAGRAPH.LEFT
        i += 1

    if table_rows:
        flush_table(doc, table_rows)

    doc.save(DST)
    print(f"OK: {DST}")


if __name__ == "__main__":
    main()
