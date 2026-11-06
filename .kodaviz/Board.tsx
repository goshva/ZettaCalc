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

const parts = [
  { id: "01", caption: "Раздел 1 · Записка владельцу", el: <Memo /> },
  { id: "02", caption: "Раздел 2 · Путь клиента", el: <Journey /> },
  { id: "03", caption: "Раздел 3 · Граница продукта и деньги", el: <Money /> },
  { id: "04", caption: "Раздел 4 · Архитектура и технический путь", el: <Architecture /> },
  { id: "05", caption: "Раздел 5 · Дорожная карта и бюджет", el: <Roadmap /> },
  { id: "06", caption: "Раздел 6 · Агенты: производство и маркетинг", el: <Agents /> },
  { id: "07", caption: "Раздел 7 · Мониторинг ИТ-инфраструктуры", el: <Monitoring /> },
  { id: "08", caption: "Раздел 8 · Вопросы к реализатору", el: <Questions /> },
  { id: "09", caption: "Раздел 9 · Риски", el: <Risks /> },
  { id: "10", caption: "Раздел 10 · Решение и договорённость", el: <Decision /> },
];

/* Полная записка одним полотном: для чтения и печати в PDF. */
export default function Board() {
  return (
    <div className="screen board">
      <header className="masthead" style={{ marginBottom: 10 }}>
        <div className="eyebrow">
          Аналитическая записка целиком · Fresh House / Zetta Home · CPTO · 06.11.2026
        </div>
        <h1 className="title">Полная записка: десять разделов одним полотном</h1>
        <p className="subtitle">
          Версия для чтения подряд и печати. Интерактивные элементы (вкладки, выбор шага,
          калькуляторы) работают здесь так же, но печатается только текущее состояние — для PDF по
          разделам используйте отдельные экраны.
        </p>
      </header>
      {parts.map((p) => (
        <section key={p.id} className="board-part" aria-label={p.caption}>
          <div className="board-caption">{p.caption}</div>
          {p.el}
        </section>
      ))}
    </div>
  );
}
