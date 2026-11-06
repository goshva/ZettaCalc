import { useState } from "react";
import { Card, Chip, Callout, SectionLabel, Masthead, Footer, Table } from "../ui";
import { risks } from "../data";

type Tone = "hi" | "mid" | "lo";

const toneFor = (r: { p: number; i: number }): Tone =>
  r.p === 3 && r.i === 3 ? "hi" : r.p === 3 || r.i === 3 ? "mid" : "lo";

const toneChip: Record<Tone, "danger" | "warn" | "teal"> = { hi: "danger", mid: "warn", lo: "teal" };

/* Клетки: строка = вероятность (3 сверху), столбец = влияние (3 справа). */
function layout() {
  const cells = new Map<number, typeof risks>();
  risks.forEach((r) => {
    const row = 3 - r.p; // 0 = высокая вероятность
    const col = r.i - 1; // 2 = высокое влияние
    const key = row * 3 + col;
    cells.set(key, [...(cells.get(key) ?? []), r]);
  });

  const placed: { r: (typeof risks)[number]; left: number; top: number }[] = [];
  for (const [key, list] of cells) {
    const row = Math.floor(key / 3);
    const col = key % 3;
    const cx = ((col + 0.5) / 3) * 100;
    const cy = ((row + 0.5) / 3) * 100;
    const cols = Math.min(list.length, 3);
    const rowsNeeded = Math.ceil(list.length / cols);
    list.forEach((r, idx) => {
      const c = idx % cols;
      const rw = Math.floor(idx / cols);
      const dx = (c - (cols - 1) / 2) * 7.5;
      const dy = (rw - (rowsNeeded - 1) / 2) * 9;
      placed.push({ r, left: cx + dx, top: cy + dy });
    });
  }
  return placed;
}

export default function Risks() {
  const [sel, setSel] = useState<number | null>(1);
  const placed = layout();
  const active = risks.find((r) => r.n === sel);

  return (
    <div className="screen">
      <Masthead
        eyebrow="Экран 9 · Риски"
        title="Где эта история может стоить денег"
        subtitle="Десять рисков из ответа CPTO. Цвет точки — комбинация вероятности и влияния; красный верхний правый угол — то, что реально убивает проект."
      />

      <div className="split-38">
        <Card tone="warn" title="Матрица: вероятность × влияние">
          <div className="matrix-wrap">
            <div className="matrix-y">
              <span>вероятность →</span>
            </div>
            <div className="matrix">
              <div className="matrix-band b-hi" />
              <div className="matrix-band b-mid" />
              <div className="matrix-band b-lo" />
              {placed.map(({ r, left, top }) => (
                <button
                  key={r.n}
                  type="button"
                  className={`dot t-${toneFor(r)}`}
                  style={{ left: `${left}%`, top: `${top}%`, width: 27, height: 27 }}
                  aria-pressed={sel === r.n}
                  aria-label={`Риск ${r.n}: ${r.name}`}
                  onClick={() => setSel(r.n === sel ? null : r.n)}
                >
                  {r.n}
                </button>
              ))}
            </div>
          </div>
          <div className="matrix-x">
            <span>влияние: низкое</span>
            <span>среднее</span>
            <span>высокое</span>
          </div>
          <div className="legend" style={{ marginTop: 10 }}>
            <span className="legend-item"><span className="legend-swatch" style={{ background: "var(--danger)" }} />высокая вероятность и влияние</span>
            <span className="legend-item"><span className="legend-swatch" style={{ background: "var(--warn)" }} />одно из двух высокое</span>
            <span className="legend-item"><span className="legend-swatch" style={{ background: "var(--teal)" }} />умеренный</span>
          </div>
        </Card>

        <div className="stack">
          {active ? (
            <Card tone={toneChip[toneFor(active)] === "danger" ? "danger" : "warn"} title={`Риск №${active.n}. ${active.name}`}>
              <div className="grid grid-2" style={{ gap: 8, marginBottom: 8 }}>
                <Card tone="flat">
                  <div className="card-kicker">Вероятность</div>
                  <div className="stat-value" style={{ fontSize: 21 }}>
                    {["", "низкая", "средняя", "высокая"][active.p]}
                  </div>
                </Card>
                <Card tone="flat">
                  <div className="card-kicker">Влияние</div>
                  <div className="stat-value" style={{ fontSize: 21 }}>
                    {["", "низкое", "среднее", "высокое"][active.i]}
                  </div>
                </Card>
              </div>
              <p className="small">
                <span className="strong">Как снижаем:</span> {active.mit}
              </p>
              <p className="small">
                <span className="strong">Видимый сигнал:</span> {active.trig}
              </p>
              <p className="small">
                <span className="strong">Кто отвечает:</span> {active.owner}
              </p>
            </Card>
          ) : (
            <Callout tone="brand" title="Выберите риск на матрице">
              <p className="small">
                Ниже — общий вывод. Каждый риск имеет измеримый триггер: не «может быть плохо», а
                «если к неделе 6 техкарт меньше 10 — стоп и разбор».
              </p>
            </Callout>
          )}

          <Callout tone="danger" title="Красная зона: четыре риска, которые решаются не кодом">
            <ul className="list list--tight small">
              <li>№1 ложная точность — вопрос доверия клиента и претензий</li>
              <li>№2 данные каталога — вопрос порядка в торговле</li>
              <li>№6 эксперты на техкарты — вопрос управления людьми</li>
              <li>№8 салон против самообслуживания — вопрос мотивации и KPI</li>
            </ul>
            <p className="small" style={{ marginTop: 7 }}>
              Реализатор снижает их архитектурой, но закрыть может только решение владельца.
            </p>
          </Callout>
        </div>
      </div>

      <SectionLabel>Полный реестр</SectionLabel>
      <Card>
        <Table
          compact
          head={["#", "Риск", "В × Д", "Как снижаем", "Сигнал", "Ответственный"]}
          rows={risks.map((r) => [
            <Chip tone={toneChip[toneFor(r)]}>{r.n}</Chip>,
            r.name,
            <span className="num">{r.p} × {r.i}</span>,
            <span className="small">{r.mit}</span>,
            <span className="small">{r.trig}</span>,
            <span className="small">{r.owner}</span>,
          ])}
        />
      </Card>

      <SectionLabel>Что в реестре самое неприятное</SectionLabel>
      <div className="grid grid-3">
        <Callout tone="danger" title="№6 · Эксперты — узкое место">
          <p>
            Автоматизация упирается не в ИТ, а в <span className="strong">время технологов</span>.
            Одна техкарта — это день-два эксперта. Десять карт за шесть недель без этого ресурса
            нереальны. Компромисс: ИИ пишет черновик, эксперт утверждает и подписывает.
          </p>
        </Callout>
        <Callout tone="danger" title="№8 · Салон против сайта">
          <p>
            Менеджер зарабатывает на сопровождении. Если сайт выдаёт расчёт без него, он теряет
            контроль над цифрой и начинает её обесценивать. Решение — гибрид и <span className="strong">общие
            KPI</span>: проект, начатый онлайн, засчитывается салону.
          </p>
        </Callout>
        <Callout tone="warn" title="№1 · Ложная точность">
          <p>
            Клиент принимает предварительную цифру за итоговую. Защита — диапазон, список допущений
            и обязательная пометка, что уточнится после проверки объекта. Дешевле сказать «не знаю,
            вот диапазон», чем получить претензию.
          </p>
        </Callout>
      </div>

      <Footer>Экран 9 из 10 · Далее — что просим решить и как выглядит договорённость о найме.</Footer>
    </div>
  );
}
