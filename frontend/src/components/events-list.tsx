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
    return <p className="rounded-lg border border-slate-200 bg-white p-4">Загрузка событий...</p>;
  }

  if (isError) {
    const message = error instanceof ApiError ? error.message : "Не удалось загрузить события";
    return (
      <p className="rounded-lg border border-red-200 bg-red-50 p-4 text-red-700">
        Ошибка загрузки: {message}
      </p>
    );
  }

  if (!data || data.length === 0) {
    return (
      <div className="rounded-lg border border-slate-200 bg-white p-6">
        <p className="text-slate-700">Пока нет опубликованных событий.</p>
      </div>
    );
  }

  return (
    <div className="grid gap-4">
      {data.map((event) => (
        <article key={event.id} className="rounded-xl border border-slate-200 bg-white p-5 shadow-sm">
          <div className="mb-2 flex items-center justify-between gap-2">
            <h2 className="text-lg font-semibold">{event.title}</h2>
            <span className="rounded bg-slate-100 px-2 py-1 text-xs text-slate-600">{event.category || "Без категории"}</span>
          </div>

          <p className="mb-3 text-sm text-slate-700">{event.description}</p>

          <div className="mb-4 text-xs text-slate-500">Решение события: {formatDate(event.resolve_at)}</div>

          <Link href={`/events/${event.id}`} className="inline-flex rounded-md border border-slate-300 px-3 py-2 text-sm font-medium">
            Открыть карточку
          </Link>
        </article>
      ))}
    </div>
  );
}

