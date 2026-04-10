"use client";

import { EmptyState } from "@/components/ui-states";
import { ModerationEventsPanel } from "@/components/moderation-events-panel";
import { useAuth } from "@/lib/auth-context";

export default function ModerationPage() {
  const { canModerate, isAuthenticated } = useAuth();

  if (!isAuthenticated) {
    return <EmptyState message="Для доступа к модерации нужно войти в аккаунт." />;
  }

  if (!canModerate) {
    return <EmptyState message="Раздел модерации доступен только для moderator/admin." />;
  }

  return (
    <section className="space-y-6">
      <div className="panel">
        <p className="text-xs uppercase tracking-[0.2em] text-[var(--brand-accent)]">Admin Desk</p>
        <h2 className="mt-2 text-4xl font-medium leading-[1.05] md:text-5xl">Очередь модерации событий</h2>
        <p className="mt-4 max-w-2xl text-sm text-slate-600 md:text-base">
          Проверяйте пользовательские события и переводите их дальше в публикацию или отменяйте с причиной.
        </p>
      </div>

      <ModerationEventsPanel />
    </section>
  );
}
