import { useState } from "react";
import { Card, Chip, Callout, SectionLabel, Masthead, Footer, Table } from "../ui";
import { Block, LineChart } from "../charts";
import { monitor } from "../data";

/* Детали по каждому уровню наблюдения. */
const details: Record<
  string,
  { why: string; kpis: { label: string; value: string }[]; owner: string; when: string; risk: string }
> = {
  Сервисы: {
    why: "Клиент не ждёт: если расчёт собирается дольше 3 секунд, сессия срывается до того, как он увидел цифру. Это напрямую бьёт по метрике «доля расчётов без сотрудника».",
    kpis: [
      { label: "Целевая латентность расчёта", value: "p95 < 3 с" },
      { label: "Допустимая доля ошибок", value: "< 0,5%" },
      { label: "Глубина хранения метрик", value: "13 мес." },
    ],
    owner: "DevOps + бэкенд",
    when: "волна 0, до первого прода",
    risk: "риск №1: тихий сбой считалки даёт клиенту пустой или неверный итог",
  },
  "Базы данных": {
    why: "Один медленный запрос в подборе техкарт превращает расчёт в 30 секунд. Плюс база — источник цены: рассинхрон снимков цен заметен клиенту в деньгах.",
    kpis: [
      { label: "Медленные запросы", value: "> 1 с в алерте" },
      { label: "Проверка снимка цен", value: "ежесуточно" },
      { label: "Репликация", value: "отставание < 60 с" },
    ],
    owner: "дата-инженер",
    when: "волна 0",
    risk: "риск №2: устаревшая цена уходит в смету клиента",
  },
  "Поток данных": {
    why: "Ночная выгрузка каталога не прошла — и весь день расчёты идут по вчерашним ценам. Простой в 24 часа здесь дороже простоя сайта.",
    kpis: [
      { label: "Задержка ODS→DDS", value: "целевая < 4 ч" },
      { label: "S2T-отказы", value: "0 без ручной приёмки" },
      { label: "Dead-letter", value: "алерт при первом элементе" },
    ],
    owner: "дата-инженер",
    when: "волна 0–1",
    risk: "риск №2 и №3: пустой или кривой каталог",
  },
  "ИИ-агенты": {
    why: "Агент может уйти в цикл и сжечь бюджет за ночь. Поэтому стоимость, длина и зацикливание каждой сессии видны как отдельный дашборд, а не в счёте от провайдера постфактум.",
    kpis: [
      { label: "Потолок сессии", value: "20 шагов / 50 000 токенов / 50 ₽" },
      { label: "Стоимость на проект", value: "контролируется трейсом" },
      { label: "anti-loop", value: "3 повтора = стоп" },
    ],
    owner: "CPTO + ML-инженер",
    when: "волна 0 (механизм уже в коде)",
    risk: "риск №3: агент додумывает — трейс это показывает",
  },
  "Бизнес-метрики": {
    why: "Технически всё может быть зелено, а продукт уже врёт: план расходится с фактом. Эти метрики — единственные, которые владелец читает каждую неделю.",
    kpis: [
      { label: "Доля расчётов без сотрудника", value: "≥ 40%" },
      { label: "Авто-ЦПР без правок (pass^k)", value: "50–60% в v1" },
      { label: "План/факт сметы", value: "в заявленном диапазоне" },
    ],
    owner: "CPTO",
    when: "волна 1, с первого реального проекта",
    risk: "риск №1 и №7: ложная точность и недостижимая цель 90%",
  },
  "Серверы и контейнеры": {
    why: "Самый скучный уровень, и именно он объясняет 80% ночных инцидентов: закончился диск, кончилась память, упал контейнер.",
    kpis: [
      { label: "Диск", value: "алерт при 85%" },
      { label: "Память контейнера", value: "алерт при 90% лимита" },
      { label: "Ротация логов трейса", value: "включена" },
    ],
    owner: "DevOps",
    when: "волна 0",
    risk: "риск №5: логи содержат данные клиента — ротация и доступ обязательны",
  },
};

const alertFlow = [
  { name: "Ядро считает", sub: "estimate / rule engine / cpr build", tone: "layer--core" },
  { name: "Trace каждого вызова", sub: "cpr_id, version, duration_ms, assumptions_count", tone: "" },
  { name: "Prometheus собирает", sub: "латентность, ошибки, ₽/сессия, бизнес-метрики", tone: "" },
  { name: "Grafana показывает", sub: "три дашборда: продукт, данные, ИИ", tone: "" },
  { name: "Алерт → дежурный", sub: "триггер из реестра рисков = правило алерта", tone: "layer--gate" },
];

