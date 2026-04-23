import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import {
  Loader2,
  Search,
  Cpu,
  MemoryStick,
  HardDrive,
  AppWindow,
  Database,
  Pencil,
  Gift,
  CreditCard,
  Building2,
  Sparkles,
  Info,
} from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { PageHelp } from "@/components/layout/PageHelp";
import { adminApi } from "@/api/admin";
import { orgsApi, type Organization } from "@/api/orgs";
import { cn } from "@/lib/utils";

const editSchema = z.object({
  custom_cpu_cores: z.number().int().min(1).max(256),
  custom_ram_mb: z.number().int().min(128),
  custom_disk_gb: z.number().int().min(1),
  custom_max_apps: z.number().int().min(1).max(1000),
  custom_max_databases: z.number().int().min(0).max(500),
  billing_type: z.enum(["free", "paid"]),
  price_amount: z.number().min(0).optional(),
  currency: z.enum(["USD", "EUR", "GBP", "UGX", "KES", "ZAR"]),
  billing_cycle: z.enum(["monthly", "yearly", "one_time"]),
});
type EditForm = z.infer<typeof editSchema>;

const PRESETS = [
  { name: "Free", cpu: 1, ram: 512, disk: 5, apps: 2, dbs: 1 },
  { name: "Starter", cpu: 2, ram: 2048, disk: 20, apps: 5, dbs: 3 },
  { name: "Pro", cpu: 4, ram: 8192, disk: 50, apps: 20, dbs: 10 },
  { name: "Enterprise", cpu: 16, ram: 32768, disk: 200, apps: 100, dbs: 50 },
];

export default function AdminOrgs() {
  const [query, setQuery] = useState("");
  const [editingOrg, setEditingOrg] = useState<Organization | null>(null);

  const { data, isLoading } = useQuery({
    queryKey: ["admin-orgs"],
    queryFn: () => adminApi.listAllOrgs(),
  });

  const orgs: Organization[] = data?.data?.data || [];
  const filtered = query
    ? orgs.filter(
        (o) =>
          o.name.toLowerCase().includes(query.toLowerCase()) ||
          o.slug.toLowerCase().includes(query.toLowerCase())
      )
    : orgs;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div>
        <div className="flex items-center gap-2">
          <h1 className="font-heading text-2xl font-semibold tracking-tight">
            Organizations
          </h1>
          <PageHelp
            title="Admin — Organizations"
            summary="Every tenant on your Orbita instance, their current quotas, and billing status."
            steps={[
              {
                title: "Find an org",
                body: "Use the filter to search by name or slug. The Custom badge means the org has resource overrides (not using a plan template).",
              },
              {
                title: "Edit resources",
                body: "Click Edit on any row. Change CPU, RAM, disk, app/DB limits, or flip billing between free and paid.",
              },
              {
                title: "Changes apply live",
                body: "If cgroup v2 enforcement is enabled on the host, new limits are written to /sys/fs/cgroup immediately — no restart needed.",
              },
            ]}
            nextLinks={[
              {
                label: "Create a new org",
                to: `/orgs/new`,
                description: "Onboard a new tenant with a custom quota + billing",
              },
              {
                label: "Nodes",
                to: `/admin/nodes`,
                description: "Scale capacity by adding more worker machines",
              },
            ]}
          />
        </div>
        <p className="mt-1 text-sm text-muted-foreground">
          All tenants on this Orbita instance. Edit resource quotas and billing
          for any org.
        </p>
      </div>

      {/* Stats */}
      {orgs.length > 0 && (
        <div className="grid gap-3 sm:grid-cols-3">
          <StatCard label="Total orgs" value={orgs.length} />
          <StatCard
            label="Paid tenants"
            value={orgs.filter((o) => o.billing_type === "paid").length}
            tone="brand"
          />
          <StatCard
            label="Free tenants"
            value={orgs.filter((o) => o.billing_type === "free").length}
            tone="success"
          />
        </div>
      )}

      {/* Search */}
      {orgs.length > 0 && (
        <div className="relative max-w-md">
          <Search className="pointer-events-none absolute left-3 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Filter by name or slug..."
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            className="pl-9"
          />
        </div>
      )}

      {/* Table */}
      {isLoading ? (
        <div className="flex justify-center py-16">
          <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
        </div>
      ) : orgs.length === 0 ? (
        <EmptyState />
      ) : (
        <div className="overflow-hidden rounded-xl border border-border bg-card shadow-xs">
          <div className="hidden grid-cols-[1fr_100px_100px_100px_110px_80px] gap-4 border-b border-border bg-muted/30 px-5 py-3 text-[11px] font-medium uppercase tracking-wider text-muted-foreground md:grid">
            <div>Organization</div>
            <div className="text-right">CPU</div>
            <div className="text-right">RAM</div>
            <div className="text-right">Disk</div>
            <div>Billing</div>
            <div></div>
          </div>
          <div className="divide-y divide-border">
            {filtered.map((org) => (
              <OrgRow
                key={org.id}
                org={org}
                onEdit={() => setEditingOrg(org)}
              />
            ))}
            {filtered.length === 0 && (
              <div className="px-5 py-10 text-center text-sm text-muted-foreground">
                No orgs match "{query}"
              </div>
            )}
          </div>
        </div>
      )}

      {/* Edit dialog */}
      <Dialog
        open={editingOrg !== null}
        onOpenChange={(open) => !open && setEditingOrg(null)}
      >
        {editingOrg && (
          <EditDialog
            org={editingOrg}
            onClose={() => setEditingOrg(null)}
          />
        )}
      </Dialog>
    </div>
  );
}

