"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { ApiError, createEvent } from "@/lib/api";
import { useAuth } from "@/lib/auth-context";

const createEventSchema = z.object({
  title: z.string().min(3, "Минимум 3 символа"),
  description: z.string().min(10, "Минимум 10 символов"),
  category: z.string().optional(),
  resolve_at: z.string().min(1, "Укажите дату и время"),
});

type CreateEventFormValues = z.infer<typeof createEventSchema>;

export function CreateEventForm() {
  const { token, isAuthenticated } = useAuth();

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<CreateEventFormValues>({
    resolver: zodResolver(createEventSchema),
    defaultValues: {
      title: "",
      description: "",
      category: "",
      resolve_at: "",
    },
  });

  const mutation = useMutation({
    mutationFn: async (values: CreateEventFormValues) => {
      const resolveAtIso = new Date(values.resolve_at).toISOString();

      return createEvent(
        {
          title: values.title,
          description: values.description,
          category: values.category ?? "",
          resolve_at: resolveAtIso,
        },
        token ?? "",
      );
    },
  });

  const onSubmit = handleSubmit(async (values) => {
    await mutation.mutateAsync(values);
    reset();
  });

  if (!isAuthenticated) {
    return (
      <div className="rounded-xl border border-amber-200 bg-amber-50 p-4 text-sm text-amber-900">
        Для создания события нужно войти в аккаунт.
      </div>
    );
  }

  return (
    <form onSubmit={onSubmit} className="space-y-4 rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
      <div className="grid gap-1">
        <label htmlFor="title" className="text-sm font-medium">
          Заголовок
        </label>
        <input id="title" type="text" className="rounded-md border border-slate-300 px-3 py-2 text-sm" {...register("title")} />
        {errors.title && <p className="text-sm text-red-600">{errors.title.message}</p>}
      </div>

      <div className="grid gap-1">
        <label htmlFor="description" className="text-sm font-medium">
          Описание
        </label>
        <textarea
          id="description"
          rows={4}
          className="rounded-md border border-slate-300 px-3 py-2 text-sm"
          {...register("description")}
        />
        {errors.description && <p className="text-sm text-red-600">{errors.description.message}</p>}
      </div>

      <div className="grid gap-1">
        <label htmlFor="category" className="text-sm font-medium">
          Категория
        </label>
        <input id="category" type="text" className="rounded-md border border-slate-300 px-3 py-2 text-sm" {...register("category")} />
      </div>

      <div className="grid gap-1">
        <label htmlFor="resolve_at" className="text-sm font-medium">
          Дата и время решения
        </label>
        <input
          id="resolve_at"
          type="datetime-local"
          className="rounded-md border border-slate-300 px-3 py-2 text-sm"
          {...register("resolve_at")}
        />
        {errors.resolve_at && <p className="text-sm text-red-600">{errors.resolve_at.message}</p>}
      </div>

      {mutation.isError && (
        <p className="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-700">
          {mutation.error instanceof ApiError ? mutation.error.message : "Не удалось создать событие"}
        </p>
      )}

      {mutation.isSuccess && (
        <p className="rounded-md border border-emerald-200 bg-emerald-50 p-3 text-sm text-emerald-700">
          Событие создано: {mutation.data.id}. Сейчас оно в статусе {mutation.data.status}.
        </p>
      )}

      <button
        type="submit"
        disabled={mutation.isPending}
        className="inline-flex items-center rounded-md bg-slate-900 px-4 py-2 text-sm font-medium text-white disabled:opacity-50"
      >
        {mutation.isPending ? "Создаём..." : "Создать событие"}
      </button>
    </form>
  );
}
