"use client";

import { useQuery } from "@tanstack/react-query";

import { EmptyState, ErrorState, LoadingState } from "@/components/ui-states";
import { WalletBalanceCard } from "@/components/wallet-balance-card";
import { WalletTransactionsList } from "@/components/wallet-transactions-list";
import { ApiError, getWallet, getWalletTransactions } from "@/lib/api";
import { useAuth } from "@/lib/auth-context";

export default function WalletPage() {
  const { isAuthenticated, token } = useAuth();

  const walletQuery = useQuery({
    queryKey: ["wallet"],
    queryFn: () => getWallet(token ?? ""),
    enabled: isAuthenticated && Boolean(token),
  });

  const txQuery = useQuery({
    queryKey: ["wallet-transactions"],
    queryFn: () => getWalletTransactions(token ?? ""),
    enabled: isAuthenticated && Boolean(token),
  });

  if (!isAuthenticated) {
    return <EmptyState message="Для просмотра кошелька нужно войти в аккаунт." />;
  }

  if (walletQuery.isLoading || txQuery.isLoading) {
    return <LoadingState message="Загрузка данных кошелька..." />;
  }

  if (walletQuery.isError) {
    const message = walletQuery.error instanceof ApiError ? walletQuery.error.message : "Не удалось загрузить кошелёк";
    return <ErrorState message={message} />;
  }

  if (txQuery.isError) {
    const message = txQuery.error instanceof ApiError ? txQuery.error.message : "Не удалось загрузить транзакции";
    return <ErrorState message={message} />;
  }

  if (!walletQuery.data) {
    return <EmptyState message="Кошелёк не найден." />;
  }

  return (
    <section className="space-y-6">
      <div className="panel">
        <p className="text-xs uppercase tracking-[0.2em] text-[var(--brand-accent)]">Account Ledger</p>
        <h2 className="text-4xl font-medium leading-[1.05] md:text-5xl">Кошелёк</h2>
        <p className="mt-3 text-sm text-slate-600 md:text-base">Текущий баланс и история всех финансовых операций.</p>
      </div>

      <WalletBalanceCard wallet={walletQuery.data} />

      {txQuery.data && txQuery.data.length > 0 ? (
        <WalletTransactionsList items={txQuery.data} />
      ) : (
        <EmptyState message="Транзакций пока нет." />
      )}
    </section>
  );
}
