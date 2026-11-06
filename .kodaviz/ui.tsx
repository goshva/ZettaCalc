import type { CSSProperties, ReactNode } from "react";

export function Masthead({
  eyebrow,
  title,
  subtitle,
}: {
  eyebrow: string;
  title: string;
  subtitle?: ReactNode;
}) {
  return (
    <header className="masthead">
      <div className="eyebrow">{eyebrow}</div>
      <h1 className="title">{title}</h1>
      {subtitle ? <p className="subtitle">{subtitle}</p> : null}
    </header>
  );
}

export function Card({
  title,
  kicker,
  tone,
  style,
  children,
}: {
  title?: ReactNode;
  kicker?: string;
  tone?: "brand" | "teal" | "violet" | "warn" | "danger" | "flat";
  style?: CSSProperties;
  children: ReactNode;
}) {
  const cls = tone === "flat" ? "card card--flat" : tone ? `card card--${tone}` : "card";
  return (
    <section className={cls} style={style}>
      {kicker ? <div className="card-kicker">{kicker}</div> : null}
      {title ? <h2 className="card-title">{title}</h2> : null}
      <div className="card-body">{children}</div>
    </section>
  );
}

export function Stat({
  value,
  label,
  note,
  hero,
}: {
  value: string;
  label: string;
  note?: ReactNode;
  hero?: boolean;
}) {
  return (
    <div className={hero ? "card stat--hero" : "card"}>
      <div className="stat-value">{value}</div>
      <div className="stat-label">{label}</div>
      {note ? <div className="stat-note">{note}</div> : null}
    </div>
  );
}

export function Chip({
  children,
  tone,
}: {
  children: ReactNode;
  tone?: "brand" | "teal" | "warn" | "danger" | "violet" | "free" | "paid" | "flat";
}) {
  return <span className={tone && tone !== "flat" ? `chip chip--${tone}` : "chip"}>{children}</span>;
}

export function Tabs({
  items,
  value,
  onChange,
  label,
}: {
  items: { id: string; label: string }[];
  value: string;
  onChange: (id: string) => void;
  label: string;
}) {
  return (
    <div className="tabs" role="tablist" aria-label={label}>
      {items.map((it) => (
        <button
          key={it.id}
          type="button"
          role="tab"
          className="tab"
          aria-selected={value === it.id}
          onClick={() => onChange(it.id)}
        >
          {it.label}
        </button>
      ))}
    </div>
  );
}

export function Bar({
  name,
  value,
  max,
  right,
  tone,
}: {
  name: string;
  value: number;
  max: number;
  right: ReactNode;
  tone?: string;
}) {
  const pct = Math.max(2, Math.min(100, Math.round((value / max) * 100)));
  return (
    <div className="bar-row">
      <div className="bar-name">{name}</div>
      <div className="bar-track">
        <div className="bar-fill" style={{ width: `${pct}%`, background: tone ?? "var(--brand-2)" }} />
      </div>
      <div className="bar-right num">{right}</div>
    </div>
  );
}

export function Table({
  head,
  rows,
  compact,
}: {
  head: string[];
  rows: ReactNode[][];
  compact?: boolean;
}) {
  return (
    <table className={compact ? "table table--compact" : "table"}>
      <thead>
        <tr>
          {head.map((h) => (
            <th key={h}>{h}</th>
          ))}
        </tr>
      </thead>
      <tbody>
        {rows.map((r, i) => (
          <tr key={i}>
            {r.map((c, j) => (
              <td key={j}>{c}</td>
            ))}
          </tr>
        ))}
      </tbody>
    </table>
  );
}

export function Callout({
  tone,
  title,
  children,
}: {
  tone: "brand" | "teal" | "warn" | "danger" | "violet";
  title: string;
  children: ReactNode;
}) {
  return (
    <div className={`callout callout--${tone}`}>
      <div className="callout-title">{title}</div>
      <div>{children}</div>
    </div>
  );
}

export function SectionLabel({ children }: { children: ReactNode }) {
  return <div className="section-label">{children}</div>;
}

export function Footer({ children }: { children: ReactNode }) {
  return <div className="footer">{children}</div>;
}

export function Slider({
  label,
  value,
  min,
  max,
  step,
  unit,
  onChange,
}: {
  label: string;
  value: number;
  min: number;
  max: number;
  step: number;
  unit: string;
  onChange: (v: number) => void;
}) {
  return (
    <label style={{ display: "block", marginBottom: 11 }}>
      <span style={{ display: "flex", justifyContent: "space-between", fontSize: 12, fontWeight: 650, color: "var(--ink)" }}>
        <span>{label}</span>
        <span className="num" style={{ color: "var(--brand-2)" }}>
          {value.toLocaleString("ru-RU")} {unit}
        </span>
      </span>
      <input
        type="range"
        min={min}
        max={max}
        step={step}
        value={value}
        aria-label={label}
        onChange={(e) => onChange(Number(e.target.value))}
      />
    </label>
  );
}
