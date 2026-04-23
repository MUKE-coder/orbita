import { useState } from "react";
import { useForm, useWatch } from "react-hook-form";
import { useQuery } from "@tanstack/react-query";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { useNavigate } from "react-router-dom";
import {
  Loader2,
  Cpu,
  MemoryStick,
  HardDrive,
  AppWindow,
  Database,
  ArrowLeft,
  Sparkles,
  CreditCard,
  Gift,
  Info,
  Server,
  AlertTriangle,
} from "lucide-react";
import { toast } from "sonner";

import Logo from "@/components/layout/Logo";
import ThemeToggle from "@/components/layout/ThemeToggle";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { PageHelp } from "@/components/layout/PageHelp";
import { orgsApi, type CreateOrgInput } from "@/api/orgs";
import { adminApi } from "@/api/admin";
import { cn } from "@/lib/utils";

// ---------- schema ----------

const schema = z.object({
  name: z.string().min(2, "Name must be at least 2 characters"),
  slug: z
    .string()
    .min(2, "Slug must be at least 2 characters")
    .regex(/^[a-z0-9-]+$/, "Lowercase letters, numbers, and hyphens only"),
  description: z.string().optional(),

  // Resource quotas (registered with valueAsNumber so strings from <input type=number> coerce to number)
  custom_cpu_cores: z.number().int().min(1).max(256),
  custom_ram_mb: z.number().int().min(128),
  custom_disk_gb: z.number().int().min(1),
  custom_max_apps: z.number().int().min(1).max(1000),
  custom_max_databases: z.number().int().min(0).max(500),

  // Billing
  billing_type: z.enum(["free", "paid"]),
  price_amount: z.number().min(0).optional(),
  currency: z.enum(["USD", "EUR", "GBP", "UGX", "KES", "ZAR"]),
  billing_cycle: z.enum(["monthly", "yearly", "one_time"]),
});
type FormValues = z.infer<typeof schema>;

const PRESETS = [
  { name: "Free", cpu: 1, ram: 512, disk: 5, apps: 2, dbs: 1 },
  { name: "Starter", cpu: 2, ram: 2048, disk: 20, apps: 5, dbs: 3 },
  { name: "Pro", cpu: 4, ram: 8192, disk: 50, apps: 20, dbs: 10 },
  { name: "Enterprise", cpu: 16, ram: 32768, disk: 200, apps: 100, dbs: 50 },
];

