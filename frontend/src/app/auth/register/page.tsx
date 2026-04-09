import { AuthRegisterForm } from "@/components/auth-register-form";

export default function RegisterPage() {
  return (
    <section className="space-y-6">
      <div className="panel-dark">
        <h2 className="text-4xl font-medium leading-[1.05] md:text-5xl">Регистрация</h2>
        <p className="mt-3 text-sm text-white/75 md:text-base">Создайте аккаунт, затем подтвердите email и выполните вход.</p>
      </div>

      <AuthRegisterForm />
    </section>
  );
}
