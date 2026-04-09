import { CreateEventForm } from "@/components/create-event-form";

export default function NewEventPage() {
  return (
    <section className="space-y-6">
      <div className="panel-dark">
        <h2 className="text-4xl font-medium leading-[1.05] md:text-5xl">Создание события</h2>
        <p className="mt-3 text-sm text-white/75 md:text-base">Событие может создать только авторизованный пользователь.</p>
      </div>

      <CreateEventForm />
    </section>
  );
}
