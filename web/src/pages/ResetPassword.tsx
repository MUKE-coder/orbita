import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { useNavigate, Link } from "react-router-dom";
import { Loader2, ArrowLeft } from "lucide-react";
import { toast } from "sonner";

import AuthShell from "@/components/layout/AuthShell";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { authApi } from "@/api/auth";

const resetSchema = z.object({
  email: z.string().email("Invalid email address"),
  otp: z.string().length(6, "Code must be 6 digits"),
  new_password: z.string().min(8, "Password must be at least 8 characters"),
});

type ResetForm = z.infer<typeof resetSchema>;

function ResetPassword() {
  const navigate = useNavigate();
  const [isLoading, setIsLoading] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<ResetForm>({
    resolver: zodResolver(resetSchema),
  });

  const onSubmit = async (data: ResetForm) => {
    setIsLoading(true);
    try {
      await authApi.resetPassword(data);
      toast.success("Password reset successfully");
      navigate("/login");
    } catch (err: unknown) {
      const message =
        (err as { response?: { data?: { error?: { message?: string } } } })
          ?.response?.data?.error?.message || "Reset failed";
      toast.error(message);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <AuthShell
      title="Set a new password"
      description="Enter the 6-digit code we sent you"
      footer={
        <Link
          to="/login"
          className="inline-flex items-center gap-1 text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-3.5 w-3.5" />
          Back to sign in
        </Link>
      }
    >
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-5">
        <div className="space-y-1.5">
          <Label htmlFor="email">Email address</Label>
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
          <Label htmlFor="otp">Reset code</Label>
          <Input
            id="otp"
            placeholder="000000"
            maxLength={6}
            inputMode="numeric"
            pattern="[0-9]{6}"
            autoComplete="one-time-code"
            className="text-center font-mono text-xl tracking-[0.5em] h-12"
            {...register("otp")}
          />
          {errors.otp && (
            <p className="text-xs text-destructive">{errors.otp.message}</p>
          )}
        </div>

        <div className="space-y-1.5">
          <Label htmlFor="new_password">New password</Label>
          <Input
            id="new_password"
            type="password"
            placeholder="Min. 8 characters"
            autoComplete="new-password"
            {...register("new_password")}
          />
          {errors.new_password && (
            <p className="text-xs text-destructive">
              {errors.new_password.message}
            </p>
          )}
        </div>

        <Button
          type="submit"
          variant="brand"
          size="xl"
          className="w-full"
          disabled={isLoading}
        >
          {isLoading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
          Reset password
        </Button>
      </form>
    </AuthShell>
  );
}

export default ResetPassword;
