import { AuthLoginForm } from "@/components/auth-login-form";

export default function LoginPage() {
  return (
    <section className="space-y-6">
      <div className="panel-dark">
        <h2 className="text-4xl font-medium leading-[1.05] md:text-5xl">Вход</h2>
        <p className="mt-3 text-sm text-white/75 md:text-base">Войдите после подтверждения email.</p>
      </div>

      <AuthLoginForm />
    </section>
  );
}
