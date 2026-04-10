"use client";

import Link from "next/link";

import { useAuth } from "@/lib/auth-context";

export function AuthNav() {
  const { isAuthenticated, email, role, canModerate, signOut } = useAuth();

  if (!isAuthenticated) {
    return (
      <div className="flex items-center gap-2">
        <Link href="/auth/register" className="btn-secondary !rounded-full !px-4 !py-2 !text-xs">
          Регистрация
        </Link>
        <Link href="/auth/login" className="btn-primary !rounded-full !px-4 !py-2 !text-xs">
          Вход
        </Link>
      </div>
    );
  }

  return (
    <div className="flex flex-wrap items-center gap-2">
      <Link href="/wallet" className="btn-secondary !rounded-full !px-4 !py-2 !text-xs">
        Кошелёк
      </Link>
      <Link href="/bets/my" className="btn-secondary !rounded-full !px-4 !py-2 !text-xs">
        Мои ставки
      </Link>
      {canModerate && (
        <Link href="/moderation" className="btn-secondary !rounded-full !px-4 !py-2 !text-xs">
          Модерация
        </Link>
      )}
      <span className="rounded-full border border-emerald-200 bg-emerald-50 px-3 py-1 text-xs text-emerald-700">
        {email ?? "Авторизован"}
      </span>
      {role && <span className="rounded-full border border-white/20 bg-white/10 px-3 py-1 text-xs text-white">{role}</span>}
      <button
        type="button"
        onClick={signOut}
        className="btn-secondary !rounded-full !px-4 !py-2 !text-xs"
      >
        Выйти
      </button>
    </div>
  );
}