function CreateOrg() {
  const navigate = useNavigate();
  const [isLoading, setIsLoading] = useState(false);

  // Host capacity — fetched for super admin only; silently falls back if 403
  const { data: capacityRes } = useQuery({
    queryKey: ["platform-capacity"],
    queryFn: () => adminApi.getPlatformCapacity(),
    retry: false,
  });
  const capacity = capacityRes?.data?.data;

  const {
    register,
    handleSubmit,
    setValue,
    control,
    formState: { errors },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      custom_cpu_cores: 2,
      custom_ram_mb: 2048,
      custom_disk_gb: 20,
      custom_max_apps: 5,
      custom_max_databases: 3,
      billing_type: "free",
      price_amount: 0,
      currency: "USD",
      billing_cycle: "monthly",
    },
  });

  const billingType = useWatch({ control, name: "billing_type" });
  const watchedCPU = useWatch({ control, name: "custom_cpu_cores" });
  const watchedRAM = useWatch({ control, name: "custom_ram_mb" });
  const watchedDisk = useWatch({ control, name: "custom_disk_gb" });

  const overAllocation = capacity
    ? {
        cpu: (watchedCPU ?? 0) > capacity.available.cpu_cores,
        ram: (watchedRAM ?? 0) > capacity.available.ram_mb,
        disk: (watchedDisk ?? 0) > capacity.available.disk_gb,
      }
    : null;

  const handleNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const slug = e.target.value
      .toLowerCase()
      .replace(/[^a-z0-9]+/g, "-")
      .replace(/^-|-$/g, "");
    setValue("slug", slug);
  };

  const applyPreset = (preset: (typeof PRESETS)[number]) => {
    setValue("custom_cpu_cores", preset.cpu);
    setValue("custom_ram_mb", preset.ram);
    setValue("custom_disk_gb", preset.disk);
    setValue("custom_max_apps", preset.apps);
    setValue("custom_max_databases", preset.dbs);
    toast.success(`Applied ${preset.name} preset`);
  };

  const onSubmit = async (data: FormValues) => {
    setIsLoading(true);
    try {
      const payload: CreateOrgInput = {
        name: data.name,
        slug: data.slug,
        description: data.description || undefined,
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
            : undefined,
      };
      await orgsApi.create(payload);
      toast.success("Organization created");
      navigate("/dashboard");
    } catch (err: unknown) {
      const message =
        (err as { response?: { data?: { error?: { message?: string } } } })
          ?.response?.data?.error?.message || "Failed to create organization";
      toast.error(message);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="relative min-h-screen bg-background text-foreground">
      <div className="pointer-events-none absolute inset-0 bg-grid opacity-40" aria-hidden />
      <div className="pointer-events-none absolute inset-0 bg-radial-glow" aria-hidden />

      {/* Top bar */}
      <header className="relative z-10 flex h-14 items-center justify-between border-b border-border bg-background/80 px-6 backdrop-blur">
        <Logo size="md" to="/dashboard" />
        <div className="flex items-center gap-1">
          <button
            onClick={() => navigate("/dashboard")}
            className="inline-flex h-8 items-center gap-1.5 rounded-md px-2.5 text-xs font-medium text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
          >
            <ArrowLeft className="h-3.5 w-3.5" />
            Back to dashboard
          </button>
          <ThemeToggle />
        </div>
      </header>

      <main className="relative z-10 mx-auto w-full max-w-3xl px-6 py-12">
        <div className="mb-8 text-center">
          <div className="mx-auto mb-4 flex h-12 w-12 items-center justify-center rounded-xl border border-border bg-card shadow-sm">
            <Sparkles className="h-5 w-5 text-brand" />
          </div>
          <div className="flex items-center justify-center gap-2">
            <h1 className="font-heading text-3xl font-semibold tracking-tight">
              Create an organization
            </h1>
            <PageHelp
              title="Create organization"
              summary="A new tenant on your Orbita instance. Each org is fully isolated — own Docker network, cgroup slice, encryption key."
              steps={[
                {
                  title: "Review host capacity",
                  body: "The top panel shows your VPS total + what's already allocated to other orgs. Don't over-commit.",
                },
                {
                  title: "Pick a preset or set exact values",
                  body: "Free/Starter/Pro/Enterprise presets fill the fields. Override any number if you want custom.",
                },
                {
                  title: "Decide billing",
                  body: "Free for internal/complimentary tenants. Paid stores the price — you run invoicing yourself (Stripe integration coming later).",
                },
                {
                  title: "Create",
                  body: "The org gets a Docker network and (if cgroup v2 is live) a kernel-enforced resource slice.",
                },
              ]}
              nextLinks={[
                {
                  label: "Admin → Organizations",
                  to: `/admin/orgs`,
                  description: "Manage all orgs + edit quotas after creation",
                },
              ]}
            />
          </div>
          <p className="mt-2 text-sm text-muted-foreground">
            Set the resource quota and billing for this tenant. You can adjust
            these anytime from the admin area.
          </p>
        </div>

        <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
          {/* --- Basic info --- */}
          <Section title="Basics" description="Identify this organization.">
            <div className="space-y-1.5">
              <Label htmlFor="name">Organization name</Label>
              <Input
                id="name"
                placeholder="Limibooks"
                autoFocus
                {...register("name", { onChange: handleNameChange })}
              />
              {errors.name && (
                <p className="text-xs text-destructive">{errors.name.message}</p>
              )}
            </div>

            <div className="space-y-1.5">
              <Label htmlFor="slug">Slug</Label>
              <Input
                id="slug"
                placeholder="limibooks"
                className="font-mono"
                {...register("slug")}
              />
              {errors.slug && (
                <p className="text-xs text-destructive">{errors.slug.message}</p>
              )}
              <p className="text-xs text-muted-foreground">
                Used in URLs and Docker network names.
              </p>
            </div>

            <div className="space-y-1.5">
              <Label htmlFor="description">Description</Label>
              <Input
                id="description"
                placeholder="Optional — a one-liner about this tenant"
                {...register("description")}
              />
            </div>
          </Section>

          {/* --- Host capacity summary --- */}
          {capacity && <CapacityPanel capacity={capacity} />}

          {/* --- Resource quota --- */}
          <Section
            title="Resource quota"
            description="Set the hard limits this organization can consume on the host. Start with a preset or enter exact values."
          >
            {/* Presets */}
            <div className="flex flex-wrap gap-2">
              {PRESETS.map((p) => (
                <button
                  key={p.name}
                  type="button"
                  onClick={() => applyPreset(p)}
                  className="inline-flex items-center gap-1.5 rounded-md border border-border bg-background px-2.5 py-1 text-xs font-medium text-foreground transition-colors hover:border-brand/40 hover:bg-accent"
                >
                  <Sparkles className="h-3 w-3 text-brand" />
                  {p.name}
                </button>
              ))}
            </div>

            <div className="grid gap-4 sm:grid-cols-2">
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
                label="Max databases"
                unit="dbs"
                {...register("custom_max_databases", { valueAsNumber: true })}
                error={errors.custom_max_databases?.message}
              />
            </div>

            {overAllocation && (overAllocation.cpu || overAllocation.ram || overAllocation.disk) && (
              <div className="flex items-start gap-2 rounded-md border border-destructive/30 bg-destructive/5 px-3 py-2.5 text-xs text-destructive">
                <AlertTriangle className="mt-0.5 h-3.5 w-3.5 shrink-0" />
                <div>
                  <div className="font-medium">Over-allocating host capacity</div>
                  <div className="mt-0.5 text-[11px] opacity-90">
                    {overAllocation.cpu && "CPU "}
                    {overAllocation.ram && "RAM "}
                    {overAllocation.disk && "Disk "}
                    exceed what's available. Other orgs may experience contention.
                  </div>
                </div>
              </div>
            )}
          </Section>

          {/* --- Billing --- */}
          <Section
            title="Billing"
            description="Mark this tenant as free or set a monthly/yearly price."
          >
            <div className="grid grid-cols-2 gap-3">
              <BillingOption
                active={billingType === "free"}
                icon={Gift}
                label="Free"
                desc="Internal or complimentary tenant"
                onClick={() => setValue("billing_type", "free")}
              />
              <BillingOption
                active={billingType === "paid"}
                icon={CreditCard}
                label="Paid"
                desc="Charge this client a recurring amount"
                onClick={() => setValue("billing_type", "paid")}
              />
            </div>
            <input type="hidden" {...register("billing_type")} />

            {billingType === "paid" && (
              <div className="grid gap-4 sm:grid-cols-3">
                <div className="space-y-1.5">
                  <Label htmlFor="price_amount">Price</Label>
                  <Input
                    id="price_amount"
                    type="number"
                    step="0.01"
                    min="0"
                    placeholder="29.99"
                    {...register("price_amount", { valueAsNumber: true })}
                  />
                  {errors.price_amount && (
                    <p className="text-xs text-destructive">
                      {errors.price_amount.message}
                    </p>
                  )}
                </div>
                <div className="space-y-1.5">
                  <Label htmlFor="currency">Currency</Label>
                  <select
                    id="currency"
                    {...register("currency")}
                    className="flex h-10 w-full rounded-lg border border-input bg-background/50 px-3 text-sm outline-none transition-colors focus-visible:border-ring focus-visible:ring-2 focus-visible:ring-ring/30 dark:bg-input/40"
                  >
                    <option value="USD">USD — US Dollar</option>
                    <option value="EUR">EUR — Euro</option>
                    <option value="GBP">GBP — British Pound</option>
                    <option value="UGX">UGX — Uganda Shilling</option>
                    <option value="KES">KES — Kenya Shilling</option>
                    <option value="ZAR">ZAR — S. African Rand</option>
                  </select>
                </div>
                <div className="space-y-1.5">
                  <Label htmlFor="billing_cycle">Cycle</Label>
                  <select
                    id="billing_cycle"
                    {...register("billing_cycle")}
                    className="flex h-10 w-full rounded-lg border border-input bg-background/50 px-3 text-sm outline-none transition-colors focus-visible:border-ring focus-visible:ring-2 focus-visible:ring-ring/30 dark:bg-input/40"
                  >
                    <option value="monthly">Monthly</option>
                    <option value="yearly">Yearly</option>
                    <option value="one_time">One-time</option>
                  </select>
                </div>
              </div>
            )}

            {billingType === "paid" && (
              <div className="flex items-start gap-2 rounded-md border border-border bg-muted/40 px-3 py-2.5 text-xs text-muted-foreground">
                <Info className="mt-0.5 h-3.5 w-3.5 shrink-0 text-brand" />
                <span>
                  Orbita stores the price — you run the invoicing yourself. A
                  Stripe integration for automated charging is on the roadmap.
                </span>
              </div>
            )}
          </Section>

          <div className="flex items-center justify-end gap-2 pt-2">
            <Button
              type="button"
              variant="ghost"
              onClick={() => navigate("/dashboard")}
              disabled={isLoading}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              variant="brand"
              size="lg"
              disabled={isLoading}
            >
              {isLoading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Create organization
            </Button>
          </div>
        </form>
      </main>
    </div>
  );
}

// ---------- subcomponents ----------

function Section({
  title,
  description,
  children,
}: {
  title: string;
  description?: string;
  children: React.ReactNode;
}) {
  return (
    <section className="overflow-hidden rounded-xl border border-border bg-card shadow-xs">
      <header className="border-b border-border bg-muted/30 px-6 py-4">
        <h2 className="font-heading text-[15px] font-semibold tracking-tight">
          {title}
        </h2>
        {description && (
          <p className="mt-0.5 text-xs text-muted-foreground">{description}</p>
        )}
      </header>
      <div className="space-y-4 px-6 py-5">{children}</div>
    </section>
  );
}

function BillingOption({
  active,
  icon: Icon,
  label,
  desc,
  onClick,
}: {
  active: boolean;
  icon: typeof Gift;
  label: string;
  desc: string;
  onClick: () => void;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        "flex items-start gap-3 rounded-lg border p-4 text-left transition-colors",
        active
          ? "border-brand bg-brand/5 ring-1 ring-brand/20"
          : "border-border bg-background hover:border-foreground/20 hover:bg-accent/50"
      )}
    >
      <div
        className={cn(
          "flex h-9 w-9 shrink-0 items-center justify-center rounded-md",
          active ? "bg-brand/15 text-brand" : "bg-muted text-muted-foreground"
        )}
      >
        <Icon className="h-4 w-4" />
      </div>
      <div className="min-w-0">
        <div className="text-sm font-semibold">{label}</div>
        <div className="mt-0.5 text-xs text-muted-foreground">{desc}</div>
      </div>
    </button>
  );
}

