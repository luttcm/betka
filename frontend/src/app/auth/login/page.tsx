"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";

import { AuthLoginForm } from "@/components/auth-login-form";
import { useAuth } from "@/lib/auth-context";

export default function LoginPage() {
  const { isAuthenticated } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (isAuthenticated) {
      router.replace("/");
    }
  }, [isAuthenticated, router]);

  if (isAuthenticated) {
    return null;
  }

  return (
    <section className="space-y-6">
      <div className="panel-dark">
        <p className="text-xs uppercase tracking-[0.2em] text-[var(--brand-accent)]">Account Access</p>
        <h2 className="text-4xl font-medium leading-[1.05] md:text-5xl">Вход</h2>
        <p className="mt-3 text-sm text-slate-600 md:text-base">Войдите после подтверждения email.</p>
      </div>

      <AuthLoginForm />
    </section>
  );
}
