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
          <header className="border-b border-[color:var(--muted-border)] bg-[#4a76a8] text-white">
            <div className="mx-auto flex w-full max-w-6xl items-center justify-between gap-4 px-4 py-3 md:px-6">
              <div className="flex items-center gap-4">
                <Link href="/" className="text-sm font-semibold text-white hover:text-white/90">
                  ДОДЕП (Аренда 60-54-04)
                </Link>
                <nav className="flex items-center gap-2">
                  <Link href="/" className="rounded-md bg-white/15 px-3 py-1.5 text-sm text-white hover:bg-white/15">
                    Каталог
                  </Link>
                  <Link href="/events/new" className="rounded-md px-3 py-1.5 text-sm text-white/90 hover:bg-white/15 hover:text-white">
                    Создать событие
                  </Link>
                </nav>
              </div>
              <AuthNav />
            </div>
          </header>

          <main className="app-shell">
            <header className="mb-6">
              <h1 className="text-3xl font-semibold leading-tight text-[#2c2d2e] md:text-4xl">Линия пользовательских событий</h1>
            </header>

            {children}
          </main>
        </Providers>
      </body>
    </html>
  );
}