// ---------- row ----------

function OrgRow({ org, onEdit }: { org: Organization; onEdit: () => void }) {
  const effective = {
    cpu: org.custom_cpu_cores ?? org.plan?.max_cpu_cores ?? 1,
    ram: org.custom_ram_mb ?? org.plan?.max_ram_mb ?? 1024,
    disk: org.custom_disk_gb ?? org.plan?.max_disk_gb ?? 10,
  };
  const overriding =
    org.custom_cpu_cores !== null ||
    org.custom_ram_mb !== null ||
    org.custom_disk_gb !== null ||
    org.custom_max_apps !== null ||
    org.custom_max_databases !== null;

  return (
    <div className="grid grid-cols-1 items-center gap-3 px-5 py-4 transition-colors hover:bg-accent/30 md:grid-cols-[1fr_100px_100px_100px_110px_80px] md:gap-4">
      {/* Name */}
      <div className="flex min-w-0 items-center gap-3">
        <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-brand/10 text-xs font-semibold text-brand">
          {org.name.slice(0, 2).toUpperCase()}
        </div>
        <div className="min-w-0">
          <div className="flex items-center gap-2">
            <span className="truncate text-sm font-medium">{org.name}</span>
            {overriding && (
              <Badge variant="brand" className="h-4 px-1.5 text-[9px]">
                Custom
              </Badge>
            )}
          </div>
          <div className="font-mono text-[11px] text-muted-foreground">
            {org.slug}
          </div>
        </div>
      </div>

      {/* CPU */}
      <div className="text-right text-sm">
        <span className="font-medium">{effective.cpu}</span>{" "}
        <span className="text-muted-foreground">cores</span>
      </div>

      {/* RAM */}
      <div className="text-right text-sm">
        <span className="font-medium">{effective.ram}</span>{" "}
        <span className="text-muted-foreground">MB</span>
      </div>

      {/* Disk */}
      <div className="text-right text-sm">
        <span className="font-medium">{effective.disk}</span>{" "}
        <span className="text-muted-foreground">GB</span>
      </div>

      {/* Billing */}
      <div>
        {org.billing_type === "paid" && org.price_monthly_cents ? (
          <Badge variant="brand">
            <CreditCard className="h-3 w-3" />
            {formatPrice(org.price_monthly_cents, org.currency, org.billing_cycle)}
          </Badge>
        ) : (
          <Badge variant="success">
            <Gift className="h-3 w-3" />
            Free
          </Badge>
        )}
      </div>

      {/* Actions */}
      <div className="flex justify-end">
        <Button variant="outline" size="sm" onClick={onEdit}>
          <Pencil className="h-3 w-3" />
          Edit
        </Button>
      </div>
    </div>
  );
}

