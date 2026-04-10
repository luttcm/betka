"use client";

import Link from "next/link";

import { useAuth } from "@/lib/auth-context";

export function AuthNav() {
  const { isAuthenticated, email, role, canModerate, signOut } = useAuth();

  if (!isAuthenticated) {
    return (
      <div className="flex items-center gap-3 text-sm">
        <Link href="/auth/register" className="text-white/85 hover:text-white">
          Регистрация
        </Link>
        <Link href="/auth/login" className="rounded-md bg-white/15 px-3 py-1.5 text-white hover:bg-white/20">
          Вход
        </Link>
      </div>
    );
  }

  return (
    <div className="flex flex-wrap items-center gap-3 text-sm">
      <Link href="/wallet" className="text-white/90 hover:text-white">
        Кошелёк
      </Link>
      <Link href="/bets/my" className="text-white/90 hover:text-white">
        Мои ставки
      </Link>
      {canModerate && (
        <Link href="/moderation" className="text-white/90 hover:text-white">
          Модерация
        </Link>
      )}
      <span className="rounded-md bg-white/15 px-2.5 py-1 text-xs text-white/90">
        {email ?? "Авторизован"}
      </span>
      {role && (
        <span className="rounded-md bg-white/15 px-2.5 py-1 text-xs uppercase tracking-wide text-white/90">
          {role}
        </span>
      )}
      <button
        type="button"
        onClick={signOut}
        className="text-white/90 hover:text-white"
      >
        Выйти
      </button>
    </div>
  );
}
