import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Link } from "react-router-dom";
import { Loader2, ArrowLeft, Mail } from "lucide-react";
import { toast } from "sonner";

import AuthShell from "@/components/layout/AuthShell";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { authApi } from "@/api/auth";

const forgotSchema = z.object({
  email: z.string().email("Invalid email address"),
});

type ForgotForm = z.infer<typeof forgotSchema>;

function ForgotPassword() {
  const [isLoading, setIsLoading] = useState(false);
  const [sent, setSent] = useState(false);
  const [sentEmail, setSentEmail] = useState("");

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<ForgotForm>({
    resolver: zodResolver(forgotSchema),
  });

  const onSubmit = async (data: ForgotForm) => {
    setIsLoading(true);
    try {
      await authApi.forgotPassword(data.email);
      setSentEmail(data.email);
      setSent(true);
      toast.success("Reset code sent to your email");
    } catch {
      toast.error("Something went wrong");
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <AuthShell
      title={sent ? "Check your email" : "Reset your password"}
      description={
        sent
          ? `We sent a 6-digit code to ${sentEmail}`
          : "We'll email you a code to reset your password"
      }
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
      {sent ? (
        <div className="space-y-5">
          <div className="flex items-center justify-center">
            <div className="flex h-12 w-12 items-center justify-center rounded-full bg-brand/10">
              <Mail className="h-5 w-5 text-brand" />
            </div>
          </div>
          <p className="text-center text-sm leading-relaxed text-muted-foreground">
            The code expires in 15 minutes. Didn't get it? Check your spam folder
            or try again.
          </p>
          <Link to="/reset-password" className="block">
            <Button variant="brand" size="xl" className="w-full">
              Enter reset code
            </Button>
          </Link>
          <button
            type="button"
            onClick={() => setSent(false)}
            className="w-full text-center text-xs text-muted-foreground hover:text-foreground"
          >
            Use a different email
          </button>
        </div>
      ) : (
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-5">
          <div className="space-y-1.5">
            <Label htmlFor="email">Email address</Label>
            <Input
              id="email"
              type="email"
              placeholder="you@example.com"
              autoComplete="email"
              autoFocus
              {...register("email")}
            />
            {errors.email && (
              <p className="text-xs text-destructive">{errors.email.message}</p>
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
            Send reset code
          </Button>
        </form>
      )}
    </AuthShell>
  );
}

export default ForgotPassword;
