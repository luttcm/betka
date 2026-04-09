"use client";

import Link from "next/link";
import { useQuery } from "@tanstack/react-query";

import { ApiError, getEvents } from "@/lib/api";

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

  if (isLoading) {
    return <p className="panel">Загрузка событий...</p>;
  }

  if (isError) {
    const message = error instanceof ApiError ? error.message : "Не удалось загрузить события";
    return (
      <p className="rounded-2xl border border-red-200 bg-red-50 p-4 text-red-700">
        Ошибка загрузки: {message}
      </p>
    );
  }

  if (!data || data.length === 0) {
    return (
      <div className="panel">
        <p className="text-slate-700">Пока нет опубликованных событий.</p>
      </div>
    );
  }

  return (
    <div className="grid gap-4">
      {data.map((event) => (
        <article key={event.id} className="panel">
          <div className="mb-2 flex items-center justify-between gap-2">
            <h2 className="text-xl font-semibold">{event.title}</h2>
            <span className="rounded-full border border-[color:var(--muted-border)] bg-[var(--cool-surface)] px-3 py-1 text-xs text-slate-700">
              {event.category || "Без категории"}
            </span>
          </div>

          <p className="mb-3 text-sm text-slate-700">{event.description}</p>

          <div className="mb-4 flex flex-wrap items-center gap-2 text-xs text-slate-500">
            <span>Решение события: {formatDate(event.resolve_at)}</span>
            <span className="status-pill border-blue-200 bg-blue-50 text-blue-700">{event.status}</span>
          </div>

          <Link href={`/events/${event.id}`} className="btn-secondary">
            Открыть карточку
          </Link>
        </article>
      ))}
    </div>
  );
}
