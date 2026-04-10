"use client";

import { useQuery } from "@tanstack/react-query";

import { EmptyState, ErrorState, LoadingState } from "@/components/ui-states";
import { MyBetsList } from "@/components/my-bets-list";
import { ApiError, getMyBets } from "@/lib/api";
import { useAuth } from "@/lib/auth-context";

export default function MyBetsPage() {
  const { isAuthenticated, token } = useAuth();

  const betsQuery = useQuery({
    queryKey: ["my-bets"],
    queryFn: () => getMyBets(token ?? ""),
    enabled: isAuthenticated && Boolean(token),
  });

  if (!isAuthenticated) {
    return <EmptyState message="Для просмотра ставок нужно войти в аккаунт." />;
  }

  if (betsQuery.isLoading) {
    return <LoadingState message="Загрузка ваших ставок..." />;
  }

  if (betsQuery.isError) {
    const message = betsQuery.error instanceof ApiError ? betsQuery.error.message : "Не удалось загрузить ставки";
    return <ErrorState message={message} />;
  }

  if (!betsQuery.data || betsQuery.data.length === 0) {
    return <EmptyState message="У вас пока нет ставок." />;
  }

  return (
    <section className="space-y-6">
      <div className="panel">
        <p className="text-xs uppercase tracking-[0.2em] text-[var(--brand-accent)]">Bet History</p>
        <h2 className="text-4xl font-medium leading-[1.05] md:text-5xl">Мои ставки</h2>
        <p className="mt-3 text-sm text-slate-600 md:text-base">Список ставок с сортировкой по времени и фильтром по статусу.</p>
      </div>

      <MyBetsList items={betsQuery.data} />
    </section>
  );
}
