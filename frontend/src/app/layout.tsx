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
          <main className="mx-auto min-h-screen w-full max-w-5xl px-4 py-8">
            <header className="mb-8 rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
              <div className="flex flex-wrap items-center justify-between gap-3">
                <div>
                  <p className="text-sm text-slate-500">Bet MVP</p>
                  <h1 className="text-xl font-semibold">Пользовательские события</h1>
                </div>
                <div className="flex items-center gap-6 text-sm">
                  <nav className="flex items-center gap-4 text-sm">
                    <Link href="/">Каталог</Link>
                    <Link href="/events/new">Создать событие</Link>
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
