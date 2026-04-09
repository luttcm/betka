import { CreateEventForm } from "@/components/create-event-form";

export default function NewEventPage() {
  return (
    <section className="space-y-4">
      <div>
        <h2 className="text-2xl font-semibold">Создание события</h2>
        <p className="text-sm text-slate-600">Событие может создать только авторизованный пользователь.</p>
      </div>

      <CreateEventForm />
    </section>
  );
}
