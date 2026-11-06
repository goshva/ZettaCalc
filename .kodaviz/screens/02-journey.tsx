import { useState } from "react";
import { Card, Chip, Callout, SectionLabel, Masthead, Footer } from "../ui";
import { LineChart, Block } from "../charts";
import { journey, funnel } from "../data";

const toneByState: Record<string, string> = {
  "работает": "var(--teal)",
  "частично": "var(--warn)",
  "описано": "var(--brand-2)",
  "нет в коде": "var(--danger)",
  "контракт есть": "var(--violet)",
};

export default function Journey() {
  const [sel, setSel] = useState(4);
  const step = journey[sel];

  return (
    <div className="screen">
      <Masthead
        eyebrow="Экран 2 · Продукт"
        title="Путь клиента: от идеи интерьера до заказа"
        subtitle="Нажмите на шаг — увидите, кто его выполняет, сколько это занимает и готово ли это сегодня в коде."
      />

      <div className="split-38">
        <Card tone="brand" title="Клиентский путь (стратегия, раздел 3)">
          <div className="steps">
            {journey.map((s, i) => (
              <button
                key={s.name}
                type="button"
                className="step"
                aria-pressed={i === sel}
                onClick={() => setSel(i)}
              >
                <span className="step-no">{String(i + 1).padStart(2, "0")}</span>
                <span>
                  <span className="step-name">{s.name}</span>
                  <span className="step-sub">{s.sub}</span>
                </span>
                <span className="step-meta">
                  <Chip tone={s.cost === "бесплатно" ? "free" : "paid"}>{s.cost}</Chip>
                </span>
              </button>
            ))}
          </div>
          <p className="xs muted" style={{ marginTop: 8 }}>
            Цвет слева показывает готовность: бирюзовый — работает в коде, красный — только в
            документах.
          </p>
        </Card>

        <div className="stack">
          <Card title={`Шаг ${sel + 1}. ${step.name}`} tone={step.cost === "бесплатно" ? "teal" : "violet"}>
            <div className="grid grid-2" style={{ gap: 8 }}>
              <Block label="Кто выполняет" value={step.who} color="var(--brand-2)" />
              <Block label="Время (цель)" value={step.time} color="var(--teal)" />
            </div>
            <div style={{ marginTop: 8 }}>
              <Block
                label="Готовность сегодня"
                value={
                  <span style={{ color: toneByState[step.state] ?? "var(--ink)" }}>{step.state}</span>
                }
                color={toneByState[step.state] ?? "var(--line-2)"}
              />
            </div>
            <p className="small" style={{ marginTop: 10 }}>{step.sub}</p>
          </Card>

          <Callout tone="warn" title="Где обрыв">
            <p>
              Шаги 2–4 (варианты реализации, бриф, планировка) есть в стратегии и в прототипах, но{" "}
              <span className="strong">не сделаны в коде</span>. Именно они превращают красивый
              прототип в продукт: без них расчёт невозможно пересчитать при изменении выбора клиента.
            </p>
          </Callout>

          <Callout tone="brand" title="Что должно удивить клиента на первом визите">
            <p>
              Он бесплатно видит <span className="strong">артикулы, бренды и цены по позициям</span>,
              состав работ и диапазон неопределённости. Обычно эту информацию приходится выбивать
              месяцами через сметчиков.
            </p>
          </Callout>
        </div>
      </div>

      <SectionLabel>Воронка: сколько людей доходит до конца</SectionLabel>
      <div className="split-38">
        <Card title="Воронка волны 1 (целевые доли, %)">
          <div className="funnel">
            {funnel.map((f, i) => (
              <div key={f.name} className="funnel-row">
                <span className="funnel-label">{f.name}</span>
                <span>
                  <span
                    className="funnel-band"
                    style={{
                      display: "block",
                      width: `${f.value}%`,
                      background: i >= 4 ? "var(--violet)" : "var(--brand-2)",
                      opacity: 0.55 + 0.45 * (f.value / 100),
                    }}
                  />
                </span>
                <span className="funnel-val num">{f.value}%</span>
              </div>
            ))}
          </div>
          <p className="xs muted" style={{ marginTop: 8 }}>
            40% — целевая метрика из ответа CPTO. Ниже строки «купил ЦПР» и «заказ» — гипотезы без
            исходных данных: их измеряем в первые 4 недели пилота, а не обещаем.
          </p>
        </Card>

        <Card tone="violet" title="Как метрики должны вырасти за 26 недель">
          <LineChart
            xLabels={["старт", "нед. 6", "нед. 14", "нед. 26"]}
            yMax={100}
            series={[
              { name: "Доля авто-ЦПР", color: "var(--brand-2)", values: [0, 20, 55, 68] },
              { name: "Сессии с расчётом", color: "var(--teal)", values: [0, 15, 40, 52] },
              { name: "Цель стратегии 90%", color: "var(--ink-3)", values: [90, 90, 90, 90], dashed: true },
            ]}
          />
          <div className="legend" style={{ marginTop: 6 }}>
            <span className="legend-item"><span className="legend-swatch" style={{ background: "var(--brand-2)" }} />авто-ЦПР без правок</span>
            <span className="legend-item"><span className="legend-swatch" style={{ background: "var(--teal)" }} />расчёты без сотрудника</span>
            <span className="legend-item"><span className="legend-swatch" style={{ background: "var(--ink-3)" }} />цель 90% — ориентир, не обещание v1</span>
          </div>
          <p className="small muted" style={{ marginTop: 8 }}>
            Рост плавный, а не вертикальный: он ограничен скоростью появления техкарт и планировок, а
            не желанием команды.
          </p>
        </Card>
      </div>

      <SectionLabel>Что берём из прототипов А и В</SectionLabel>
      <div className="grid grid-3">
        <Card tone="flat">
          <div className="card-kicker">Из прототипа А</div>
          <p className="small">
            Каталог и карточки товаров, интерьерный центр и блок доверия, вход «загрузи свой
            проект», FAQ про замер и смету. Плюс структура разделов.
          </p>
        </Card>
        <Card tone="flat">
          <div className="card-kicker">Из прототипа В</div>
          <p className="small">
            Ядро пути: квартиры со сметами из реального архива Zetta Home, аукцион со сравнением цены,
            сроков и рейтинга, кабинет. И лёгкий вес файла — основа для сайта.
          </p>
        </Card>
        <Card tone="warn">
          <div className="card-kicker">Нет ни в одном прототипе</div>
          <p className="small">
            Ввод размеров, выбор планировки, экран брифа с допущениями, 3–4 варианта реализации с
            пересчётом, SKU и цены на странице, диапазон неопределённости. Это содержание волны 1.
          </p>
        </Card>
      </div>

      <Footer>Экран 2 из 10 · Далее — что именно продаём: расчёт, базовый ЦПР и опции.</Footer>
    </div>
  );
}
