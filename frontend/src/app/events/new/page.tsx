import { CreateEventForm } from "@/components/create-event-form";

export default function NewEventPage() {
  return (
    <section className="space-y-6">
      <div className="panel-dark">
        <p className="text-xs uppercase tracking-[0.2em] text-[var(--brand-accent)]">Create Market</p>
        <h2 className="text-4xl font-medium leading-[1.05] md:text-5xl">Создание события</h2>
        <p className="mt-3 text-sm text-slate-600 md:text-base">Событие может создать только авторизованный пользователь.</p>
      </div>

      <CreateEventForm />
    </section>
  );
}