// ---------- dialog ----------

function EditDialog({
  org,
  onClose,
}: {
  org: Organization;
  onClose: () => void;
}) {
  const qc = useQueryClient();

  const {
    register,
    handleSubmit,
    setValue,
    watch,
    formState: { errors },
  } = useForm<EditForm>({
    resolver: zodResolver(editSchema),
    defaultValues: {
      custom_cpu_cores:
        org.custom_cpu_cores ?? org.plan?.max_cpu_cores ?? 2,
      custom_ram_mb: org.custom_ram_mb ?? org.plan?.max_ram_mb ?? 2048,
      custom_disk_gb: org.custom_disk_gb ?? org.plan?.max_disk_gb ?? 20,
      custom_max_apps: org.custom_max_apps ?? org.plan?.max_apps ?? 5,
      custom_max_databases:
        org.custom_max_databases ?? org.plan?.max_databases ?? 3,
      billing_type: org.billing_type,
      price_amount:
        org.price_monthly_cents != null ? org.price_monthly_cents / 100 : 0,
      currency: (org.currency as EditForm["currency"]) || "USD",
      billing_cycle: org.billing_cycle,
    },
  });

  const billingType = watch("billing_type");

  const applyPreset = (p: (typeof PRESETS)[number]) => {
    setValue("custom_cpu_cores", p.cpu);
    setValue("custom_ram_mb", p.ram);
    setValue("custom_disk_gb", p.disk);
    setValue("custom_max_apps", p.apps);
    setValue("custom_max_databases", p.dbs);
  };

  const mutation = useMutation({
    mutationFn: (data: EditForm) =>
      orgsApi.updateResources(org.slug, {
        custom_cpu_cores: data.custom_cpu_cores,
        custom_ram_mb: data.custom_ram_mb,
        custom_disk_gb: data.custom_disk_gb,
        custom_max_apps: data.custom_max_apps,
        custom_max_databases: data.custom_max_databases,
        billing_type: data.billing_type,
        currency: data.currency,
        billing_cycle: data.billing_cycle,
        price_monthly_cents:
          data.billing_type === "paid" && data.price_amount
            ? Math.round(data.price_amount * 100)
            : 0,
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["admin-orgs"] });
      qc.invalidateQueries({ queryKey: ["org", org.slug] });
      toast.success("Resources updated");
      onClose();
    },
    onError: (err: unknown) => {
      const msg =
        (err as { response?: { data?: { error?: { message?: string } } } })
          ?.response?.data?.error?.message || "Update failed";
      toast.error(msg);
    },
  });

  return (
    <DialogContent className="max-w-2xl">
      <DialogHeader>
        <DialogTitle className="flex items-center gap-2">
          <Building2 className="h-4 w-4 text-brand" />
          {org.name}
        </DialogTitle>
        <DialogDescription>
          Edit resource quota and billing for{" "}
          <code className="font-mono text-xs">{org.slug}</code>. Changes take
          effect immediately.
        </DialogDescription>
      </DialogHeader>

      <form
        onSubmit={handleSubmit((d) => mutation.mutate(d))}
        className="space-y-5"
      >
        {/* Resource presets */}
        <div>
          <div className="mb-2 text-[11px] font-medium uppercase tracking-wider text-muted-foreground">
            Quick presets
          </div>
          <div className="flex flex-wrap gap-2">
            {PRESETS.map((p) => (
              <button
                key={p.name}
                type="button"
                onClick={() => applyPreset(p)}
                className="inline-flex items-center gap-1.5 rounded-md border border-border bg-background px-2.5 py-1 text-xs font-medium transition-colors hover:border-brand/40 hover:bg-accent"
              >
                <Sparkles className="h-3 w-3 text-brand" />
                {p.name}
              </button>
            ))}
          </div>
        </div>

        {/* Resource fields */}
        <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
          <NumField
            icon={Cpu}
            label="CPU cores"
            unit="cores"
            {...register("custom_cpu_cores", { valueAsNumber: true })}
            error={errors.custom_cpu_cores?.message}
          />
          <NumField
            icon={MemoryStick}
            label="RAM"
            unit="MB"
            {...register("custom_ram_mb", { valueAsNumber: true })}
            error={errors.custom_ram_mb?.message}
          />
          <NumField
            icon={HardDrive}
            label="Disk"
            unit="GB"
            {...register("custom_disk_gb", { valueAsNumber: true })}
            error={errors.custom_disk_gb?.message}
          />
          <NumField
            icon={AppWindow}
            label="Max apps"
            unit="apps"
            {...register("custom_max_apps", { valueAsNumber: true })}
            error={errors.custom_max_apps?.message}
          />
          <NumField
            icon={Database}
            label="Max DBs"
            unit="max"
            {...register("custom_max_databases", { valueAsNumber: true })}
            error={errors.custom_max_databases?.message}
          />
        </div>

        {/* Billing */}
        <div className="border-t border-border pt-4">
          <div className="mb-2 text-[11px] font-medium uppercase tracking-wider text-muted-foreground">
            Billing
          </div>
          <div className="grid grid-cols-2 gap-2">
            <BillingOption
              active={billingType === "free"}
              icon={Gift}
              label="Free"
              onClick={() => setValue("billing_type", "free")}
            />
            <BillingOption
              active={billingType === "paid"}
              icon={CreditCard}
              label="Paid"
              onClick={() => setValue("billing_type", "paid")}
            />
          </div>
          <input type="hidden" {...register("billing_type")} />

          {billingType === "paid" && (
            <div className="mt-3 grid grid-cols-3 gap-3">
              <div className="space-y-1">
                <Label htmlFor="price_amount" className="text-xs">Price</Label>
                <Input
                  id="price_amount"
                  type="number"
                  step="0.01"
                  min="0"
                  {...register("price_amount", { valueAsNumber: true })}
                />
              </div>
              <div className="space-y-1">
                <Label htmlFor="currency" className="text-xs">Currency</Label>
                <select
                  id="currency"
                  {...register("currency")}
                  className="flex h-10 w-full rounded-lg border border-input bg-background/50 px-3 text-sm outline-none focus-visible:border-ring focus-visible:ring-2 focus-visible:ring-ring/30 dark:bg-input/40"
                >
                  <option value="USD">USD</option>
                  <option value="EUR">EUR</option>
                  <option value="GBP">GBP</option>
                  <option value="UGX">UGX</option>
                  <option value="KES">KES</option>
                  <option value="ZAR">ZAR</option>
                </select>
              </div>
              <div className="space-y-1">
                <Label htmlFor="billing_cycle" className="text-xs">Cycle</Label>
                <select
                  id="billing_cycle"
                  {...register("billing_cycle")}
                  className="flex h-10 w-full rounded-lg border border-input bg-background/50 px-3 text-sm outline-none focus-visible:border-ring focus-visible:ring-2 focus-visible:ring-ring/30 dark:bg-input/40"
                >
                  <option value="monthly">Monthly</option>
                  <option value="yearly">Yearly</option>
                  <option value="one_time">One-time</option>
                </select>
              </div>
            </div>
          )}
        </div>

        <div className="flex items-start gap-2 rounded-md border border-border bg-muted/40 px-3 py-2 text-xs text-muted-foreground">
          <Info className="mt-0.5 h-3.5 w-3.5 shrink-0 text-brand" />
          <span>
            Saving writes to <code className="font-mono">PUT /admin/orgs/{org.slug}/resources</code>.
            If cgroup v2 enforcement is enabled on the host, new limits are applied
            to the org's cgroup slice immediately.
          </span>
        </div>

        <div className="flex justify-end gap-2 pt-1">
          <Button type="button" variant="ghost" onClick={onClose}>
            Cancel
          </Button>
          <Button type="submit" variant="brand" disabled={mutation.isPending}>
            {mutation.isPending && (
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            )}
            Save changes
          </Button>
        </div>
      </form>
    </DialogContent>
  );
}

