import { WalletTransaction } from "@/lib/types";

interface WalletTransactionsListProps {
  items: WalletTransaction[];
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

export function WalletTransactionsList({ items }: WalletTransactionsListProps) {
  return (
    <section className="panel space-y-4">
      <h3 className="text-xl font-semibold text-slate-900">История транзакций</h3>
      <div className="space-y-2">
        {items.map((tx) => (
          <article key={tx.id} className="rounded-xl border border-[color:var(--muted-border)] bg-[var(--cool-surface)] p-4">
            <div className="flex flex-wrap items-center justify-between gap-2">
              <p className="text-sm font-medium text-slate-800">{tx.type}</p>
              <p className="text-sm font-semibold text-slate-900">{formatAmount(tx.amount_tokens)} TOK</p>
            </div>
            <p className="mt-1 text-xs text-slate-600">{formatDate(tx.created_at)}</p>
            {(tx.ref_type || tx.ref_id) && (
              <p className="mt-1 text-xs text-slate-500">
                Связь: {tx.ref_type ?? "—"} / {tx.ref_id ?? "—"}
              </p>
            )}
          </article>
        ))}
      </div>
    </section>
  );
}

