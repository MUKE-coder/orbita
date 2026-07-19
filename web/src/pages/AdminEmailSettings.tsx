import { useEffect, useState } from "react";
import { useForm } from "react-hook-form";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Loader2, Mail, CheckCircle2, AlertTriangle, Send } from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { adminApi } from "@/api/admin";

type Form = {
  api_key: string;
  email_from: string;
  email_from_name: string;
};

/**
 * Instance-wide email provider settings.
 *
 * Email is optional: without it Orbita can't send invites or password resets,
 * so the super admin creates accounts with a password instead and hands the
 * credentials over. This page makes that trade-off visible rather than leaving
 * invites to fail silently.
 */
function AdminEmailSettings() {
  const qc = useQueryClient();
  const [testTo, setTestTo] = useState("");

  const { data, isLoading } = useQuery({
    queryKey: ["admin-email-settings"],
    queryFn: () => adminApi.getEmailSettings(),
  });
  const settings = data?.data?.data;

  const { register, handleSubmit, reset } = useForm<Form>({
    defaultValues: { api_key: "", email_from: "", email_from_name: "" },
  });

  useEffect(() => {
    if (settings) {
      reset({
        api_key: "",
        email_from: settings.email_from || "",
        email_from_name: settings.email_from_name || "",
      });
    }
  }, [settings, reset]);

  const save = useMutation({
    mutationFn: (d: Form) =>
      adminApi.updateEmailSettings({
        // Blank means "leave the stored key alone" — the key is never sent back
        // to the browser, so an empty field must not wipe it.
        api_key: d.api_key.trim() === "" ? undefined : d.api_key.trim(),
        email_from: d.email_from,
        email_from_name: d.email_from_name,
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["admin-email-settings"] });
      toast.success("Email settings saved");
    },
    onError: (err: any) =>
      toast.error(err?.response?.data?.error?.message || "Could not save settings"),
  });

  const clearKey = useMutation({
    mutationFn: () =>
      adminApi.updateEmailSettings({
        api_key: "",
        email_from: settings?.email_from || "",
        email_from_name: settings?.email_from_name || "",
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["admin-email-settings"] });
      toast.success("API key removed — Orbita will stop sending email");
    },
  });

  const sendTest = useMutation({
    mutationFn: () => adminApi.sendTestEmail(testTo),
    onSuccess: (res) => toast.success(res.data?.data?.message || "Test email sent"),
    onError: (err: any) =>
      toast.error(err?.response?.data?.error?.message || "Test send failed"),
  });

  if (isLoading) {
    return (
      <div className="flex h-64 items-center justify-center">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
      </div>
    );
  }

  const configured = settings?.configured;

  return (
    <div className="mx-auto max-w-2xl space-y-6 p-6">
      <div>
        <h1 className="font-heading text-xl font-semibold tracking-tight">Email</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Used for invitations, password resets, and deploy notifications.
        </p>
      </div>

      {/* Status — states plainly what does and doesn't work right now. */}
      <div
        className={`flex items-start gap-3 rounded-xl border p-4 ${
          configured
            ? "border-green-500/25 bg-green-500/5"
            : "border-amber-500/25 bg-amber-500/5"
        }`}
      >
        {configured ? (
          <CheckCircle2 className="mt-0.5 h-4 w-4 shrink-0 text-green-500" />
        ) : (
          <AlertTriangle className="mt-0.5 h-4 w-4 shrink-0 text-amber-500" />
        )}
        <div className="text-sm">
          <p className="font-medium">
            {configured ? "Email is configured" : "Email is not configured"}
          </p>
          <p className="mt-0.5 text-muted-foreground">
            {configured ? (
              <>
                Invitations and password resets will be delivered.
                {settings?.source === "environment" && (
                  <> Values come from environment variables.</>
                )}
              </>
            ) : (
              <>
                Invites and password resets can't be sent. Create users directly
                from <strong>Organisations → New tenant</strong> and hand over the
                generated password instead.
              </>
            )}
          </p>
        </div>
      </div>

      <form onSubmit={handleSubmit((d) => save.mutate(d))} className="space-y-4 rounded-xl border border-border bg-card p-5">
        <div className="space-y-1.5">
          <Label htmlFor="api_key">Resend API key</Label>
          <Input
            id="api_key"
            type="password"
            autoComplete="off"
            placeholder={settings?.has_api_key ? "•••••••• (leave blank to keep)" : "re_..."}
            {...register("api_key")}
          />
          <p className="text-[11px] text-muted-foreground">
            Stored encrypted. Get one at{" "}
            <a href="https://resend.com/api-keys" target="_blank" rel="noreferrer" className="text-brand hover:underline">
              resend.com/api-keys
            </a>
            . Overrides the <code>RESEND_API_KEY</code> environment variable.
          </p>
        </div>

        <div className="grid gap-4 sm:grid-cols-2">
          <div className="space-y-1.5">
            <Label htmlFor="email_from">From address</Label>
            <Input id="email_from" placeholder="noreply@yourdomain.com" {...register("email_from")} />
            <p className="text-[11px] text-muted-foreground">
              Must be on a domain you've verified with Resend.
            </p>
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="email_from_name">From name</Label>
            <Input id="email_from_name" placeholder="Orbita" {...register("email_from_name")} />
          </div>
        </div>

        <div className="flex flex-wrap items-center gap-3 pt-1">
          <Button type="submit" variant="brand" disabled={save.isPending}>
            {save.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
            Save
          </Button>
          {settings?.has_api_key && settings.source === "dashboard" && (
            <Button
              type="button"
              variant="ghost"
              disabled={clearKey.isPending}
              onClick={() => clearKey.mutate()}
            >
              Remove key
            </Button>
          )}
        </div>
      </form>

      {/* A test send is the only way to know the key and domain really work. */}
      <div className="space-y-3 rounded-xl border border-border bg-card p-5">
        <div className="flex items-center gap-2">
          <Mail className="h-4 w-4 text-muted-foreground" />
          <h2 className="text-sm font-semibold">Send a test</h2>
        </div>
        <p className="text-xs text-muted-foreground">
          Confirm delivery works before relying on invitations.
        </p>
        <div className="flex flex-wrap gap-2">
          <Input
            type="email"
            placeholder="you@example.com"
            value={testTo}
            onChange={(e) => setTestTo(e.target.value)}
            className="max-w-xs"
          />
          <Button
            type="button"
            variant="outline"
            disabled={!testTo || sendTest.isPending}
            onClick={() => sendTest.mutate()}
          >
            {sendTest.isPending ? (
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            ) : (
              <Send className="mr-2 h-3.5 w-3.5" />
            )}
            Send test
          </Button>
        </div>
      </div>
    </div>
  );
}

export default AdminEmailSettings;
