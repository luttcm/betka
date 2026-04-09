import { AuthRegisterForm } from "@/components/auth-register-form";

export default function RegisterPage() {
  return (
    <section className="space-y-4">
      <div>
        <h2 className="text-2xl font-semibold">Регистрация</h2>
        <p className="text-sm text-slate-600">Создайте аккаунт, затем подтвердите email и выполните вход.</p>
      </div>

      <AuthRegisterForm />
    </section>
  );
}

