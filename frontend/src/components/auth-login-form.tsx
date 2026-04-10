"use client";

import { useRouter } from "next/navigation";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { ApiError, login } from "@/lib/api";
import { useAuth } from "@/lib/auth-context";

const loginSchema = z.object({
  email: z.string().email("Некорректный email"),
  password: z.string().min(1, "Введите пароль"),
});

type LoginFormValues = z.infer<typeof loginSchema>;

export function AuthLoginForm() {
  const { signIn, isAuthenticated } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (isAuthenticated) {
      router.replace("/");
    }
  }, [isAuthenticated, router]);

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginFormValues>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      email: "",
      password: "",
    },
  });

  const mutation = useMutation({
    mutationFn: login,
    onSuccess: (result, variables) => {
      signIn(result.access_token, variables.email);
      router.replace("/");
    },
  });

  const onSubmit = handleSubmit(async (values) => {
    await mutation.mutateAsync(values);
  });

  return (
    <form onSubmit={onSubmit} className="panel space-y-4">
      <div className="grid gap-1">
        <label htmlFor="email" className="field-label">
          Email
        </label>
        <input
          id="email"
          type="email"
          className="text-input"
          placeholder="user@example.com"
          {...register("email")}
        />
        {errors.email && <p className="text-sm text-red-700">{errors.email.message}</p>}
      </div>

      <div className="grid gap-1">
        <label htmlFor="password" className="field-label">
          Пароль
        </label>
        <input
          id="password"
          type="password"
          className="text-input"
          {...register("password")}
        />
        {errors.password && <p className="text-sm text-red-700">{errors.password.message}</p>}
      </div>

      {mutation.isError && (
        <p className="border border-red-200 bg-red-50 p-3 text-sm text-red-700">
          {mutation.error instanceof ApiError ? mutation.error.message : "Ошибка входа"}
        </p>
      )}

      {mutation.isSuccess && (
        <p className="border border-emerald-200 bg-emerald-50 p-3 text-sm text-emerald-700">
          Вход выполнен успешно.
        </p>
      )}

      <button
        type="submit"
        disabled={mutation.isPending}
        className="btn-primary"
      >
        {mutation.isPending ? "Входим..." : "Войти"}
      </button>
    </form>
  );
}
