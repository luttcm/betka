"use client";

import Link from "next/link";
import { useQuery } from "@tanstack/react-query";
import { useMemo } from "react";

import { ApiError, getEvents } from "@/lib/api";
import { EmptyState, ErrorState, LoadingState } from "@/components/ui-states";

type EventList = Awaited<ReturnType<typeof getEvents>>;
type EventGroup = { category: string; items: EventList };

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

export function EventsList() {
  const { data, isLoading, isError, error } = useQuery({
    queryKey: ["events"],
    queryFn: getEvents,
  });

  const grouped = useMemo(() => {
    if (!data) {
      return [] as EventGroup[];
    }

    const map = new Map<string, EventList>();
    for (const event of data) {
      const key = event.category?.trim() || "Без категории";
      const next = map.get(key) ?? [];
      next.push(event);
      map.set(key, next);
    }

    return Array.from(map.entries())
      .map(([category, items]) => ({ category, items }))
      .sort((a, b) => b.items.length - a.items.length);
  }, [data]);

  if (isLoading) {
    return <LoadingState message="Загрузка событий..." />;
  }

  if (isError) {
    const message = error instanceof ApiError ? error.message : "Не удалось загрузить события";
    return <ErrorState message={message} />;
  }

  if (!data || data.length === 0) {
    return <EmptyState message="Пока нет опубликованных событий." />;
  }

  return (
    <div className="grid gap-4 lg:grid-cols-3">
      {grouped.map((column) => (
        <section key={column.category} className="panel space-y-3">
          <header className="flex items-center justify-between gap-2 border-b border-[color:var(--muted-border)] pb-2">
            <h3 className="text-sm font-semibold text-slate-700">{column.category}</h3>
            <span className="status-pill border-[color:var(--muted-border)] bg-[#f5f7fa] text-slate-600">{column.items?.length ?? 0}</span>
          </header>

          <div className="space-y-3">
            {(column.items ?? []).map((event: EventList[number]) => (
              <article key={event.id} className="rounded-lg border border-[color:var(--muted-border)] bg-[#f8fafc] p-3">
                <h2 className="text-base font-semibold text-slate-800">{event.title}</h2>
                <p className="mt-1 text-sm text-slate-600">{event.description}</p>

                <div className="mt-2 flex flex-wrap items-center gap-2 text-xs text-slate-500">
                  <span>Решение: {formatDate(event.resolve_at)}</span>
                  <span className="status-pill border-blue-200 bg-blue-50 text-blue-700">{event.status}</span>
                </div>

                <Link href={`/events/${event.id}`} className="btn-secondary mt-3">
                  Открыть карточку
                </Link>
              </article>
            ))}
          </div>
        </section>
      ))}
    </div>
  );
}
