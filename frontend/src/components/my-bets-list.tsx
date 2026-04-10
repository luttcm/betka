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

      <div className="grid gap-3">
        {filtered.map((bet) => (
          <article key={bet.id} className="panel">
            <div className="flex flex-wrap items-center justify-between gap-2">
              <h3 className="text-lg font-semibold text-slate-900">Ставка {bet.id}</h3>
              <span className="status-pill border-blue-200 bg-blue-50 text-blue-700">{bet.status}</span>
            </div>

            <dl className="mt-3 grid gap-1 text-sm text-slate-700">
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
                <dt className="inline font-medium">Потенциальная выплата:</dt>{" "}
                <dd className="inline">{formatAmount(bet.potential_payout)} TOK</dd>
              </div>
              <div>
                <dt className="inline font-medium">Размещена:</dt> <dd className="inline">{formatDate(bet.placed_at)}</dd>
              </div>
            </dl>
          </article>
        ))}
      </div>
    </section>
  );
}

