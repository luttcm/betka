"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";

import {
  ApiError,
  approveModerationEvent,
  getModerationEvents,
  rejectModerationEvent,
} from "@/lib/api";
import { useAuth } from "@/lib/auth-context";

export function ModerationEventsPanel() {
  const queryClient = useQueryClient();
  const { token, canModerate } = useAuth();
  const [rejectReasons, setRejectReasons] = useState<Record<string, string>>({});

  const { data, isLoading, isError, error } = useQuery({
    queryKey: ["moderation-events"],
    queryFn: () => getModerationEvents(token ?? ""),
    enabled: canModerate && Boolean(token),
  });

  const approveMutation = useMutation({
    mutationFn: async (eventId: string) => approveModerationEvent(eventId, token ?? ""),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["moderation-events"] });
    },
  });

  const rejectMutation = useMutation({
    mutationFn: async ({ eventId, reason }: { eventId: string; reason: string }) =>
      rejectModerationEvent(eventId, reason, token ?? ""),
    onSuccess: async (_, variables) => {
      setRejectReasons((prev) => ({ ...prev, [variables.eventId]: "" }));
      await queryClient.invalidateQueries({ queryKey: ["moderation-events"] });
    },
  });

  if (!canModerate) {
    return (
      <div className="rounded-2xl border border-amber-200 bg-amber-50 p-4 text-sm text-amber-900">
        Доступ к вкладке модерации есть только у moderator/admin.
      </div>
    );
  }

  if (isLoading) {
    return <p className="panel">Загрузка очереди модерации...</p>;
  }

  if (isError) {
    const message = error instanceof ApiError ? error.message : "Не удалось загрузить очередь модерации";
    return <p className="rounded-2xl border border-red-200 bg-red-50 p-4 text-red-700">Ошибка: {message}</p>;
  }

  if (!data || data.length === 0) {
    return (
      <div className="panel">
        <p className="text-slate-700">Событий на модерации сейчас нет.</p>
      </div>
    );
  }

  return (
    <div className="grid gap-4">
      {data.map((queueItem) => {
        const event = queueItem.event;
        const rejectReason = rejectReasons[event.id] ?? "";

        return (
          <article key={event.id} className="panel space-y-4">
            <header className="space-y-2">
              <div className="flex flex-wrap items-center gap-2">
                <h3 className="text-2xl font-semibold leading-tight">{event.title}</h3>
                <span className="status-pill border-amber-200 bg-amber-50 text-amber-700">{event.status}</span>
              </div>
              <p className="text-sm text-slate-600">Категория: {event.category || "Без категории"}</p>
            </header>

            <p className="text-sm text-slate-700">{event.description}</p>

            <div className="grid gap-2">
              <label htmlFor={`reject-${event.id}`} className="field-label">
                Причина отмены (если отклоняете)
              </label>
              <textarea
                id={`reject-${event.id}`}
                rows={3}
                className="text-input"
                placeholder="Например: формулировка неоднозначна"
                value={rejectReason}
                onChange={(e) => setRejectReasons((prev) => ({ ...prev, [event.id]: e.target.value }))}
              />
            </div>

            <div className="flex flex-wrap gap-2">
              <button
                type="button"
                className="btn-primary"
                disabled={approveMutation.isPending || rejectMutation.isPending}
                onClick={() => approveMutation.mutate(event.id)}
              >
                {approveMutation.isPending ? "Пропускаем..." : "Пропустить дальше"}
              </button>
              <button
                type="button"
                className="btn-danger"
                disabled={rejectMutation.isPending || rejectReason.trim().length === 0}
                onClick={() => rejectMutation.mutate({ eventId: event.id, reason: rejectReason.trim() })}
              >
                {rejectMutation.isPending ? "Отменяем..." : "Отменить"}
              </button>
            </div>
          </article>
        );
      })}

      {approveMutation.isError && (
        <p className="rounded-2xl border border-red-200 bg-red-50 p-4 text-sm text-red-700">
          {approveMutation.error instanceof ApiError ? approveMutation.error.message : "Не удалось одобрить событие"}
        </p>
      )}
      {rejectMutation.isError && (
        <p className="rounded-2xl border border-red-200 bg-red-50 p-4 text-sm text-red-700">
          {rejectMutation.error instanceof ApiError ? rejectMutation.error.message : "Не удалось отклонить событие"}
        </p>
      )}
    </div>
  );
}