export default function Monitoring() {
  const [sel, setSel] = useState(3);
  const level = monitor[sel];
  const d = details[level.level] ?? details["Сервисы"];

  return (
    <div className="screen">
      <Masthead
        eyebrow="Экран 7 · Мониторинг ИТ-инфраструктуры предприятия"
        title="Предприятие должно видеть, что система считает правильно"
        subtitle="Наблюдаемость — не «прикрутим потом». Триггеры рисков из реестра превращаются в правила алертов: если порог пройден, об этом узнаёт человек, а не клиент."
      />

      <div className="split-38">
        <Card tone="violet" title="Шесть уровней наблюдения">
          <div className="svc-grid">
            {monitor.map((m, i) => (
              <button key={m.level} type="button" className="svc" aria-pressed={i === sel} onClick={() => setSel(i)}>
                <span>
                  <span className="svc-name">{m.level}</span>
                  <span className="svc-sub">{m.what}</span>
                </span>
                <Chip tone={i === sel ? "brand" : "flat"}>подробно</Chip>
              </button>
            ))}
          </div>
        </Card>

        <div className="stack">
          <Card title={level.level} tone="brand">
            <p className="small">{d.why}</p>
            <div className="stack-8" style={{ marginTop: 10 }}>
              {d.kpis.map((k) => (
                <Block key={k.label} label={k.label} value={k.value} color="var(--brand-2)" />
              ))}
            </div>
            <div className="grid grid-2" style={{ gap: 8, marginTop: 10 }}>
              <Block label="Кто отвечает" value={d.owner} color="var(--teal)" />
              <Block label="Когда вводим" value={d.when} color="var(--warn)" />
            </div>
            <p className="xs warn-text" style={{ marginTop: 8 }}>
              Связанный риск: {d.risk}
            </p>
          </Card>
          <Callout tone="warn" title="Правило: триггер риска = правило алерта">
            <p className="small">
              Риск №1 в реестре говорит «план/факт вне диапазона в 20% проектов». В мониторинге это
              буквально запрос к витрине и алерт на CPTO. Реестр рисков перестаёт быть документом и
              становится работающей сигнализацией.
            </p>
          </Callout>
        </div>
      </div>

      <SectionLabel>Как сигнал проходит от расчёта до человека</SectionLabel>
      <div className="flow">
        {alertFlow.map((f, i) => (
          <div key={f.name} style={{ display: "contents" }}>
            <div className={`flow-node ${f.tone}`}>
              <div className="flow-node-name">{f.name}</div>
              <div className="flow-node-sub">{f.sub}</div>
            </div>
            {i < alertFlow.length - 1 ? <div className="flow-arrow">→</div> : null}
          </div>
        ))}
      </div>

      <SectionLabel>Что видно на экране дежурного</SectionLabel>
      <div className="grid grid-2">
        <Card title="Латентность расчёта, p95 (цель — 3 с)">
          <LineChart
            xLabels={["нед. 1", "нед. 3", "нед. 6", "нед. 9", "нед. 12", "нед. 14"]}
            yMax={6}
            yUnit=" с"
            series={[
              { name: "p95 расчёт", color: "var(--brand-2)", values: [4.6, 3.8, 3.1, 2.6, 2.2, 1.9] },
              { name: "p95 сборка ЦПР", color: "var(--violet)", values: [5.4, 5.0, 4.4, 3.6, 3.0, 2.6] },
              { name: "цель", color: "var(--danger)", values: [3, 3, 3, 3, 3, 3], dashed: true },
            ]}
          />
          <p className="xs muted" style={{ marginTop: 6 }}>
            Форма кривых — иллюстрация того, как должен выглядеть прогресс, а не измеренные данные:
            замер начинается на staging в волне 0.
          </p>
        </Card>

        <Card tone="teal" title="Стоимость ИИ-сессий на проект">
          <Table
            compact
            head={["Что", "Механизм контроля", "Где в коде"]}
            rows={[
              ["Потолок шагов", "runtime останавливает цикл", "src/harness/runtime.py"],
              ["Потолок токенов", "сумма по трейсу, стоп при превышении", "то же"],
              ["Потолок рублей", "бюджет на задачу, фиксация stop_reason", "то же"],
              ["Зацикливание", "3 одинаковых вызова = anti_loop", "то же"],
              ["Запись без подтверждения", "write_denied, dry-run, Idempotency-Key", "guardrails + OpenAPI"],
            ]}
          />
          <p className="xs muted" style={{ marginTop: 6 }}>
            Ограничения уже написаны и покрыты тестами — не «запланируем мониторинг агентов», а
            снимаем с них метрики.
          </p>
        </Card>
      </div>

      <SectionLabel>Что делаем в первую очередь</SectionLabel>
      <div className="grid grid-3">
        <Card tone="teal" kicker="волна 0 · до первого клиента">
          <ul className="list list--tight">
            <li>health-эндпоинты и метрики по каждому сервису</li>
            <li>трейс каждого расчёта: длительность, число допущений</li>
            <li>алерты на поток данных: выгрузка каталога прошла/нет</li>
            <li>учёт ₽ и токенов на сессию агента</li>
            <li>доступ к логом с данными клиента — только по ACL</li>
          </ul>
        </Card>
        <Card tone="brand" kicker="волна 1 · пилот Zetta Home">
          <ul className="list list--tight">
            <li>дашборд продукта: доля расчётов, путь до первого итога</li>
            <li>дашборд надёжности: pass^k, доля pending_review</li>
            <li>план/факт по первым ремонтам — главный алерт владельца</li>
            <li>агентские эффекты производства в тех же метриках</li>
          </ul>
        </Card>
        <Card tone="violet" kicker="волна 2 · масштаб">
          <ul className="list list--tight">
            <li>сводные витрины по категориям и подрядчикам</li>
            <li>Text2SQL по семантическому слою, только SELECT</li>
            <li>качественная оценка выдачи (LLM-as-judge) с калибровкой</li>
            <li>бюджеты на агента как статья расходов, а не «когда-нибудь»</li>
          </ul>
        </Card>
      </div>

      <Callout tone="danger" title="Чего не будет в этой системе">
        <p className="small">
          Мониторинг не «додумывает» показатели и не показывает точность там, где её нет: каждая
          метрика либо измерена, либо явно помечена как гипотеза до замера. Дашборд, который
          приукрашивает надёжность, опаснее отсутствия дашборда.
        </p>
      </Callout>

      <Footer>Экран 7 из 10 · Далее — вопросы владельцу, без которых план остаётся оценкой.</Footer>
    </div>
  );
}
