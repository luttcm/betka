import type { Metadata } from "next";
import Link from "next/link";
import { ReactNode } from "react";

import "@/app/globals.css";
import { Providers } from "@/app/providers";
import { AuthNav } from "@/components/auth-nav";

export const metadata: Metadata = {
  title: "Bet MVP Frontend",
  description: "Frontend for bet MVP platform",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: ReactNode;
}>) {
  return (
    <html lang="ru">
      <body>
        <Providers>
          <main className="app-shell">
            <header className="mb-8 rounded-[32px] bg-[var(--near-black)] p-6 text-white md:p-8">
              <div className="flex flex-wrap items-center justify-between gap-4">
                <div>
                  <p className="text-sm text-white/70">Bet MVP</p>
                  <h1 className="text-3xl font-medium leading-tight md:text-4xl">Пользовательские события</h1>
                </div>
                <div className="flex flex-col items-start gap-3 text-sm md:items-end">
                  <nav className="flex flex-wrap items-center gap-2">
                    <Link href="/" className="btn-secondary !rounded-full !px-4 !py-2 !text-xs">
                      Каталог
                    </Link>
                    <Link href="/events/new" className="btn-secondary !rounded-full !px-4 !py-2 !text-xs">
                      Создать событие
                    </Link>
                    <Link href="/moderation" className="btn-secondary !rounded-full !px-4 !py-2 !text-xs">
                      Модерация
                    </Link>
                  </nav>
                  <AuthNav />
                </div>
              </div>
            </header>

            {children}
          </main>
        </Providers>
      </body>
    </html>
  );
}