// ---------- helpers ----------

function EmptyState() {
  return (
    <div className="flex flex-col items-center gap-3 rounded-xl border border-dashed border-border bg-card/50 px-6 py-16 text-center">
      <Building2 className="h-8 w-8 text-muted-foreground" />
      <p className="text-sm text-muted-foreground">No organizations yet.</p>
    </div>
  );
}

function StatCard({
  label,
  value,
  tone,
}: {
  label: string;
  value: number;
  tone?: "brand" | "success";
}) {
  const toneCls = {
    brand: "bg-brand/10 text-brand",
    success: "bg-success/10 text-success",
  };
  return (
    <div className="rounded-xl border border-border bg-card p-4 shadow-xs">
      <div className="text-[11px] font-medium uppercase tracking-wider text-muted-foreground">
        {label}
      </div>
      <div className="mt-2 flex items-baseline gap-2">
        <span className="font-heading text-2xl font-semibold tracking-tight">
          {value}
        </span>
        {tone && (
          <span className={cn("h-1.5 w-1.5 rounded-full", toneCls[tone])}>
            &nbsp;
          </span>
        )}
      </div>
    </div>
  );
}

function formatPrice(cents: number, currency: string, cycle: string) {
  const amount = cents / 100;
  const cycleLabel = cycle === "yearly" ? "/yr" : cycle === "one_time" ? "" : "/mo";
  try {
    return `${new Intl.NumberFormat("en-US", {
      style: "currency",
      currency,
      minimumFractionDigits: 0,
      maximumFractionDigits: 2,
    }).format(amount)}${cycleLabel}`;
  } catch {
    return `${amount} ${currency}${cycleLabel}`;
  }
}

