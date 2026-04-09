import Link from "next/link";

import { EventDetails } from "@/components/event-details";

interface EventPageProps {
  params: Promise<{ id: string }>;
}

export default async function EventPage({ params }: EventPageProps) {
  const { id } = await params;

  return (
    <section className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-2xl font-semibold">Карточка события</h2>
        <Link href="/" className="text-sm">
          ← Назад в каталог
        </Link>
      </div>

      <EventDetails eventId={id} />
    </section>
  );
}

