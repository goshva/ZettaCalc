import { useState } from "react";
import { Card, Chip, Callout, SectionLabel, Masthead, Footer, Slider, Table } from "../ui";
import { RangeBars, Block } from "../charts";
import { boundary, economics } from "../data";

export default function Money() {
  const [clients, setClients] = useState(200);
  const [calcRate, setCalcRate] = useState(40);
  const [cprConv, setCprConv] = useState(8);
  const [cprPrice, setCprPrice] = useState(15000);
  const [avgOrder, setAvgOrder] = useState(350000);
  const [goodsMargin, setGoodsMargin] = useState(22);
  const [auctShare, setAuctShare] = useState(40);
  const [auctComm, setAuctComm] = useState(5);
  const [manualShare, setManualShare] = useState(45);

  const calc = Math.round((clients * calcRate) / 100);
  const cpr = Math.round((calc * cprConv) / 100);
  const goodsIncome = Math.round((cpr * avgOrder * goodsMargin) / 100);
  const cprIncome = cpr * cprPrice;
  const auctIncome = Math.round((cpr * auctShare) / 100) * ((avgOrder * auctComm) / 100);
  const total = goodsIncome + cprIncome + auctIncome;
  const manualHours = Math.round((cpr * manualShare) / 100) * 6;
  const manualCost = manualHours * 1200;
  const net = total - manualCost;

  const fmt = (n: number) => Math.round(n).toLocaleString("ru-RU");

  return (
    <div className="screen">
      <Masthead
        eyebrow="Экран 3 · Экономика"
        title="Что именно продаём и сколько это может приносить"
        subtitle="Слева — что входит в каждый продукт и чем он отличается от соседнего. Справа — перетащите ползунки: модель дохода мгновенно пересчитается. Все базовые цифры — гипотезы, их замеряют на пилоте."
      />

      <SectionLabel>Три продукта, одна система</SectionLabel>
      <div className="grid grid-3">
        {boundary.map((b, i) => (
          <Card
            key={b.name}
            tone={i === 0 ? "teal" : i === 1 ? "brand" : "violet"}
            title={b.name}
          >
            <div style={{ display: "flex", gap: 6, marginBottom: 8, flexWrap: "wrap" }}>
              <Chip tone={i === 0 ? "free" : "paid"}>{b.price}</Chip>
              <Chip>{b.answers}</Chip>
            </div>
            <ul className="list list--tight">
              {b.includes.map((x) => (
                <li key={x}>{x}</li>
              ))}
            </ul>
            <p className="xs muted" style={{ marginTop: 8 }}>
              Участие эксперта: <span className="strong">{b.expert}</span>
            </p>
          </Card>
        ))}
      </div>

      <Callout tone="warn" title="Почему расчёт и ЦПР надо разделить ценой">
        <p>
          В прототипах встречается формулировка «при заказе ремонта паспорт входит в стоимость». Пока
          ЦПР бесплатный «внутри» ремонта, мы не узнаем, готовы ли за него платить, и не сможем
          посчитать экономику платформы. Отдельная цена и отдельный SLA — это ещё и инструмент
          измерения спроса.
        </p>
      </Callout>

      <SectionLabel>Модель дохода: четыре потока</SectionLabel>
      <Card>
        <Table
          compact
          head={["Поток", "Формула", "Что измеряем на пилоте", "Проверяем"]}
          rows={economics.map((e) => [
            <span className="strong">{e.flow}</span>,
            <code style={{ fontSize: 11.5 }}>{e.formula}</code>,
            e.measure,
            <Chip tone="brand">{e.order}</Chip>,
          ])}
        />
      </Card>

      <SectionLabel>Калькулятор гипотез (месяц пилота)</SectionLabel>
      <div className="split-38">
        <Card title="Параметры">
          <div className="grid grid-2">
            <div>
              <Slider label="Сессий на сайте" value={clients} min={50} max={2000} step={50} unit="" onChange={setClients} />
              <Slider label="Дошли до расчёта" value={calcRate} min={5} max={80} step={1} unit="%" onChange={setCalcRate} />
              <Slider label="Купили ЦПР" value={cprConv} min={1} max={30} step={1} unit="%" onChange={setCprConv} />
              <Slider label="Цена базового ЦПР" value={cprPrice} min={3000} max={60000} step={1000} unit="₽" onChange={setCprPrice} />
            </div>
            <div>
              <Slider label="Средняя корзина по проекту" value={avgOrder} min={50000} max={1500000} step={10000} unit="₽" onChange={setAvgOrder} />
              <Slider label="Маржа на товарах" value={goodsMargin} min={5} max={45} step={1} unit="%" onChange={setGoodsMargin} />
              <Slider label="ЦПР ушли в аукцион" value={auctShare} min={0} max={100} step={5} unit="%" onChange={setAuctShare} />
              <Slider label="Комиссия подрядчика" value={auctComm} min={1} max={20} step={1} unit="%" onChange={setAuctComm} />
            </div>
          </div>
          <div style={{ borderTop: "1px dashed var(--line-2)", paddingTop: 10, marginTop: 4 }}>
            <Slider
              label="Доля ЦПР, где требовался человек (pending_review)"
              value={manualShare}
              min={5}
              max={80}
              step={5}
              unit="%"
              onChange={setManualShare}
            />
            <p className="xs muted">
              Принято: 6 человеко-часов на доработку ЦПР экспертом, 1 200 ₽/час. Это единственная
              заметная себестоимость в модели — и она прямо зависит от надёжности автоматики.
            </p>
          </div>
        </Card>

        <div className="stack">
          <div className="grid grid-2" style={{ gap: 8 }}>
            <Block label="Расчётов в месяц" value={fmt(calc)} color="var(--teal)" />
            <Block label="Продано ЦПР" value={fmt(cpr)} color="var(--brand-2)" />
            <Block label="Товары (маржа)" value={`${fmt(goodsIncome)} ₽`} color="var(--teal)" />
            <Block label="ЦПР" value={`${fmt(cprIncome)} ₽`} color="var(--brand-2)" />
            <Block label="Комиссия подрядчиков" value={`${fmt(auctIncome)} ₽`} color="var(--violet)" />
            <Block label="Труд на pending_review" value={`−${fmt(manualCost)} ₽`} color="var(--danger)" />
          </div>
          <Card tone="brand" title="Валовой доход за месяц">
            <div className="stat-value" style={{ fontSize: 32 }}>
              {fmt(net)} ₽
            </div>
            <p className="small muted" style={{ marginTop: 6 }}>
              {fmt(total)} ₽ дохода минус {fmt(manualCost)} ₽ ручного труда. Не включает зарплаты
              команды, инфраструктуру и маркетинг — это валовая маржа продукта, а не прибыль
              компании.
            </p>
          </Card>
          <Callout tone="danger" title="Что ломает модель">
            <p>
              Подвигайте ползунок «доля ЦПР с человеком». При 45% себестоимость растёт быстрее, чем
              выручка: каждый процент ручной доработки — это минус маржа. Поэтому{" "}
              <span className="strong">цена ЦПР должна считаться от измеренной доли, а не от
              целевой</span>.
            </p>
          </Callout>
        </div>
      </div>

      <SectionLabel>Диапазоны неопределённости, которые честно закладываем</SectionLabel>
      <Card tone="flat">
        <RangeBars
          unit=""
          max={26}
          rows={[
            { name: "Срок до первого расчёта на сайте", min: 3, max: 8, color: "var(--teal)", note: "минуты. Цель — не более 5; выше 8 клиент уходит." },
            { name: "Доля ЦПР без участия человека", min: 50, max: 60, color: "var(--brand-2)", note: "проценты в v1. Стратегические 90% — цель, не обещание." },
            { name: "Бюджет волны 0", min: 1, max: 2, color: "var(--warn)", note: "млн ₽. Данные и правила: S2T, PIM, 10 техкарт." },
            { name: "Бюджет волны 1", min: 3, max: 5, color: "var(--violet)", note: "млн ₽. Расчёт и базовый ЦПР на пилоте Zetta Home." },
            { name: "Бюджет волны 2", min: 4, max: 6, color: "#334155", note: "млн ₽. Категории, Text2SQL, зрелые ИИ-сценарии." },
          ]}
        />
      </Card>

      <Footer>
        Экран 3 из 10 · Цифры — гипотезы из ответа CPTO и docs/economics.md. Тарифы намеренно не
        заданы: нет исходных данных компании.
      </Footer>
    </div>
  );
}
