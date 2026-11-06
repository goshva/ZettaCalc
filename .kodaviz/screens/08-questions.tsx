import { useState } from "react";
import { Card, Chip, Callout, SectionLabel, Masthead, Footer } from "../ui";
import { questions } from "../data";

const areaTone: Record<string, "brand" | "teal" | "warn" | "danger" | "violet"> = {
  "Экономика": "brand",
  "Организация": "warn",
  "Данные": "teal",
  "Экспертиза": "violet",
  "ИТ": "teal",
  "Подрядчики": "warn",
  "Правовое": "danger",
};

export default function Questions() {
  const [asked, setAsked] = useState<Record<string, boolean>>({});
  const [open, setOpen] = useState<string | null>("q3");
  const done = Object.values(asked).filter(Boolean).length;

  return (
    <div className="screen">
      <Masthead
        eyebrow="Экран 8 · Вопросы владельцу и бизнесу"
        title="Что нужно узнать, чтобы план не оказался фантазией"
        subtitle="Восемь вопросов, каждый из которых блокирует конкретное решение. Отмечайте галочкой те, на которые уже есть ответ — прогресс сверху покажет, готовы ли вы к старту."
      />

      <Card tone={done >= 6 ? "teal" : done >= 3 ? "warn" : "danger"}>
        <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", gap: 12, flexWrap: "wrap" }}>
          <div>
            <div className="strong">Отвечено {done} из {questions.length}</div>
            <p className="xs muted" style={{ marginTop: 3 }}>
              {done < 3
                ? "Меньше трёх ответов — бюджет и сроки остаются оценкой «из воздуха», а не планом."
                : done < 6
                  ? "Половина ответов есть: волну 0 начинать можно, волну 1 — только после остальных."
                  : "Достаточно, чтобы зафиксировать волну 0 в календаре и согласовать бюджет."}
            </p>
          </div>
          <div className="progress" style={{ minWidth: 220 }}>
            <div className="bar-track" style={{ width: 220 }}>
              <div
                className="bar-fill"
                style={{
                  width: `${(done / questions.length) * 100}%`,
                  background: done >= 6 ? "var(--teal)" : done >= 3 ? "var(--warn)" : "var(--danger)",
                }}
              />
            </div>
            <span className="num">{Math.round((done / questions.length) * 100)}%</span>
          </div>
        </div>
      </Card>

      <SectionLabel>Список вопросов</SectionLabel>
      <div className="stack-8">
        {questions.map((q) => {
          const isOpen = open === q.id;
          const isAsked = !!asked[q.id];
          return (
            <div key={q.id} className="card" style={{ padding: 0, overflow: "hidden" }}>
              <div style={{ display: "grid", gridTemplateColumns: "auto minmax(0,1fr) auto", gap: 12, alignItems: "center", padding: "11px 14px" }}>
                <button
                  type="button"
                  className="qrow"
                  style={{ margin: 0, border: "none", padding: 0, background: "transparent" }}
                  aria-pressed={isAsked}
                  aria-label={`Отметить вопрос как закрытый: ${q.q}`}
                  onClick={() => setAsked((s) => ({ ...s, [q.id]: !s[q.id] }))}
                >
                  <span className="qbox">{isAsked ? "✓" : ""}</span>
                </button>
                <button
                  type="button"
                  onClick={() => setOpen(isOpen ? null : q.id)}
                  style={{ background: "transparent", border: "none", font: "inherit", textAlign: "left", cursor: "pointer", color: "var(--ink)" }}
                >
                  <span className="strong" style={{ fontSize: 13.5 }}>{q.q}</span>
                  {isOpen ? (
                    <span className="small" style={{ display: "block", marginTop: 6, color: "var(--ink-2)" }}>
                      <span className="strong">Почему спрашиваю:</span> {q.why}
                      <br />
                      <span className="strong">Что блокирует:</span> {q.blocks}
                    </span>
                  ) : null}
                </button>
                <div style={{ display: "flex", gap: 6, alignItems: "center", flexWrap: "wrap", justifyContent: "flex-end" }}>
                  <Chip tone={areaTone[q.area] ?? "brand"}>{q.area}</Chip>
                  <Chip tone={isAsked ? "teal" : "flat"}>{isAsked ? "ответ есть" : "открыт"}</Chip>
                </div>
              </div>
            </div>
          );
        })}
      </div>

      <SectionLabel>Что из ответов следует для решения о найме</SectionLabel>
      <div className="grid grid-3">
        <Callout tone="brand" title="Если ответы на 1–2 есть">
          <p className="small">
            Можно считать экономику и согласовать цену ЦПР. Без них любая цифра в плане — украшение
            доклада, которое нельзя защитить перед советом директоров.
          </p>
        </Callout>
        <Callout tone="warn" title="Если ответа на 3 нет">
          <p className="small">
            Не начинайте волну 1. Гибрид без договорённости о делении выручки превратится в
            внутренний конфликт, и проект закроют не из-за технологии.
          </p>
        </Callout>
        <Callout tone="violet" title="Если нет ответов на 4–5">
          <p className="small">
            Волна 0 растянется с 6 недель на 12–14: сначала сбор данных и договорённости с
            технологами, потом разработка. Честнее назвать срок сразу.
          </p>
        </Callout>
      </div>

      <Footer>Экран 8 из 10 · Далее — риски и что пересмотреть в стратегии.</Footer>
    </div>
  );
}
