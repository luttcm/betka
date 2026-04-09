"use client";

import { useQuery } from "@tanstack/react-query";

import { ApiError, getEventById } from "@/lib/api";

interface EventDetailsProps {
  eventId: string;
}

function formatDate(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return new Intl.DateTimeFormat("ru-RU", {
    dateStyle: "full",
    timeStyle: "short",
  }).format(date);
}

export function EventDetails({ eventId }: EventDetailsProps) {
  const { data, isLoading, isError, error } = useQuery({
    queryKey: ["event", eventId],
    queryFn: () => getEventById(eventId),
  });

  if (isLoading) {
    return <p className="rounded-lg border border-slate-200 bg-white p-4">Загрузка события...</p>;
  }

  if (isError) {
    const message = error instanceof ApiError ? error.message : "Не удалось получить событие";
    return (
      <p className="rounded-lg border border-red-200 bg-red-50 p-4 text-red-700">
        Ошибка: {message}
      </p>
    );
  }

  if (!data) {
    return <p className="rounded-lg border border-slate-200 bg-white p-4">Событие не найдено.</p>;
  }

  return (
    <article className="space-y-4 rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
      <header className="space-y-2">
        <h2 className="text-2xl font-semibold">{data.title}</h2>
        <p className="text-sm text-slate-600">Категория: {data.category || "Без категории"}</p>
      </header>

      <p className="text-slate-800">{data.description}</p>

      <dl className="grid gap-2 text-sm text-slate-600">
        <div>
          <dt className="font-medium">ID события</dt>
          <dd>{data.id}</dd>
        </div>
        <div>
          <dt className="font-medium">Статус</dt>
          <dd>{data.status}</dd>
        </div>
        <div>
          <dt className="font-medium">Дата решения</dt>
          <dd>{formatDate(data.resolve_at)}</dd>
        </div>
      </dl>
    </article>
  );
}

