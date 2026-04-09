"use client";

import Link from "next/link";
import { useQuery } from "@tanstack/react-query";

import { ApiError, getEventById } from "@/lib/api";
import { useAuth } from "@/lib/auth-context";

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
  const { canModerate } = useAuth();

  const { data, isLoading, isError, error } = useQuery({
    queryKey: ["event", eventId],
    queryFn: () => getEventById(eventId),
  });

  if (isLoading) {
    return <p className="panel">Загрузка события...</p>;
  }

  if (isError) {
    const message = error instanceof ApiError ? error.message : "Не удалось получить событие";
    return (
      <p className="rounded-2xl border border-red-200 bg-red-50 p-4 text-red-700">
        Ошибка: {message}
      </p>
    );
  }

  if (!data) {
    return <p className="panel">Событие не найдено.</p>;
  }

  return (
    <article className="panel space-y-4">
      <header className="space-y-2">
        <h2 className="text-3xl font-semibold leading-tight">{data.title}</h2>
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
          <dd>
            <span className="status-pill border-blue-200 bg-blue-50 text-blue-700">{data.status}</span>
          </dd>
        </div>
        <div>
          <dt className="font-medium">Дата решения</dt>
          <dd>{formatDate(data.resolve_at)}</dd>
        </div>
      </dl>

      {canModerate && (
        <section className="rounded-2xl border border-[color:var(--muted-border)] bg-[var(--cool-surface)] p-4 text-sm text-slate-700">
          Для действий модерации используйте вкладку <Link href="/moderation">«Модерация»</Link>.
        </section>
      )}
    </article>
  );
}
