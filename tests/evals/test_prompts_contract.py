"""Промпты содержат обязательные запреты (проверка кодом, не LLM-судьёй)."""
from pathlib import Path

PROMPTS = Path(__file__).parents[2] / ".ai" / "prompts"

REQUIRED = {
    "brief.md": ["assumptions", "open_questions", "Не придумывай"],
    "geometry.md": ["confidence", "open_question", "Не меняй пропорции"],
    "cpr_review.md": ["Не пересчитывай смету", "passed"],
}


def test_prompts_keep_guard_phrases():
    for name, phrases in REQUIRED.items():
        text = (PROMPTS / name).read_text(encoding="utf-8")
        for phrase in phrases:
            assert phrase in text, f"{name}: нет '{phrase}'"
