import { useState } from "react";
import { Card, Chip, Callout, SectionLabel, Masthead, Footer, Table } from "../ui";
import { Block } from "../charts";
import { layers, services, implemented, missing } from "../data";

export default function Architecture() {
  const [svc, setSvc] = useState(3);
  const s = services[svc];

  return (
    <div className="screen">
      <Masthead
        eyebrow="Экран 4 · Архитектура"
        title="Как устроено «чтобы считалось само»"
        subtitle="Технический путь одной фразой: данные проходят через несколько фильтров-ворот, и на каждом уровне остаётся только то, во что можно ставить цену. Нажмите на сервис — увидите ответственность и SLA."
      />

      <SectionLabel>Слои данных: от сырой выгрузки до цены для клиента</SectionLabel>
      <div className="stack">
        {layers.map((l, i) => (
          <div key={l.name} style={{ display: "grid", gridTemplateColumns: "minmax(0,1fr) 22px", gap: 8, alignItems: "start" }}>
            <div className={`layer ${l.tone}`}>
              <div className="layer-name">{l.name}</div>
              <div className="layer-items">
                {l.items.map((it) => (
                  <Chip key={it}>{it}</Chip>
                ))}
              </div>
            </div>
            <div style={{ paddingTop: 34, color: "var(--ink-3)" }}>{i < layers.length - 1 ? "↓" : ""}</div>
          </div>
        ))}
      </div>

      <div className="split-11">
        <Callout tone="warn" title="Ворота S2T — самая дешёвая страховка проекта">
          <p>
            Пока у поля каталога нет правила, примера и владельца, оно не попадает в рабочие данные.
            Это скучно и поначалу бесит: первая загрузка кухни займёт недели. Зато цена с
            перепутанными миллиметрами и метрами не уйдёт в смету клиенту.{" "}
            <span className="strong">Одна такая ошибка стоит дороже, чем полгода ворот.</span>
          </p>
        </Callout>
        <Callout tone="brand" title="Почему снимок цен, а не «текущая цена»">
          <p>
            ЦПР ссылается на снимок каталога и не меняется задним числом. Через полгода поставщик
            поднимет цену — а выданное клиенту documento останется ровно таким, каким его утверждали.
            Это защита и от претензий, и от спора с подрядчиком.
          </p>
        </Callout>
      </div>

      <SectionLabel>Сервисы и их обещания по скорости</SectionLabel>
      <div className="split-38">
        <Card tone="brand" title="Детерминированное ядро — без ИИ">
          <div className="svc-grid">
            {services.map((x, i) => (
              <button key={x.name} type="button" className="svc" aria-pressed={i === svc} onClick={() => setSvc(i)}>
                <span>
                  <span className="svc-name">{x.name}</span>
                  <span className="svc-sub">{x.sub}</span>
                </span>
                <Chip
                  tone={
                    x.state === "работает"
                      ? "teal"
                      : x.state === "скелет" || x.state === "1 техкарта" || x.state === "без чертежей" || x.state === "контракт"
                        ? "warn"
                        : "danger"
                  }
                >
                  {x.state}
                </Chip>
              </button>
            ))}
          </div>
        </Card>
        <div className="stack">
          <Card title={s.name} tone="brand">
            <div className="grid grid-2" style={{ gap: 8 }}>
              <Block label="SLA (цель v1)" value={s.sla} color="var(--brand-2)" />
              <Block label="Готовность" value={s.state} color="var(--teal)" />
            </div>
            <p className="small" style={{ marginTop: 10 }}>{s.sub}</p>
          </Card>
          <Callout tone="teal" title="Что означает SLA для клиента">
            <p>
              Расчёт должен приходить быстрее 3 секунд, ЦПР — собираться меньше минуты. Иначе
              посетитель закроет вкладку до того, как увидит цифру. Это не «приятно бы», а условие
              работы воронки.
            </p>
          </Callout>
          <Callout tone="violet" title="ИИ живёт снаружи ядра">
            <p>
              Единственная точка обращения к модели — harness со своими бюджетами, трейсом и
              ограничениями. Ядро про ИИ не знает. Если модель начнёт выдумывать артикулы, она упрётся
              в guardrails и тесты, а не в смету клиента.
            </p>
          </Callout>
        </div>
      </div>

      <SectionLabel>Что уже работает против того, что только описано</SectionLabel>
      <div className="grid grid-2">
        <Card tone="teal" title="Есть в коде (проверяется запуском)">
          <Table
            compact
            head={["Модуль", "Что делает"]}
            rows={implemented.map((x) => [
              <span className="strong">{x.name}</span>,
              <span>
                {x.note} <Chip tone={x.state === "работает" ? "teal" : x.state === "локально" ? "warn" : "warn"}>{x.state}</Chip>
              </span>,
            ])}
          />
        </Card>
        <Card tone="danger" title="Нет в коде (только в документах)">
          <Table
            compact
            head={["Разрыв", "Чем это грозит"]}
            rows={missing.map((x) => [
              <span className="strong">{x.name}</span>,
              <span>{x.impact}</span>,
            ])}
          />
        </Card>
      </div>

      <Footer>
        Экран 4 из 10 · Контракты: OpenAPI на 13 операций, JSON Schema ЦПР, два примера паспорта.
        Проверки: ruff, mypy strict, 66 тестов.
      </Footer>
    </div>
  );
}
