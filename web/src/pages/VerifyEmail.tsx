import { useEffect, useState } from "react";
import { useSearchParams, Link } from "react-router-dom";
import { CheckCircle2, XCircle, Loader2 } from "lucide-react";

import AuthShell from "@/components/layout/AuthShell";
import { Button } from "@/components/ui/button";
import { authApi } from "@/api/auth";

function VerifyEmail() {
  const [searchParams] = useSearchParams();
  const token = searchParams.get("token");
  const [status, setStatus] = useState<"loading" | "success" | "error">(
    "loading"
  );

  useEffect(() => {
    if (!token) {
      setStatus("error");
      return;
    }

    authApi
      .verifyEmail(token)
      .then(() => setStatus("success"))
      .catch(() => setStatus("error"));
  }, [token]);

  const config = {
    loading: {
      title: "Verifying your email",
      description: "Just a moment...",
    },
    success: {
      title: "Email verified",
      description: "Your email address has been confirmed",
    },
    error: {
      title: "Verification failed",
      description: "This link is invalid or has expired",
    },
  }[status];

  return (
    <AuthShell
      title={config.title}
      description={config.description}
      footer={
        <Link to="/login" className="text-muted-foreground hover:text-foreground">
          Back to sign in
        </Link>
      }
    >
      <div className="flex flex-col items-center gap-5 py-4">
        {status === "loading" && (
          <div className="flex h-16 w-16 items-center justify-center rounded-full bg-muted">
            <Loader2 className="h-7 w-7 animate-spin text-muted-foreground" />
          </div>
        )}
        {status === "success" && (
          <>
            <div className="flex h-16 w-16 items-center justify-center rounded-full bg-success/10 ring-8 ring-success/5">
              <CheckCircle2 className="h-8 w-8 text-success" />
            </div>
            <p className="text-center text-sm leading-relaxed text-muted-foreground">
              You can now sign in to your account and start deploying.
            </p>
            <Link to="/login" className="w-full">
              <Button variant="brand" size="xl" className="w-full">
                Continue to sign in
              </Button>
            </Link>
          </>
        )}
        {status === "error" && (
          <>
            <div className="flex h-16 w-16 items-center justify-center rounded-full bg-destructive/10 ring-8 ring-destructive/5">
              <XCircle className="h-8 w-8 text-destructive" />
            </div>
            <p className="text-center text-sm leading-relaxed text-muted-foreground">
              Request a new verification email from your account settings after
              signing in.
            </p>
            <Link to="/login" className="w-full">
              <Button variant="outline" size="xl" className="w-full">
                Back to sign in
              </Button>
            </Link>
          </>
        )}
      </div>
    </AuthShell>
  );
}

export default VerifyEmail;