function CapacityPanel({
  capacity,
}: {
  capacity: {
    host: { cpu_cores: number; ram_mb: number; disk_gb: number };
    allocated: { cpu_cores: number; ram_mb: number; disk_gb: number };
    available: { cpu_cores: number; ram_mb: number; disk_gb: number };
    orgs: Array<{ slug: string; name: string; cpu_cores: number; ram_mb: number; disk_gb: number }>;
  };
}) {
  const { host, allocated, available, orgs } = capacity;

  return (
    <section className="overflow-hidden rounded-xl border border-border bg-card shadow-xs">
      <header className="flex items-center gap-2 border-b border-border bg-muted/30 px-6 py-4">
        <Server className="h-4 w-4 text-brand" />
        <h2 className="font-heading text-[15px] font-semibold tracking-tight">
          Host capacity
        </h2>
        <span className="ml-auto text-[11px] text-muted-foreground">
          {orgs.length} org{orgs.length === 1 ? "" : "s"} already allocated
        </span>
      </header>

      <div className="grid grid-cols-3 divide-x divide-border">
        <CapacityColumn
          icon={Cpu}
          label="CPU"
          unit="cores"
          total={host.cpu_cores}
          allocated={allocated.cpu_cores}
          available={available.cpu_cores}
        />
        <CapacityColumn
          icon={MemoryStick}
          label="RAM"
          unit="MB"
          total={host.ram_mb}
          allocated={allocated.ram_mb}
          available={available.ram_mb}
        />
        <CapacityColumn
          icon={HardDrive}
          label="Disk"
          unit="GB"
          total={host.disk_gb}
          allocated={allocated.disk_gb}
          available={available.disk_gb}
        />
      </div>

      {orgs.length > 0 && (
        <div className="border-t border-border">
          <div className="px-6 py-3 text-[11px] font-medium uppercase tracking-wider text-muted-foreground">
            Existing organizations
          </div>
          <div className="divide-y divide-border">
            {orgs.map((o) => (
              <div
                key={o.slug}
                className="grid grid-cols-4 items-center gap-3 px-6 py-2.5 text-sm"
              >
                <div className="flex items-center gap-2">
                  <div className="flex h-6 w-6 items-center justify-center rounded-md bg-brand/10 text-[10px] font-semibold text-brand">
                    {o.name.slice(0, 2).toUpperCase()}
                  </div>
                  <span className="truncate font-medium">{o.name}</span>
                </div>
                <div className="text-right text-xs text-muted-foreground">
                  <span className="font-medium text-foreground">{o.cpu_cores}</span> cores
                </div>
                <div className="text-right text-xs text-muted-foreground">
                  <span className="font-medium text-foreground">{o.ram_mb.toLocaleString()}</span> MB
                </div>
                <div className="text-right text-xs text-muted-foreground">
                  <span className="font-medium text-foreground">{o.disk_gb}</span> GB
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </section>
  );
}

function CapacityColumn({
  icon: Icon,
  label,
  unit,
  total,
  allocated,
  available,
}: {
  icon: typeof Cpu;
  label: string;
  unit: string;
  total: number;
  allocated: number;
  available: number;
}) {
  const pct = total > 0 ? Math.round((allocated / total) * 100) : 0;
  const barColor =
    pct > 90 ? "bg-destructive" : pct > 75 ? "bg-warning" : "bg-brand";

  return (
    <div className="p-5">
      <div className="flex items-center gap-1.5 text-[11px] font-medium uppercase tracking-wider text-muted-foreground">
        <Icon className="h-3 w-3" />
        {label}
      </div>
      <div className="mt-2.5 flex items-baseline gap-1.5">
        <span className="font-heading text-xl font-semibold tracking-tight">
          {available.toLocaleString()}
        </span>
        <span className="text-xs text-muted-foreground">
          / {total.toLocaleString()} {unit} free
        </span>
      </div>
      <div className="mt-2 h-1.5 w-full overflow-hidden rounded-full bg-muted">
        <div
          className={cn("h-full rounded-full transition-all", barColor)}
          style={{ width: `${Math.min(100, pct)}%` }}
        />
      </div>
      <div className="mt-1.5 text-[10px] text-muted-foreground">
        {allocated.toLocaleString()} {unit} allocated ({pct}%)
      </div>
    </div>
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
  <div className="space-y-1.5">
    <Label>
      <div className="flex items-center gap-1.5">
        <Icon className="h-3.5 w-3.5 text-muted-foreground" />
        {label}
      </div>
    </Label>
    <div className="relative">
      <Input type="number" className="pr-14" {...rest} />
      <span className="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-xs text-muted-foreground">
        {unit}
      </span>
    </div>
    {error && <p className="text-xs text-destructive">{error}</p>}
  </div>
);

export default CreateOrg;
