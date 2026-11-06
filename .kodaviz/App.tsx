import { useState } from "react";
import Memo from "./screens/01-memo";
import Journey from "./screens/02-journey";
import Money from "./screens/03-money";
import Architecture from "./screens/04-architecture";
import Roadmap from "./screens/05-roadmap";
import Agents from "./screens/06-agents";
import Monitoring from "./screens/07-monitoring";
import Questions from "./screens/08-questions";
import Risks from "./screens/09-risks";
import Decision from "./screens/10-decision";

const screens = [
  { id: "01", label: "1. Записка владельцу", el: <Memo /> },
  { id: "02", label: "2. Путь клиента", el: <Journey /> },
  { id: "03", label: "3. Граница продукта и деньги", el: <Money /> },
  { id: "04", label: "4. Архитектура", el: <Architecture /> },
  { id: "05", label: "5. Дорожная карта", el: <Roadmap /> },
  { id: "06", label: "6. Агенты: производство и маркетинг", el: <Agents /> },
  { id: "07", label: "7. Мониторинг ИТ-инфраструктуры", el: <Monitoring /> },
  { id: "08", label: "8. Вопросы к реализатору", el: <Questions /> },
  { id: "09", label: "9. Риски", el: <Risks /> },
  { id: "10", label: "10. Решение и договорённость", el: <Decision /> },
];

export default function App() {
  const [active, setActive] = useState("01");
  const current = screens.find((s) => s.id === active) ?? screens[0];

  return (
    <div style={{ height: "100%", display: "flex", flexDirection: "column", background: "var(--bg)" }}>
      <nav
        className="nav"
        aria-label="Навигация по записке"
        style={{
          padding: "10px 22px",
          borderBottom: "1px solid var(--line)",
          background: "var(--surface)",
          flex: "0 0 auto",
        }}
      >
        {screens.map((s) => (
          <button
            key={s.id}
            type="button"
            className="nav-chip"
            aria-pressed={s.id === active}
            style={
              s.id === active
                ? { borderStyle: "solid", borderColor: "var(--brand)", background: "var(--brand-soft)", color: "var(--brand)" }
                : undefined
            }
            onClick={() => setActive(s.id)}
          >
            {s.label}
          </button>
        ))}
      </nav>
      <div style={{ flex: "1 1 auto", minHeight: 0 }}>{current.el}</div>
    </div>
  );
}
