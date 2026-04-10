import { EventsList } from "@/components/events-list";

export default function HomePage() {
  return (
    <section className="space-y-6">
      <div className="panel">
        <p className="text-xs uppercase tracking-[0.2em] text-[var(--brand-accent)]">Live Market</p>
        <h2 className="mt-2 text-4xl font-medium leading-[1.05] md:text-5xl">Каталог событий</h2>
        <p className="mt-4 max-w-2xl text-sm text-slate-600 md:text-base">
          Опубликованные события, доступные для ставок. Для создания собственных событий перейдите во вкладку создания.
        </p>
      </div>

      <EventsList />
    </section>
  );
}
