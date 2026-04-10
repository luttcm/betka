"use client";

import { useMemo, useState } from "react";

import { BetItem, BetStatus } from "@/lib/types";

interface MyBetsListProps {
  items: BetItem[];
}

const STATUS_OPTIONS: Array<{ value: "all" | BetStatus; label: string }> = [
  { value: "all", label: "Все" },
  { value: "open", label: "open" },
  { value: "won", label: "won" },
  { value: "lost", label: "lost" },
  { value: "refunded", label: "refunded" },
];

function formatAmount(value: number): string {
  return new Intl.NumberFormat("ru-RU", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(value);
}

function formatDate(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return new Intl.DateTimeFormat("ru-RU", {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date);
}

export function MyBetsList({ items }: MyBetsListProps) {
  const [statusFilter, setStatusFilter] = useState<"all" | BetStatus>("all");

  const filtered = useMemo(() => {
    const sorted = [...items].sort((a, b) => new Date(b.placed_at).getTime() - new Date(a.placed_at).getTime());
    if (statusFilter === "all") {
      return sorted;
    }

    return sorted.filter((item) => item.status === statusFilter);
  }, [items, statusFilter]);

  const columns: Array<{ key: "open" | "won" | "lost" | "refunded"; title: string }> = [
    { key: "open", title: "Открытые" },
    { key: "won", title: "Выиграли" },
    { key: "lost", title: "Проиграли" },
    { key: "refunded", title: "Возвраты" },
  ];

  const grouped = columns.map((column) => ({
    ...column,
    items: filtered.filter((bet) => bet.status === column.key),
  }));

  const gridColumns = statusFilter === "all" ? grouped : grouped.filter((column) => column.key === statusFilter);

  return (
    <section className="space-y-4">
      <div className="panel">
        <label htmlFor="bet-status-filter" className="field-label">
          Фильтр по статусу
        </label>
        <select
          id="bet-status-filter"
          className="text-input mt-2 max-w-xs"
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value as "all" | BetStatus)}
        >
          {STATUS_OPTIONS.map((option) => (
            <option key={option.value} value={option.value}>
              {option.label}
            </option>
          ))}
        </select>
      </div>

      <div className="grid gap-3 lg:grid-cols-4">
        {gridColumns.map((column) => (
          <section key={column.key} className="panel space-y-3">
            <header className="flex items-center justify-between gap-2">
              <h3 className="text-sm font-semibold text-slate-700">{column.title}</h3>
              <span className="status-pill border-[color:var(--muted-border)] bg-[#f5f7fa] text-slate-600">{column.items.length}</span>
            </header>

            <div className="space-y-3">
              {column.items.length === 0 && <p className="text-xs text-slate-500">Пусто</p>}

              {column.items.map((bet) => (
                <article key={bet.id} className="rounded-lg border border-[color:var(--muted-border)] bg-[#f8fafc] p-3">
                  <div className="flex flex-wrap items-center justify-between gap-2">
                    <h4 className="text-sm font-semibold text-slate-800">Ставка {bet.id}</h4>
                    <span className="status-pill border-[color:var(--muted-border)] bg-white text-slate-600">{bet.status}</span>
                  </div>

                  <dl className="mt-2 grid gap-1 text-xs text-slate-600">
                    <div>
                      <dt className="inline font-medium">Событие:</dt> <dd className="inline">{bet.event_id}</dd>
                    </div>
                    <div>
                      <dt className="inline font-medium">Исход:</dt> <dd className="inline">{bet.outcome_code}</dd>
                    </div>
                    <div>
                      <dt className="inline font-medium">Сумма:</dt> <dd className="inline">{formatAmount(bet.stake)} TOK</dd>
                    </div>
                    <div>
                      <dt className="inline font-medium">Выплата:</dt> <dd className="inline">{formatAmount(bet.potential_payout)} TOK</dd>
                    </div>
                    <div>
                      <dt className="inline font-medium">Дата:</dt> <dd className="inline">{formatDate(bet.placed_at)}</dd>
                    </div>
                  </dl>
                </article>
              ))}
            </div>
          </section>
        ))}
      </div>
    </section>
  );
}
