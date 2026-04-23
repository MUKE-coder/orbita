import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Link, useNavigate } from "react-router-dom";
import { Loader2, Check } from "lucide-react";
import { toast } from "sonner";

import AuthShell from "@/components/layout/AuthShell";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { authApi } from "@/api/auth";
import { useAuthStore } from "@/stores/auth";

const registerSchema = z.object({
  name: z.string().min(2, "Name must be at least 2 characters"),
  email: z.string().email("Invalid email address"),
  password: z.string().min(8, "Password must be at least 8 characters"),
});

type RegisterForm = z.infer<typeof registerSchema>;

function PasswordHint({ password }: { password: string }) {
  const rules = [
    { label: "At least 8 characters", pass: password.length >= 8 },
    { label: "Contains a number", pass: /\d/.test(password) },
    { label: "Contains a letter", pass: /[a-zA-Z]/.test(password) },
  ];
  return (
    <ul className="mt-2 space-y-1">
      {rules.map((r) => (
        <li
          key={r.label}
          className={`flex items-center gap-1.5 text-[11px] ${
            r.pass ? "text-success" : "text-muted-foreground"
          }`}
        >
          <Check className={`h-3 w-3 ${r.pass ? "opacity-100" : "opacity-40"}`} />
          {r.label}
        </li>
      ))}
    </ul>
  );
}

function Register() {
  const navigate = useNavigate();
  const setAccessToken = useAuthStore((s) => s.setAccessToken);
  const [isLoading, setIsLoading] = useState(false);

  const {
    register,
    handleSubmit,
    watch,
    formState: { errors },
  } = useForm<RegisterForm>({
    resolver: zodResolver(registerSchema),
  });

  const password = watch("password") || "";

  const onSubmit = async (data: RegisterForm) => {
    setIsLoading(true);
    try {
      const res = await authApi.register(data);
      setAccessToken(res.data.data.access_token);
      toast.success("Account created — check your email to verify.");
      navigate("/dashboard");
    } catch (err: unknown) {
      const message =
        (err as { response?: { data?: { error?: { message?: string } } } })
          ?.response?.data?.error?.message || "Registration failed";
      toast.error(message);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <AuthShell
      title="Create your account"
      description="Deploy your first app in under 60 seconds"
      footer={
        <>
          Already have an account?{" "}
          <Link
            to="/login"
            className="font-medium text-foreground underline-offset-4 hover:underline"
          >
            Sign in
          </Link>
        </>
      }
    >
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-5">
        <div className="space-y-1.5">
          <Label htmlFor="name">Full name</Label>
          <Input
            id="name"
            placeholder="Ada Lovelace"
            autoComplete="name"
            autoFocus
            {...register("name")}
          />
          {errors.name && (
            <p className="text-xs text-destructive">{errors.name.message}</p>
          )}
        </div>

        <div className="space-y-1.5">
          <Label htmlFor="email">Work email</Label>
          <Input
            id="email"
            type="email"
            placeholder="you@example.com"
            autoComplete="email"
            {...register("email")}
          />
          {errors.email && (
            <p className="text-xs text-destructive">{errors.email.message}</p>
          )}
        </div>

        <div className="space-y-1.5">
          <Label htmlFor="password">Password</Label>
          <Input
            id="password"
            type="password"
            placeholder="Min. 8 characters"
            autoComplete="new-password"
            {...register("password")}
          />
          {errors.password && (
            <p className="text-xs text-destructive">{errors.password.message}</p>
          )}
          {password && <PasswordHint password={password} />}
        </div>

        <Button
          type="submit"
          variant="brand"
          size="xl"
          className="w-full"
          disabled={isLoading}
        >
          {isLoading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
          Create account
        </Button>

        <p className="text-center text-[11px] leading-relaxed text-muted-foreground">
          By creating an account, you agree to our{" "}
          <a href="#" className="underline-offset-4 hover:underline">Terms</a> and{" "}
          <a href="#" className="underline-offset-4 hover:underline">Privacy Policy</a>.
        </p>
      </form>
    </AuthShell>
  );
}

export default Register;