function BillingOption({
  active,
  icon: Icon,
  label,
  onClick,
}: {
  active: boolean;
  icon: typeof Gift;
  label: string;
  onClick: () => void;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        "flex items-center gap-2 rounded-lg border px-3 py-2.5 text-left transition-colors",
        active
          ? "border-brand bg-brand/5 ring-1 ring-brand/20"
          : "border-border bg-background hover:border-foreground/20"
      )}
    >
      <div
        className={cn(
          "flex h-7 w-7 shrink-0 items-center justify-center rounded-md",
          active ? "bg-brand/15 text-brand" : "bg-muted text-muted-foreground"
        )}
      >
        <Icon className="h-3.5 w-3.5" />
      </div>
      <span className="text-sm font-medium">{label}</span>
    </button>
  );
}

const NumField = ({
  icon: Icon,
  label,
  unit,
  error,
  ...rest
}: {
  icon: typeof Cpu;
  label: string;
  unit: string;
  error?: string;
} & React.InputHTMLAttributes<HTMLInputElement>) => (
  <div className="space-y-1">
    <Label className="text-xs">
      <span className="flex items-center gap-1.5">
        <Icon className="h-3 w-3 text-muted-foreground" />
        {label}
      </span>
    </Label>
    <div className="relative">
      <Input type="number" className="h-9 pr-14 text-sm" {...rest} />
      <span className="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-[10px] text-muted-foreground">
        {unit}
      </span>
    </div>
    {error && <p className="text-[11px] text-destructive">{error}</p>}
  </div>
);
