"use client";

import Link from "next/link";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { ApiError, register } from "@/lib/api";

const registerSchema = z.object({
  email: z.string().email("Некорректный email"),
  password: z.string().min(6, "Минимум 6 символов"),
});

type RegisterFormValues = z.infer<typeof registerSchema>;

export function AuthRegisterForm() {
  const {
    register: formRegister,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<RegisterFormValues>({
    resolver: zodResolver(registerSchema),
    defaultValues: {
      email: "",
      password: "",
    },
  });

  const mutation = useMutation({
    mutationFn: register,
  });

  const onSubmit = handleSubmit(async (values) => {
    await mutation.mutateAsync(values);
    reset();
  });

  return (
    <form onSubmit={onSubmit} className="space-y-4 rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
      <div className="grid gap-1">
        <label htmlFor="email" className="text-sm font-medium">
          Email
        </label>
        <input
          id="email"
          type="email"
          className="rounded-md border border-slate-300 px-3 py-2 text-sm"
          placeholder="user@example.com"
          {...formRegister("email")}
        />
        {errors.email && <p className="text-sm text-red-600">{errors.email.message}</p>}
      </div>

      <div className="grid gap-1">
        <label htmlFor="password" className="text-sm font-medium">
          Пароль
        </label>
        <input
          id="password"
          type="password"
          className="rounded-md border border-slate-300 px-3 py-2 text-sm"
          placeholder="Минимум 6 символов"
          {...formRegister("password")}
        />
        {errors.password && <p className="text-sm text-red-600">{errors.password.message}</p>}
      </div>

      {mutation.isError && (
        <p className="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700">
          {mutation.error instanceof ApiError ? mutation.error.message : "Ошибка регистрации"}
        </p>
      )}

      {mutation.isSuccess && (
        <div className="space-y-2 rounded-md border border-emerald-200 bg-emerald-50 p-3 text-sm text-emerald-800">
          <p>Регистрация успешна. Подтвердите email по ссылке из письма/логов API.</p>
          <p>
            После подтверждения перейдите на <Link href="/auth/login">страницу входа</Link>.
          </p>
        </div>
      )}

      <button
        type="submit"
        disabled={mutation.isPending}
        className="inline-flex items-center rounded-md bg-slate-900 px-4 py-2 text-sm font-medium text-white disabled:opacity-50"
      >
        {mutation.isPending ? "Регистрируем..." : "Зарегистрироваться"}
      </button>
    </form>
  );
}

