import { AuthLoginForm } from "@/components/auth-login-form";

export default function LoginPage() {
  return (
    <section className="space-y-4">
      <div>
        <h2 className="text-2xl font-semibold">Вход</h2>
        <p className="text-sm text-slate-600">Войдите после подтверждения email.</p>
      </div>

      <AuthLoginForm />
    </section>
  );
}

