import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { useNavigate } from "react-router-dom";
import { Loader2 } from "lucide-react";
import { toast } from "sonner";

import { AuthShell } from "@/components/layout/AuthShell";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { authApi } from "@/api/auth";
import { useAuthStore } from "@/stores/auth";

const schema = z
  .object({
    current_password: z.string().min(1, "Enter the password you were given"),
    new_password: z.string().min(8, "Use at least 8 characters"),
    confirm_password: z.string().min(1, "Confirm your new password"),
  })
  .refine((v) => v.new_password === v.confirm_password, {
    path: ["confirm_password"],
    message: "Passwords don't match",
  })
  .refine((v) => v.new_password !== v.current_password, {
    path: ["new_password"],
    message: "Choose something different from the password you were given",
  });

type Form = z.infer<typeof schema>;

/**
 * Shown when an account still holds the password an admin handed over.
 * The API refuses every other request until it's replaced, so there is
 * deliberately no way to skip this.
 */
function ChangePassword() {
  const navigate = useNavigate();
  const setAccessToken = useAuthStore((s) => s.setAccessToken);
  const [isLoading, setIsLoading] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<Form>({ resolver: zodResolver(schema) });

  const onSubmit = async (data: Form) => {
    setIsLoading(true);
    try {
      const res = await authApi.changePassword({
        current_password: data.current_password,
        new_password: data.new_password,
      });
      // The change revokes all sessions, so adopt the fresh token or the very
      // next request would 401.
      const token = res.data?.data?.access_token;
      if (token) setAccessToken(token);
      toast.success("Password updated");
      navigate("/dashboard");
    } catch (err: unknown) {
      const message =
        (err as { response?: { data?: { error?: { message?: string } } } })
          ?.response?.data?.error?.message || "Could not change password";
      toast.error(message);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <AuthShell
      title="Set your password"
      description="Your account was created for you. Choose your own password to continue."
    >
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        <div className="space-y-1.5">
          <Label htmlFor="current_password">Password you were given</Label>
          <Input id="current_password" type="password" autoComplete="current-password" {...register("current_password")} />
          {errors.current_password && (
            <p className="text-xs text-destructive">{errors.current_password.message}</p>
          )}
        </div>

        <div className="space-y-1.5">
          <Label htmlFor="new_password">New password</Label>
          <Input id="new_password" type="password" autoComplete="new-password" {...register("new_password")} />
          {errors.new_password && (
            <p className="text-xs text-destructive">{errors.new_password.message}</p>
          )}
        </div>

        <div className="space-y-1.5">
          <Label htmlFor="confirm_password">Confirm new password</Label>
          <Input id="confirm_password" type="password" autoComplete="new-password" {...register("confirm_password")} />
          {errors.confirm_password && (
            <p className="text-xs text-destructive">{errors.confirm_password.message}</p>
          )}
        </div>

        <Button type="submit" variant="brand" className="w-full" disabled={isLoading}>
          {isLoading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
          Set password and continue
        </Button>
      </form>
    </AuthShell>
  );
}

export default ChangePassword;
