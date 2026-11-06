import { useState } from "react";
import { Card, Callout, SectionLabel, Masthead, Footer, Table } from "../ui";
import { Block } from "../charts";
import { waves, monitor } from "../data";

export default function Roadmap() {
  const [wave, setWave] = useState(0);
  const w = waves[wave];

  return (
    <div className="screen">
      <Masthead
        eyebrow="Экран 5 · План"
        title="Куда и с какой скоростью едем"
        subtitle="Три волны: фундамент, путь клиента, масштаб. Каждая — с критериями готовности, бюджетом и артефактами. Нажмите на волну — увидите детали."
      />

      <SectionLabel>Выберите волну</SectionLabel>
      <div className="waves">
        {waves.map((ww, i) => (
          <button
            key={ww.name}
            type="button"
            className="wave"
            aria-pressed={i === wave}
            onClick={() => setWave(i)}
          >
            <span className="wave-name">{ww.name}</span>
            <span className="wave-meta">
              {ww.weeks} · {ww.budget} · цель: {ww.goal}
            </span>
          </button>
        ))}
      </div>

      <Card title={w.name} tone={wave === 0 ? "teal" : wave === 1 ? "brand" : "violet"}>
        <div className="grid grid-3" style={{ gap: 8, marginBottom: 10 }}>
          <Block label="Срок" value={w.weeks} color="var(--brand-2)" />
          <Block label="Бюджет" value={w.budget} color="var(--teal)" />
          <Block label="Цель" value={w.goal} color="var(--violet)" />
        </div>
        <h3 style={{ fontSize: 13, fontWeight: 700, marginBottom: 6 }}>DoD (критерии готовности)</h3>
        <ul className="list">
          {w.dod.map((d) => (
            <li key={d}>✅ {d}</li>
          ))}
        </ul>
      </Card>

      <SectionLabel>Где вписаны новые направления</SectionLabel>
      <div className="split-38">
        <Card tone="flat">
          <Table
            compact
            head={["Направление", "Волна", "Почему не раньше"]}
            rows={[
              ["Harness + guardrails", "0", "Нужен до ИИ-агентов — это ограждения"],
              ["S2T, PIM, 10 техкарт", "0", "Фундамент без данных = трата времени"],
              ["API + UI калькулятора", "1", "Путь клиента не работает без сайта"],
              ["ЦПР + чертежи + аукцион", "1", "Продукт для продажи"],
              ["Агенты производства", "1", "Используют PIM и техкарты волны 0"],
              ["Агенты маркетинга", "1", "Контент и сегментация параллельны"],
              ["Мониторинг ИТ", "0–1", "Нужен с первого дня пилота"],
              ["Text2SQL, LLM-as-judge", "2", "Зрелые сценарии ИИ"],
            ]}
          />
        </Card>
        <div className="stack">
          <Callout tone="brand" title="Параллельные работы волны 0">
            <p>
              Пока ИТ строит ядро, дизайн готовит контент референсов и вариантов, юрисконсульт —
              шаблоны ЦПР и договоры с первыми подрядчиками. Это не тормозит, а идёт параллельно.
            </p>
          </Callout>
          <Callout tone="warn" title="Узкое место: техкарты">
            <p>
              Скорость появления 10 техкарт — это не ИТ-проблема, а проблема экспертов. ИИ делает
              черновики, но утверждение — за технологом. Если к неделе 6 нет 10 карт — весь план
              сдвигается.
            </p>
          </Callout>
          <Callout tone="teal" title="Мониторинг — с первого дня">
            <p>
              Не ждём пилота. Prometheus + Grafana ставятся на staging сразу после волны 0. Триггеры
              из реестра рисков — как алерты, а не как документ.
            </p>
          </Callout>
        </div>
      </div>

      <SectionLabel>Мониторинг ИТ-инфраструктуры</SectionLabel>
      <Card tone="violet">
        <Table
          compact
          head={["Уровень", "Что мониторим", "Инструмент", "Алерт → триггер"]}
          rows={monitor.map((m) => [
            <span className="strong">{m.level}</span>,
            m.what,
            <code style={{ fontSize: 11 }}>{m.tool}</code>,
            <span className="xs warn-text">{m.alert}</span>,
          ])}
        />
        <p className="xs muted" style={{ marginTop: 8 }}>
          Каждый вызов ядра логирует trace: cpr_id, version, duration_ms, assumptions_count. Это
          позволяет строить дашборды надёжности и находить узкие места в подборе техкарт.
        </p>
      </Card>

      <SectionLabel>Кто делает что: RACI</SectionLabel>
      <Card tone="flat">
        <Table
          compact
          head={["Работа", "CPTO", "Эксперты", "Дизайн", "Коммерция", "Подрядчик", "Юрист"]}
          rows={[
            ["Приоритеты v1", "A", "C", "C", "C", "I", "I"],
            ["Техкарты", "C", "A/R", "I", "I", "C", "I"],
            ["Каталог и цены", "C", "I", "I", "A/R", "I", "I"],
            ["Ядро и ЦПР", "A/R", "C", "I", "I", "I", "I"],
            ["Агенты производства", "A/R", "C", "I", "I", "I", "I"],
            ["Агенты маркетинга", "A/R", "I", "C", "C", "I", "I"],
            ["Мониторинг ИТ", "A/R", "I", "I", "I", "C", "I"],
            ["Шаблон договора", "C", "C", "I", "C", "C", "A/R"],
            ["Условия аукциона", "C", "I", "I", "A", "R", "C"],
            ["Метрики и отчёт", "A/R", "I", "I", "C", "I", "I"],
          ]}
        />
        <p className="xs muted" style={{ marginTop: 8 }}>
          R = исполняет, A = отвечает, C = консультирует, I = в курсе. CPTO несёт ответственность за
          ядро, агенты и мониторинг одновременно.
        </p>
      </Card>

      <Footer>
        Экран 5 из 10 · Бюджеты — оценка из ответа CPTO. Требуют согласования с финансами компании.
      </Footer>
    </div>
  );
}
