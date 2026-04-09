import { EventsList } from "@/components/events-list";

export default function HomePage() {
  return (
    <section className="space-y-4">
      <div>
        <h2 className="text-2xl font-semibold">Каталог событий</h2>
        <p className="text-sm text-slate-600">Опубликованные события, доступные для ставок.</p>
      </div>

      <EventsList />
    </section>
  );
}

