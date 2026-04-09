import Link from "next/link";

import { EventDetails } from "@/components/event-details";

interface EventPageProps {
  params: Promise<{ id: string }>;
}

export default async function EventPage({ params }: EventPageProps) {
  const { id } = await params;

  return (
    <section className="space-y-6">
      <div className="panel-dark flex flex-wrap items-center justify-between gap-3">
        <h2 className="text-4xl font-medium leading-[1.05] md:text-5xl">Карточка события</h2>
        <Link href="/" className="btn-secondary !rounded-full !px-4 !py-2 !text-xs">
          ← Назад в каталог
        </Link>
      </div>

      <EventDetails eventId={id} />
    </section>
  );
}
