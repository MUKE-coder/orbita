import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { useNavigate } from "react-router-dom";
import { useMutation, useQuery } from "@tanstack/react-query";
import { Loader2, Copy, Check, ArrowLeft, RefreshCw, ShieldCheck } from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { adminApi, type ProvisionResult } from "@/api/admin";

const schema = z.object({
  name: z.string().min(2, "Organisation name is required"),
  slug: z.string(),
  plan_id: z.string(),
  admin_email: z.string().email("Enter a valid email"),
  admin_name: z.string(),
  admin_password: z.string(),
  admin_role: z.enum(["owner", "admin"]),
  // Blank inherits the plan's value.
  custom_cpu_cores: z.string(),
  custom_ram_mb: z.string(),
  custom_disk_gb: z.string(),
  custom_max_apps: z.string(),
  custom_max_databases: z.string(),
});

type Form = z.infer<typeof schema>;

// Blank means "inherit from the plan", so only send the ones actually filled in.
const num = (v: string): number | undefined => {
  const t = v.trim();
  if (t === "") return undefined;
  const n = Number(t);
  return Number.isFinite(n) && n > 0 ? n : undefined;
};

function generatePassword(len = 16) {
  const alphabet = "abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789@#%+=?";
  const bytes = new Uint32Array(len);
  crypto.getRandomValues(bytes);
  return Array.from(bytes, (b) => alphabet[b % alphabet.length]).join("");
}

/**
 * Onboards a tenant in one step: the organisation, its resource limits, and the
 * account its admin signs in with.
 *
 * This is the path for instances without email — the credentials are shown once
 * for the operator to pass on, and the account must replace the password at
 * first sign-in.
 */
