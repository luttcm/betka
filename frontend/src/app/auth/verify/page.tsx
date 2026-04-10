import { AuthVerifyForm } from "@/components/auth-verify-form";

export default function VerifyPage() {
  return (
    <section className="space-y-6">
      <div className="panel-dark">
        <h2 className="text-4xl font-medium leading-[1.05] md:text-5xl">Подтверждение email</h2>
        <p className="mt-3 text-sm text-slate-600 md:text-base">
          Вставьте token из ссылки `/v1/auth/verify-email?token=...`, если автопереход не настроен.
        </p>
      </div>

      <AuthVerifyForm />
    </section>
  );
}
