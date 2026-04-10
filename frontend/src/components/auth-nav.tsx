"use client";

import Link from "next/link";

import { useAuth } from "@/lib/auth-context";

export function AuthNav() {
  const { isAuthenticated, email, role, canModerate, signOut } = useAuth();

  if (!isAuthenticated) {
    return (
      <div className="flex items-center gap-3 text-sm">
        <Link href="/auth/register" className="nav-link">
          Регистрация
        </Link>
        <Link href="/auth/login" className="nav-link">
          Вход
        </Link>
      </div>
    );
  }

  return (
    <div className="flex flex-wrap items-center gap-3 text-sm">
      <Link href="/wallet" className="nav-link">
        Кошелёк
      </Link>
      <Link href="/bets/my" className="nav-link">
        Мои ставки
      </Link>
      {canModerate && (
        <Link href="/moderation" className="nav-link">
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
        className="nav-link-button"
      >
        Выйти
      </button>
    </div>
  );
}
