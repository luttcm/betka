"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";

import {
  ApiError,
  approveModerationEvent,
  getAdminSettlementRequests,
  getModerationEvents,
  rejectModerationEvent,
  settleAdminEvent,
} from "@/lib/api";
import { EmptyState, ErrorState, LoadingState } from "@/components/ui-states";
import { useAuth } from "@/lib/auth-context";

export function ModerationEventsPanel() {
  const queryClient = useQueryClient();
  const { token, canModerate, isAdmin } = useAuth();
  const [rejectReasons, setRejectReasons] = useState<Record<string, string>>({});
  const [winnerByEvent, setWinnerByEvent] = useState<Record<string, "yes" | "no">>({});

  const { data, isLoading, isError, error } = useQuery({
    queryKey: ["moderation-events"],
    queryFn: () => getModerationEvents(token ?? ""),
    enabled: canModerate && Boolean(token),
  });

  const {
    data: settlementRequests,
    isLoading: isSettlementLoading,
    isError: isSettlementError,
    error: settlementError,
  } = useQuery({
    queryKey: ["moderation-settlement-requests"],
    queryFn: () => getAdminSettlementRequests(token ?? ""),
    enabled: isAdmin && Boolean(token),
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

  const settleMutation = useMutation({
    mutationFn: async ({ eventId, winner }: { eventId: string; winner: "yes" | "no" }) =>
      settleAdminEvent(eventId, winner, token ?? ""),
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["moderation-settlement-requests"] }),
        queryClient.invalidateQueries({ queryKey: ["moderation-events"] }),
        queryClient.invalidateQueries({ queryKey: ["event"] }),
        queryClient.invalidateQueries({ queryKey: ["events"] }),
      ]);
    },
  });

  if (!canModerate) {
    return (
      <div className="border border-amber-200 bg-amber-50 p-4 text-sm text-amber-700">
        Доступ к вкладке модерации есть только у moderator/admin.
      </div>
    );
  }

  if (isLoading) {
    return <LoadingState message="Загрузка очереди модерации..." />;
  }

  if (isError) {
    const message = error instanceof ApiError ? error.message : "Не удалось загрузить очередь модерации";
    return <ErrorState message={message} />;
  }

  if (isAdmin && isSettlementLoading) {
    return <LoadingState message="Загрузка очереди модерации и запросов на завершение..." />;
  }

  if (isAdmin && isSettlementError) {
    const message = settlementError instanceof ApiError ? settlementError.message : "Не удалось загрузить запросы на завершение";
    return <ErrorState message={message} />;
  }

  if ((!data || data.length === 0) && (!isAdmin || !settlementRequests || settlementRequests.length === 0)) {
    return <EmptyState message="Событий на модерации сейчас нет." />;
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
                <h3 className="text-2xl font-semibold leading-tight text-slate-800">{event.title}</h3>
                <span className="status-pill border-amber-200 bg-amber-50 text-amber-700">{event.status}</span>
              </div>
              <p className="text-sm text-slate-600">Категория: {event.category || "Без категории"}</p>
            </header>

            <p className="text-sm text-slate-600">{event.description}</p>

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
        <p className="border border-red-200 bg-red-50 p-4 text-sm text-red-700">
          {approveMutation.error instanceof ApiError ? approveMutation.error.message : "Не удалось одобрить событие"}
        </p>
      )}
      {rejectMutation.isError && (
        <p className="border border-red-200 bg-red-50 p-4 text-sm text-red-700">
          {rejectMutation.error instanceof ApiError ? rejectMutation.error.message : "Не удалось отклонить событие"}
        </p>
      )}

      {isAdmin && settlementRequests && settlementRequests.length > 0 && (
        <section className="panel space-y-4">
          <header className="space-y-2">
            <h3 className="text-2xl font-semibold leading-tight text-slate-800">Запросы на завершение событий</h3>
            <p className="text-sm text-slate-600">Создатели отправили доказательства, выберите итоговый исход и завершите событие.</p>
          </header>

          <div className="grid gap-4">
            {settlementRequests.map((event) => {
              const winner = winnerByEvent[event.id] ?? "yes";
              return (
                <article key={event.id} className="border border-[color:var(--muted-border)] bg-[#f8fafc] p-4 space-y-3">
                  <div className="flex flex-wrap items-center gap-2">
                    <h4 className="text-lg font-semibold text-slate-800">{event.title}</h4>
                    <span className="status-pill border-violet-200 bg-violet-50 text-violet-700">{event.status}</span>
                  </div>

                  <p className="text-sm text-slate-600">{event.description}</p>

                  {event.settlement_evidence_url && (
                    <p className="text-sm text-slate-600">
                      Доказательство (ссылка):{" "}
                      <a href={event.settlement_evidence_url} target="_blank" rel="noreferrer" className="text-[var(--brand-blue)] underline">
                        {event.settlement_evidence_url}
                      </a>
                    </p>
                  )}

                  {event.settlement_evidence_file_name && (
                    <p className="text-sm text-slate-600">Доказательство (файл): {event.settlement_evidence_file_name}</p>
                  )}

                  <div className="flex flex-wrap items-center gap-2">
                    <label htmlFor={`winner-${event.id}`} className="field-label">
                      Победивший исход
                    </label>
                    <select
                      id={`winner-${event.id}`}
                      className="text-input max-w-40"
                      value={winner}
                      onChange={(e) => setWinnerByEvent((prev) => ({ ...prev, [event.id]: e.target.value as "yes" | "no" }))}
                      disabled={settleMutation.isPending}
                    >
                      <option value="yes">YES</option>
                      <option value="no">NO</option>
                    </select>
                    <button
                      type="button"
                      className="btn-primary"
                      disabled={settleMutation.isPending}
                      onClick={() => settleMutation.mutate({ eventId: event.id, winner })}
                    >
                      {settleMutation.isPending ? "Завершаем..." : "Завершить событие"}
                    </button>
                  </div>
                </article>
              );
            })}
          </div>
        </section>
      )}

      {settleMutation.isError && (
        <p className="border border-red-200 bg-red-50 p-4 text-sm text-red-700">
          {settleMutation.error instanceof ApiError ? settleMutation.error.message : "Не удалось завершить событие"}
        </p>
      )}
    </div>
  );
}
