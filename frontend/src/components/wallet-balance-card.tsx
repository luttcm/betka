import { Wallet } from "@/lib/types";

interface WalletBalanceCardProps {
  wallet: Wallet;
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
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date);
}

export function WalletBalanceCard({ wallet }: WalletBalanceCardProps) {
  return (
    <article className="panel">
      <p className="text-sm text-slate-600">Текущий баланс</p>
      <p className="mt-2 text-3xl font-semibold text-slate-900">{formatAmount(wallet.balance_tokens)} TOK</p>
      <p className="mt-2 text-xs text-slate-500">Обновлено: {formatDate(wallet.updated_at)}</p>
    </article>
  );
}

