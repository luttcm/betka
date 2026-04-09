import { AuthVerifyForm } from "@/components/auth-verify-form";

export default function VerifyPage() {
  return (
    <section className="space-y-4">
      <div>
        <h2 className="text-2xl font-semibold">Подтверждение email</h2>
        <p className="text-sm text-slate-600">
          Вставьте token из ссылки `/v1/auth/verify-email?token=...`, если автопереход не настроен.
        </p>
      </div>

      <AuthVerifyForm />
    </section>
  );
}

