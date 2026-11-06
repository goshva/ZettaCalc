import type { ReactNode } from "react";

/* Небольшие SVG-графики без внешних библиотек. */

export function LineChart({
  w = 520,
  h = 210,
  xLabels,
  series,
  yMax,
  yUnit = "%",
}: {
  w?: number;
  h?: number;
  xLabels: string[];
  series: { name: string; color: string; values: number[]; dashed?: boolean }[];
  yMax: number;
  yUnit?: string;
}) {
  const padL = 38;
  const padR = 12;
  const padT = 14;
  const padB = 26;
  const iw = w - padL - padR;
  const ih = h - padT - padB;
  const n = xLabels.length;
  const x = (i: number) => padL + (n === 1 ? iw / 2 : (iw * i) / (n - 1));
  const y = (v: number) => padT + ih - (Math.max(0, Math.min(yMax, v)) / yMax) * ih;
  const ticks = [0, 0.25, 0.5, 0.75, 1].map((t) => Math.round(yMax * t));

  return (
    <svg viewBox={`0 0 ${w} ${h}`} width="100%" height={h} role="img" aria-label="График динамики">
      {ticks.map((t) => (
        <g key={t}>
          <line className="grid-line" x1={padL} x2={w - padR} y1={y(t)} y2={y(t)} />
          <text className="axis" x={padL - 6} y={y(t) + 3} textAnchor="end">
            {t}
            {yUnit}
          </text>
        </g>
      ))}
      {xLabels.map((l, i) => (
        <text key={l + i} className="axis" x={x(i)} y={h - 8} textAnchor="middle">
          {l}
        </text>
      ))}
      {series.map((s) => (
        <g key={s.name}>
          <polyline
            fill="none"
            stroke={s.color}
            strokeWidth={2.2}
            strokeDasharray={s.dashed ? "5 4" : undefined}
            points={s.values.map((v, i) => `${x(i)},${y(v)}`).join(" ")}
          />
          {s.values.map((v, i) => (
            <circle key={i} cx={x(i)} cy={y(v)} r={3.2} fill={s.color} />
          ))}
        </g>
      ))}
    </svg>
  );
}

export function Donut({
  segments,
  size = 168,
  center,
  caption,
}: {
  segments: { name: string; value: number; color: string }[];
  size?: number;
  center: string;
  caption: string;
}) {
  const total = segments.reduce((s, x) => s + x.value, 0) || 1;
  const r = size / 2 - 12;
  const c = size / 2;
  const circ = 2 * Math.PI * r;
  let acc = 0;

  return (
    <div style={{ display: "flex", gap: 14, alignItems: "center", flexWrap: "wrap" }}>
      <svg width={size} height={size} viewBox={`0 0 ${size} ${size}`} role="img" aria-label={caption}>
        <circle cx={c} cy={c} r={r} fill="none" stroke="var(--line)" strokeWidth={16} />
        {segments.map((s) => {
          const frac = s.value / total;
          const dash = frac * circ;
          const el = (
            <circle
              key={s.name}
              cx={c}
              cy={c}
              r={r}
              fill="none"
              stroke={s.color}
              strokeWidth={16}
              strokeDasharray={`${dash} ${circ - dash}`}
              strokeDashoffset={-acc * circ}
              transform={`rotate(-90 ${c} ${c})`}
              strokeLinecap="butt"
            />
          );
          acc += frac;
          return el;
        })}
        <text
          x={c}
          y={c + 2}
          textAnchor="middle"
          style={{ fontSize: 22, fontWeight: 800, fill: "var(--ink)", fontFamily: "var(--font)" }}
        >
          {center}
        </text>
        <text x={c} y={c + 20} textAnchor="middle" className="axis">
          {caption}
        </text>
      </svg>
      <div className="legend" style={{ flexDirection: "column", gap: 5 }}>
        {segments.map((s) => (
          <div key={s.name} className="legend-item">
            <span className="legend-swatch" style={{ background: s.color }} />
            <span>
              {s.name} · <b className="num">{s.value}</b>
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}

export function RangeBars({
  rows,
  max,
  unit,
}: {
  rows: { name: string; min: number; max: number; color: string; note?: string }[];
  max: number;
  unit: string;
}) {
  return (
    <div>
      {rows.map((r) => {
        const left = (r.min / max) * 100;
        const width = Math.max(4, ((r.max - r.min) / max) * 100);
        return (
          <div key={r.name} style={{ marginBottom: 10 }}>
            <div style={{ display: "flex", justifyContent: "space-between", fontSize: 12, fontWeight: 650 }}>
              <span>{r.name}</span>
              <span className="num muted">
                {r.min}–{r.max} {unit}
              </span>
            </div>
            <div style={{ position: "relative", height: 22, background: "var(--surface-2)", border: "1px solid var(--line)", borderRadius: 6, marginTop: 4 }}>
              <div
                style={{
                  position: "absolute",
                  left: `${left}%`,
                  width: `${width}%`,
                  top: 3,
                  bottom: 3,
                  borderRadius: 4,
                  background: r.color,
                }}
              />
            </div>
            {r.note ? <div className="xs muted" style={{ marginTop: 3 }}>{r.note}</div> : null}
          </div>
        );
      })}
    </div>
  );
}

export function Sparkbars({
  values,
  color = "var(--brand-2)",
  height = 46,
}: {
  values: number[];
  color?: string;
  height?: number;
}) {
  const max = Math.max(...values, 1);
  return (
    <div style={{ display: "flex", alignItems: "flex-end", gap: 3, height }} aria-hidden>
      {values.map((v, i) => (
        <div
          key={i}
          style={{
            flex: 1,
            height: `${(v / max) * 100}%`,
            background: color,
            opacity: 0.45 + 0.55 * (v / max),
            borderRadius: 3,
            minWidth: 4,
          }}
        />
      ))}
    </div>
  );
}

export function Block({
  label,
  value,
  color,
}: {
  label: string;
  value: ReactNode;
  color?: string;
}) {
  return (
    <div
      style={{
        border: "1px solid var(--line)",
        borderLeft: `4px solid ${color ?? "var(--brand-2)"}`,
        borderRadius: 10,
        padding: "9px 12px",
        background: "var(--surface)",
      }}
    >
      <div className="xs muted">{label}</div>
      <div style={{ fontSize: 15, fontWeight: 750, color: "var(--ink)" }}>{value}</div>
    </div>
  );
}
