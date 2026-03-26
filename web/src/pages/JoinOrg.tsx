import { useEffect, useState } from "react";
import { useSearchParams, useNavigate } from "react-router-dom";
import { Loader2, CheckCircle, XCircle } from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { orgsApi } from "@/api/orgs";
import { useAuthStore } from "@/stores/auth";

function JoinOrg() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const token = searchParams.get("token");
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);

  const [status, setStatus] = useState<"loading" | "info" | "success" | "error">("loading");
  const [inviteInfo, setInviteInfo] = useState<{
    organization: string;
    role: string;
    email: string;
  } | null>(null);
  const [isAccepting, setIsAccepting] = useState(false);

  useEffect(() => {
    if (!token) {
      setStatus("error");
      return;
    }

    orgsApi
      .getInviteInfo(token)
      .then((res) => {
        setInviteInfo(res.data.data);
        setStatus("info");
      })
      .catch(() => setStatus("error"));
  }, [token]);

  const handleAccept = async () => {
    if (!token) return;

    if (!isAuthenticated) {
      navigate(`/login?redirect=/join?token=${token}`);
      return;
    }

    setIsAccepting(true);
    try {
      await orgsApi.acceptInvite(token);
      setStatus("success");
      toast.success("You have joined the organization!");
    } catch {
      toast.error("Failed to accept invitation");
    } finally {
      setIsAccepting(false);
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-background">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <CardTitle className="text-2xl">Join Organization</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col items-center gap-4">
          {status === "loading" && (
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          )}

          {status === "info" && inviteInfo && (
            <>
              <p className="text-center">
                You've been invited to join{" "}
                <strong>{inviteInfo.organization}</strong>
              </p>
              <Badge variant="secondary" className="text-sm">
                Role: {inviteInfo.role}
              </Badge>
              <Button
                onClick={handleAccept}
                className="w-full"
                disabled={isAccepting}
              >
                {isAccepting && (
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                )}
                {isAuthenticated ? "Accept Invitation" : "Sign In to Accept"}
              </Button>
            </>
          )}

          {status === "success" && (
            <>
              <CheckCircle className="h-12 w-12 text-green-500" />
              <p>You have joined the organization!</p>
              <Button onClick={() => navigate("/")}>Go to Dashboard</Button>
            </>
          )}

          {status === "error" && (
            <>
              <XCircle className="h-12 w-12 text-destructive" />
              <p className="text-muted-foreground">
                This invitation is invalid or has expired.
              </p>
              <Button variant="outline" onClick={() => navigate("/")}>
                Go Home
              </Button>
            </>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

export default JoinOrg;
