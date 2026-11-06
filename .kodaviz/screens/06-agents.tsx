import { useState } from "react";
import { Card, Chip, Callout, SectionLabel, Masthead, Footer, Table, Tabs } from "../ui";
import { Block, RangeBars } from "../charts";
import { agents } from "../data";

const prod = agents.production;
const mkt = agents.marketing;

const prodNeed: string[][] = [
  ["Габариты и посадочные размеры по каждому SKU", "в тестовом каталоге 3 позиции", "нужен реальный PIM"],
  ["Техкарты с нормами расхода и допусками", "1 карта из цели 10", "технологи, волна 0"],
  ["Данные с линии (фото, штрихкод, статус)", "нет интеграции", "источник, затем агент"],
  ["Прайс работ и нормы времени", "works_rub = 0", "коммерция, волна 0"],
];

export default function Agents() {
  const [tab, setTab] = useState("prod");
  const [sel, setSel] = useState(0);
  const src = tab === "prod" ? prod : mkt;
  const item = src.items[Math.min(sel, src.items.length - 1)];

  return (
    <div className="screen">
      <Masthead
        eyebrow="Экран 6 · Автоматизация: производство, склад, маркетинг"
        title="Агенты для сотрудников на производстве и в маркетинге"
        subtitle="Один и тот же PIM, те же техкарты и те же ограждения, что у клиентского калькулятора. Агент предлагает — решение остаётся за человеком: ни одна запись в план или в рассылку не уходит без подтверждения."
      />

      <Tabs
        label="Направление"
        value={tab}
        onChange={(t) => {
          setTab(t);
          setSel(0);
        }}
        items={[
          { id: "prod", label: "Производство и склад (ZETTA)" },
          { id: "mkt", label: "Маркетинг и продажи" },
        ]}
      />

      <div className="split-38">
        <Card tone={tab === "prod" ? "brand" : "teal"} title={src.title} kicker={src.kicker}>
          <div className="steps">
            {src.items.map((it, i) => (
              <button
                key={it.name}
                type="button"
                className="step"
                aria-pressed={i === sel}
                onClick={() => setSel(i)}
              >
                <span className="step-no">{String(i + 1).padStart(2, "0")}</span>
                <span>
                  <span className="step-name">{it.name}</span>
                  <span className="step-sub">{it.detail}</span>
                </span>
                <span className="step-meta">
                  <Chip tone="flat">{it.state}</Chip>
                </span>
              </button>
            ))}
          </div>
        </Card>

        <div className="stack">
          <Card title={item.name} tone="brand">
            <p className="small">{item.detail}</p>
            <div className="stack-8" style={{ marginTop: 10 }}>
              <Block label="Эффект для бизнеса" value={item.effect} color="var(--teal)" />
              <Block label="Когда реалистично" value={item.state} color="var(--brand-2)" />
            </div>
          </Card>

          {tab === "prod" ? (
            <Callout tone="brand" title="Почему это не «отдельный ИИ-проект»">
              <p className="small">
                Агент раскроя и калькулятор клиента читают <span className="strong">один PIM и одну
                библиотеку техкарт</span>. Нет второго каталога, нет расхождения «что обещали
                клиенту» и «что реально производим», нет второй команды поддержки.
              </p>
            </Callout>
          ) : (
            <Callout tone="brand" title="Контент только из PIM">
              <p className="small">
                Агент пишет описания и письма <span className="strong">по реальным SKU, ценам и
                габаритам</span>: не «кухня мечты по выгодной цене», а «Фасад Прага, белый,
                600×400 мм, 18 900 ₽». Нельзя опубликовать то, чего нет в каталоге.
              </p>
            </Callout>
          )}

          <Callout tone="danger" title="Чего агент не делает">
            <p className="small">
              {tab === "prod" ? (
                <>
                  Не меняет техкарту, не убирает обязательные операции, не снижает норму расхода
                  ради красивого плана. Любое отклонение — <span className="strong">заявка эксперту
                  на пересмотр</span>, а не тихое решение модели.
                </>
              ) : (
                <>
                  Не влияет на расчёт, цены и состав ЦПР; не создаёт и не удаляет SKU; не запускает
                  кампанию сам. Предлагает — решает человек.
                </>
              )}
            </p>
          </Callout>
        </div>
      </div>

      {tab === "prod" ? (
        <>
          <SectionLabel>Экономический эффект: где именно появляются деньги</SectionLabel>
          <div className="split-38">
            <Card title="Что снижает себестоимость">
              <RangeBars
                max={100}
                unit="% эффекта"
                rows={[
                  { name: "Отходы материала при раскрое", min: 40, max: 70, color: "var(--teal)", note: "агрегация заказов: режем лист под несколько кухонь, а не под каждую отдельно" },
                  { name: "Брак, обнаруженный на монтаже", min: 50, max: 80, color: "var(--brand-2)", note: "контроль по фото до отгрузки дешевле переделки у клиента" },
                  { name: "Заморозка денег на складе", min: 20, max: 45, color: "var(--violet)", note: "прогноз потребности из подтверждённых проектов, а не из истории прошлого года" },
                  { name: "Простой бригад из-за нехватки деталей", min: 25, max: 50, color: "var(--warn)", note: "график загрузки учитывает наличие, а не только срок доставки" },
                ]}
              />
              <p className="xs muted" style={{ marginTop: 4 }}>
                Диапазоны — экспертная оценка эффекта, а не измеренные цифры. Замеряем на пилоте.
              </p>
            </Card>
            <Card tone="violet" title="Что нужно, чтобы это заработало">
              <Table compact head={["Что нужно", "Сегодня", "От кого зависит"]} rows={prodNeed} />
              <p className="xs warn-text" style={{ marginTop: 6 }}>
                Без первых двух пунктов агент производства — красивая визуализация неполных данных.
              </p>
            </Card>
          </div>

          <SectionLabel>Как результат агента попадает в работу</SectionLabel>
          <div className="grid grid-3">
            <Card tone="flat">
              <div className="card-kicker">Формат вывода</div>
              <p className="small">
                Структурный JSON по схеме, как и ЦПР. Карта раскроя приходит в станок или в задачу
                кладовщику машиночитаемо, а не текстом в чате.
              </p>
            </Card>
            <Card tone="flat">
              <div className="card-kicker">Проверка результата</div>
              <p className="small">
                Тот же механизм <span className="strong">pass^k</span>: k прогонов на одном входе
                дают один результат. «Каждый раз предлагает по-своему» в производство не пускаем.
              </p>
            </Card>
            <Card tone="flat">
              <div className="card-kicker">Роль человека</div>
              <p className="small">
                Агент предлагает, технолог или мастер смены подтверждает. Запись в производственный
                план — только через человека, как и заказ у клиента.
              </p>
            </Card>
          </div>
        </>
      ) : (
        <>
          <SectionLabel>Чем маркетинговый агент питается и что из этого следует</SectionLabel>
          <div className="grid grid-2">
            <Card tone="flat">
              <div className="card-kicker">Входные данные</div>
              <ul className="list list--tight">
                <li>Ответы брифа: состав семьи, бюджет, приоритеты</li>
                <li>История просмотров: какие идеи задержались дольше</li>
                <li>Допущения: что клиент ещё не подтвердил</li>
                <li>Тип жилья и география: новостройка, вторичка</li>
              </ul>
            </Card>
            <Card tone="flat">
              <div className="card-kicker">Выходные действия</div>
              <ul className="list list--tight">
                <li>Рассылка с идеями своего сегмента, а не «всем одинаково»</li>
                <li>SEO-страницы под реальные запросы с настоящими артикулами</li>
                <li>Ранжирование референсов: эконом первым, апселл рядом</li>
                <li>Прогноз популярных стилей и материалов для дизайн-направления</li>
              </ul>
            </Card>
          </div>

          <SectionLabel>Ожидаемый эффект на пилоте (проверяется A/B-тестами)</SectionLabel>
          <div className="grid grid-3">
            <Card tone="flat">
              <div className="stat-value" style={{ color: "var(--teal)" }}>≥ 40%</div>
              <div className="stat-label">Сессий доходят до расчёта</div>
              <p className="stat-note">за счёт персонализации референсов — метрика №1 плана</p>
            </Card>
            <Card tone="flat">
              <div className="stat-value" style={{ color: "var(--brand-2)" }}>−30%</div>
              <div className="stat-label">Стоимость лида</div>
              <p className="stat-note">контент и SEO из PIM, без ручного копирайтинга «с нуля»</p>
            </Card>
            <Card tone="flat">
              <div className="stat-value" style={{ color: "var(--violet)" }}>+20%</div>
              <div className="stat-label">Средний чек</div>
              <p className="stat-note">уместный апселл вместо навязывания «самого дорогого»</p>
            </Card>
          </div>
          <p className="xs muted" style={{ marginTop: 6 }}>
            Цифры — экспертная оценка. A/B-тесты дают объективные данные: решение о доработке пути
            по цифрам, а не по «мне кажется».
          </p>
        </>
      )}

      <Footer>Экран 6 из 10 · Далее — мониторинг ИТ-инфраструктуры предприятия.</Footer>
    </div>
  );
}
