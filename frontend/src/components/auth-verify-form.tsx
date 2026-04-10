"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { ApiError, verifyEmail } from "@/lib/api";

const verifySchema = z.object({
  token: z.string().min(1, "Введите токен подтверждения"),
});

type VerifyFormValues = z.infer<typeof verifySchema>;

export function AuthVerifyForm() {
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<VerifyFormValues>({
    resolver: zodResolver(verifySchema),
    defaultValues: {
      token: "",
    },
  });

  const mutation = useMutation({
    mutationFn: async (values: VerifyFormValues) => verifyEmail(values.token),
  });

  const onSubmit = handleSubmit(async (values) => {
    await mutation.mutateAsync(values);
  });

  return (
    <form onSubmit={onSubmit} className="panel space-y-4">
      <div className="grid gap-1">
        <label htmlFor="token" className="field-label">
          Токен подтверждения email
        </label>
        <input
          id="token"
          type="text"
          className="text-input"
          placeholder="Токен из ссылки verify-email"
          {...register("token")}
        />
        {errors.token && <p className="text-sm text-red-700">{errors.token.message}</p>}
      </div>

      {mutation.isError && (
        <p className="border border-red-200 bg-red-50 p-3 text-sm text-red-700">
          {mutation.error instanceof ApiError ? mutation.error.message : "Ошибка подтверждения"}
        </p>
      )}

      {mutation.isSuccess && (
        <p className="border border-emerald-200 bg-emerald-50 p-3 text-sm text-emerald-700">
          Email подтвержден. Теперь можно входить в аккаунт.
        </p>
      )}

      <button
        type="submit"
        disabled={mutation.isPending}
        className="btn-primary"
      >
        {mutation.isPending ? "Проверяем..." : "Подтвердить email"}
      </button>
    </form>
  );
}