function AdminNewTenant() {
  const navigate = useNavigate();
  const [result, setResult] = useState<ProvisionResult | null>(null);
  const [copied, setCopied] = useState<string | null>(null);

  const { data: plansData } = useQuery({
    queryKey: ["admin-plans"],
    queryFn: () => adminApi.listPlans(),
  });
  const plans = plansData?.data?.data || [];

  const {
    register,
    handleSubmit,
    setValue,
    watch,
    formState: { errors },
  } = useForm<Form>({
    resolver: zodResolver(schema),
    defaultValues: {
      name: "",
      slug: "",
      plan_id: "",
      admin_email: "",
      admin_name: "",
      admin_password: generatePassword(),
      admin_role: "owner",
      custom_cpu_cores: "",
      custom_ram_mb: "",
      custom_disk_gb: "",
      custom_max_apps: "",
      custom_max_databases: "",
    },
  });

  const provision = useMutation({
    mutationFn: (d: Form) =>
      adminApi.provisionOrg({
        name: d.name,
        slug: d.slug.trim() || undefined,
        plan_id: d.plan_id || undefined,
        custom_cpu_cores: num(d.custom_cpu_cores),
        custom_ram_mb: num(d.custom_ram_mb),
        custom_disk_gb: num(d.custom_disk_gb),
        custom_max_apps: num(d.custom_max_apps),
        custom_max_databases: num(d.custom_max_databases),
        admin_email: d.admin_email,
        admin_name: d.admin_name || undefined,
        admin_password: d.admin_password || undefined,
        admin_role: d.admin_role,
      }),
    onSuccess: (res) => setResult(res.data.data),
    onError: (err: any) =>
      toast.error(err?.response?.data?.error?.message || "Could not create the tenant"),
  });

  const copy = async (label: string, value: string) => {
    try {
      await navigator.clipboard.writeText(value);
      setCopied(label);
      setTimeout(() => setCopied(null), 1500);
    } catch {
      toast.error("Copy failed — select the text manually");
    }
  };

  // Credential handover. Shown once: the password is not stored in plaintext
  // and cannot be retrieved again.
  if (result) {
    const loginURL = `${window.location.origin}/login`;
    const handover = `Orbita access for ${result.organization.name}\nURL: ${loginURL}\nEmail: ${result.user.email}\nPassword: ${result.password}`;

    return (
      <div className="mx-auto max-w-xl space-y-5 p-6">
        <div className="flex items-center gap-3">
          <div className="flex h-10 w-10 items-center justify-center rounded-xl border border-green-500/20 bg-green-500/10 text-green-500">
            <ShieldCheck className="h-5 w-5" />
          </div>
          <div>
            <h1 className="font-heading text-lg font-semibold">{result.organization.name} is ready</h1>
            <p className="text-sm text-muted-foreground">
              Send these details to the org admin.
            </p>
          </div>
        </div>

        <div className="space-y-3 rounded-xl border border-border bg-card p-5">
          {[
            { label: "Sign-in URL", value: loginURL },
            { label: "Email", value: result.user.email },
            { label: "Password", value: result.password },
          ].map((row) => (
            <div key={row.label} className="flex items-center justify-between gap-3">
              <div className="min-w-0">
                <p className="text-[11px] uppercase tracking-wide text-muted-foreground">{row.label}</p>
                <p className="truncate font-mono text-sm">{row.value}</p>
              </div>
              <Button type="button" size="sm" variant="ghost" onClick={() => copy(row.label, row.value)}>
                {copied === row.label ? <Check className="h-3.5 w-3.5 text-green-500" /> : <Copy className="h-3.5 w-3.5" />}
              </Button>
            </div>
          ))}
        </div>

        <div className="rounded-lg border border-amber-500/25 bg-amber-500/5 p-3 text-xs text-muted-foreground">
          The password is shown <strong>once</strong> — it isn't stored and can't be
          retrieved later. {result.user.email} will be asked to set their own
          password the first time they sign in.
        </div>

        <div className="flex flex-wrap gap-2">
          <Button variant="brand" onClick={() => copy("All", handover)}>
            {copied === "All" ? <Check className="mr-2 h-3.5 w-3.5" /> : <Copy className="mr-2 h-3.5 w-3.5" />}
            Copy all details
          </Button>
          <Button variant="outline" onClick={() => navigate("/admin/orgs")}>
            Done
          </Button>
        </div>
      </div>
    );
  }

  return (
    <div className="mx-auto max-w-2xl space-y-6 p-6">
      <button
        type="button"
        onClick={() => navigate("/admin/orgs")}
        className="inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground"
      >
        <ArrowLeft className="h-3.5 w-3.5" /> Organisations
      </button>

      <div>
        <h1 className="font-heading text-xl font-semibold tracking-tight">New tenant</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Creates the organisation and the account its admin signs in with.
        </p>
      </div>

      <form onSubmit={handleSubmit((d) => provision.mutate(d))} className="space-y-5">
        <section className="space-y-4 rounded-xl border border-border bg-card p-5">
          <h2 className="text-sm font-semibold">Organisation</h2>
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-1.5">
              <Label htmlFor="name">Name</Label>
              <Input id="name" placeholder="Acme Ltd" {...register("name")} />
              {errors.name && <p className="text-xs text-destructive">{errors.name.message}</p>}
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="slug">Slug</Label>
              <Input id="slug" placeholder="(from name)" className="font-mono" {...register("slug")} />
            </div>
          </div>

          <div className="space-y-1.5">
            <Label htmlFor="plan_id">Plan</Label>
            <select
              id="plan_id"
              className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm"
              {...register("plan_id")}
            >
              <option value="">Default (Free)</option>
              {plans.map((p) => (
                <option key={p.id} value={p.id}>
                  {p.name} — {p.max_cpu_cores} CPU, {p.max_ram_mb}MB, {p.max_apps} apps
                </option>
              ))}
            </select>
            <p className="text-[11px] text-muted-foreground">
              The plan sets the baseline. Override individual limits below if this
              tenant needs something different.
            </p>
          </div>
        </section>

        <section className="space-y-4 rounded-xl border border-border bg-card p-5">
          <div>
            <h2 className="text-sm font-semibold">Resource overrides</h2>
            <p className="text-[11px] text-muted-foreground">
              Leave blank to inherit the plan's value.
            </p>
          </div>
          <div className="grid gap-3 sm:grid-cols-3">
            <div className="space-y-1.5">
              <Label htmlFor="custom_cpu_cores">CPU cores</Label>
              <Input id="custom_cpu_cores" inputMode="numeric" placeholder="—" {...register("custom_cpu_cores")} />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="custom_ram_mb">RAM (MB)</Label>
              <Input id="custom_ram_mb" inputMode="numeric" placeholder="—" {...register("custom_ram_mb")} />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="custom_disk_gb">Disk (GB)</Label>
              <Input id="custom_disk_gb" inputMode="numeric" placeholder="—" {...register("custom_disk_gb")} />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="custom_max_apps">Max apps</Label>
              <Input id="custom_max_apps" inputMode="numeric" placeholder="—" {...register("custom_max_apps")} />
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="custom_max_databases">Max databases</Label>
              <Input id="custom_max_databases" inputMode="numeric" placeholder="—" {...register("custom_max_databases")} />
            </div>
          </div>
        </section>

        <section className="space-y-4 rounded-xl border border-border bg-card p-5">
          <h2 className="text-sm font-semibold">Org admin</h2>
          <div className="grid gap-4 sm:grid-cols-2">
            <div className="space-y-1.5">
              <Label htmlFor="admin_email">Email</Label>
              <Input id="admin_email" type="email" placeholder="jane@acme.com" {...register("admin_email")} />
              {errors.admin_email && (
                <p className="text-xs text-destructive">{errors.admin_email.message}</p>
              )}
            </div>
            <div className="space-y-1.5">
              <Label htmlFor="admin_name">Name</Label>
              <Input id="admin_name" placeholder="Jane Doe" {...register("admin_name")} />
            </div>
          </div>

          <div className="space-y-1.5">
            <Label htmlFor="admin_password">Initial password</Label>
            <div className="flex gap-2">
              <Input id="admin_password" className="font-mono" {...register("admin_password")} />
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={() => setValue("admin_password", generatePassword())}
                title="Generate a new password"
              >
                <RefreshCw className="h-3.5 w-3.5" />
              </Button>
            </div>
            <p className="text-[11px] text-muted-foreground">
              Generated for you — edit it if you prefer. They'll be required to set
              their own at first sign-in.
            </p>
          </div>

          <div className="space-y-1.5">
            <Label htmlFor="admin_role">Role</Label>
            <select
              id="admin_role"
              className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm"
              {...register("admin_role")}
            >
              <option value="owner">Owner — full control of their org</option>
              <option value="admin">Admin — manage the org, but not delete it</option>
            </select>
          </div>
        </section>

        <Button type="submit" variant="brand" disabled={provision.isPending} className="w-full sm:w-auto">
          {provision.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
          Create org + admin
        </Button>
        {watch("admin_email") && (
          <p className="text-[11px] text-muted-foreground">
            Credentials will be shown once after creation.
          </p>
        )}
      </form>
    </div>
  );
}

export default AdminNewTenant;
