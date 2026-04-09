"use client";

import Link from "next/link";

import { useAuth } from "@/lib/auth-context";

export function AuthNav() {
  const { isAuthenticated, email, signOut } = useAuth();

  if (!isAuthenticated) {
    return (
      <div className="flex items-center gap-3">
        <Link href="/auth/register">Регистрация</Link>
        <Link href="/auth/login">Вход</Link>
      </div>
    );
  }

  return (
    <div className="flex items-center gap-3">
      <span className="rounded bg-emerald-50 px-2 py-1 text-xs text-emerald-700">
        {email ?? "Авторизован"}
      </span>
      <button
        type="button"
        onClick={signOut}
        className="rounded border border-slate-300 px-2 py-1 text-xs text-slate-700"
      >
        Выйти
      </button>
    </div>
  );
}

