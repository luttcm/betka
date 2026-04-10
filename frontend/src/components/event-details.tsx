"use client";

import Link from "next/link";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useMemo, useState } from "react";

import { ApiError, getEventById, placeBet } from "@/lib/api";
import { ErrorState, LoadingState } from "@/components/ui-states";
import { useAuth } from "@/lib/auth-context";
import { PlaceBetPayload } from "@/lib/types";

interface EventDetailsProps {
  eventId: string;
}

const DEFAULT_EVENT_ODDS = {
  yes: 2,
  no: 2,
} as const;

function formatOdds(value: number): string {
  return new Intl.NumberFormat("ru-RU", {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(value);
}

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
    dateStyle: "full",
    timeStyle: "short",
  }).format(date);
}

export function EventDetails({ eventId }: EventDetailsProps) {
  const queryClient = useQueryClient();
  const { canModerate, isAuthenticated, token } = useAuth();
  const [selectedOutcome, setSelectedOutcome] = useState<"yes" | "no">("yes");
  const [stake, setStake] = useState<string>("100");

  const { data, isLoading, isError, error } = useQuery({
    queryKey: ["event", eventId],
    queryFn: () => getEventById(eventId),
  });

  const betMutation = useMutation({
    mutationFn: async (payload: PlaceBetPayload) => {
      const idempotencyKey =
        typeof crypto !== "undefined" && typeof crypto.randomUUID === "function"
          ? crypto.randomUUID()
          : `${Date.now()}-${Math.random().toString(36).slice(2)}`;

      return placeBet(payload, token ?? "", idempotencyKey);
    },
    onSuccess: async () => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ["wallet"] }),
        queryClient.invalidateQueries({ queryKey: ["wallet-transactions"] }),
        queryClient.invalidateQueries({ queryKey: ["my-bets"] }),
      ]);
    },
  });

  const canBetOnEvent = data?.status === "approved";
  const parsedStake = Number(stake);
  const isStakeInvalid = Number.isNaN(parsedStake) || parsedStake <= 0;
  const selectedOdds = DEFAULT_EVENT_ODDS[selectedOutcome];
  const potentialPayout = !isStakeInvalid ? parsedStake * selectedOdds : 0;
  const placeBetError = useMemo(() => {
    if (!(betMutation.error instanceof ApiError)) {
      return "Не удалось разместить ставку";
    }

    if (betMutation.error.status === 401) {
      return "Нужна авторизация для ставки";
    }

    if (betMutation.error.message.includes("insufficient funds")) {
      return "Недостаточно средств на кошельке";
    }

    if (betMutation.error.message.includes("event is unavailable for betting")) {
      return "Событие недоступно для ставок";
    }

    return betMutation.error.message;
  }, [betMutation.error]);

  const submitBet = async () => {
    if (!isAuthenticated || !token || isStakeInvalid || !canBetOnEvent) {
      return;
    }

    await betMutation.mutateAsync({
      event_id: eventId,
      outcome_code: selectedOutcome,
      stake: parsedStake,
    });
  };

  if (isLoading) {
    return <LoadingState message="Загрузка события..." />;
  }

  if (isError) {
    const message = error instanceof ApiError ? error.message : "Не удалось получить событие";
    return <ErrorState message={message} />;
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

      <section className="rounded-2xl border border-[color:var(--muted-border)] bg-[var(--cool-surface)] p-4">
        <h3 className="text-lg font-semibold text-slate-900">Сделать ставку</h3>
        <p className="mt-1 text-sm text-slate-600">
          Выберите исход, укажите сумму и отправьте ставку с защитой от дублей по <code>Idempotency-Key</code>.
        </p>

        <div className="mt-3 grid gap-2 rounded-xl border border-[color:var(--muted-border)] bg-white p-3 text-sm text-slate-700 md:grid-cols-2">
          <p>
            Коэффициент <b>YES</b>: <span className="font-semibold">{formatOdds(DEFAULT_EVENT_ODDS.yes)}</span>
          </p>
          <p>
            Коэффициент <b>NO</b>: <span className="font-semibold">{formatOdds(DEFAULT_EVENT_ODDS.no)}</span>
          </p>
        </div>

        {!isAuthenticated && (
          <p className="mt-3 rounded-md border border-amber-200 bg-amber-50 p-3 text-sm text-amber-900">
            Для ставки нужно <Link href="/auth/login">войти</Link> в аккаунт.
          </p>
        )}

        {isAuthenticated && !canBetOnEvent && (
          <p className="mt-3 rounded-md border border-amber-200 bg-amber-50 p-3 text-sm text-amber-900">
            Ставки доступны только для событий в статусе <b>approved</b>.
          </p>
        )}

        <div className="mt-4 grid gap-3 md:grid-cols-2">
          <div className="grid gap-1">
            <label htmlFor="bet-outcome" className="field-label">
              Исход
            </label>
            <select
              id="bet-outcome"
              className="text-input"
              value={selectedOutcome}
              onChange={(e) => setSelectedOutcome(e.target.value as "yes" | "no")}
              disabled={!isAuthenticated || !canBetOnEvent || betMutation.isPending}
            >
              <option value="yes">YES (x{formatOdds(DEFAULT_EVENT_ODDS.yes)})</option>
              <option value="no">NO (x{formatOdds(DEFAULT_EVENT_ODDS.no)})</option>
            </select>
          </div>

          <div className="grid gap-1">
            <label htmlFor="bet-stake" className="field-label">
              Сумма ставки
            </label>
            <input
              id="bet-stake"
              type="number"
              min="0.01"
              step="0.01"
              className="text-input"
              value={stake}
              onChange={(e) => setStake(e.target.value)}
              disabled={!isAuthenticated || !canBetOnEvent || betMutation.isPending}
            />
          </div>
        </div>

        {isAuthenticated && isStakeInvalid && (
          <p className="mt-3 rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700">
            Укажите корректную сумму больше 0.
          </p>
        )}

        {!isStakeInvalid && (
          <p className="mt-3 rounded-md border border-blue-200 bg-blue-50 p-3 text-sm text-blue-800">
            Потенциальная выплата по выбранному исходу: <b>{formatAmount(potentialPayout)} TOK</b> (коэфф. x
            {formatOdds(selectedOdds)}).
          </p>
        )}

        {betMutation.isError && (
          <p className="mt-3 rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700">{placeBetError}</p>
        )}

        {betMutation.isSuccess && (
          <p className="mt-3 rounded-md border border-emerald-200 bg-emerald-50 p-3 text-sm text-emerald-700">
            Ставка успешно размещена: {betMutation.data.id}
          </p>
        )}

        <button
          type="button"
          className="btn-primary mt-4"
          onClick={submitBet}
          disabled={!isAuthenticated || !canBetOnEvent || isStakeInvalid || betMutation.isPending}
        >
          {betMutation.isPending ? "Размещаем ставку..." : "Поставить"}
        </button>
      </section>

      {canModerate && (
        <section className="rounded-2xl border border-[color:var(--muted-border)] bg-[var(--cool-surface)] p-4 text-sm text-slate-700">
          Для действий модерации используйте вкладку <Link href="/moderation">«Модерация»</Link>.
        </section>
      )}
    </article>
  );
}
