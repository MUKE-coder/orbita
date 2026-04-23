import { ReactNode } from "react";
import { useParams, Link, useLocation } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import {
  Settings,
  Users,
  ArrowLeft,
  Cpu,
  MemoryStick,
  HardDrive,
  AppWindow,
  Database,
  Gift,
  CreditCard,
} from "lucide-react";

import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { Badge } from "@/components/ui/badge";
import { orgsApi, type Organization } from "@/api/orgs";
import OrgMembers from "./OrgMembers";

function OrgSettings({ children }: { children?: ReactNode }) {
  const { orgSlug } = useParams<{ orgSlug: string }>();
  const location = useLocation();

  const { data } = useQuery({
    queryKey: ["org", orgSlug],
    queryFn: () => orgsApi.get(orgSlug!),
    enabled: !!orgSlug,
  });

  const org = data?.data?.data?.organization;
  const isMembers = location.pathname.includes("/members");

  return (
    <div className="flex min-h-screen bg-background">
      {/* Settings sidebar */}
      <aside className="w-64 border-r border-border bg-sidebar p-4">
        <Link
          to="/dashboard"
          className="mb-4 inline-flex items-center gap-1.5 text-xs font-medium text-muted-foreground transition-colors hover:text-foreground"
        >
          <ArrowLeft className="h-3.5 w-3.5" />
          Back to dashboard
        </Link>
        <div className="mb-4">
          <h2 className="font-heading text-base font-semibold tracking-tight">
            {org?.name || "Settings"}
          </h2>
          <p className="font-mono text-[11px] text-muted-foreground">{orgSlug}</p>
        </div>
        <Separator className="mb-3" />
        <nav className="space-y-1">
          <Link to={`/orgs/${orgSlug}/settings`}>
            <Button
              variant={!isMembers ? "secondary" : "ghost"}
              className="w-full justify-start"
              size="sm"
            >
              <Settings className="mr-2 h-4 w-4" />
              General
            </Button>
          </Link>
          <Link to={`/orgs/${orgSlug}/settings/members`}>
            <Button
              variant={isMembers ? "secondary" : "ghost"}
              className="w-full justify-start"
              size="sm"
            >
              <Users className="mr-2 h-4 w-4" />
              Members
            </Button>
          </Link>
        </nav>
      </aside>

      {/* Content */}
      <main className="flex-1 overflow-y-auto">
        <div className="mx-auto w-full max-w-3xl p-8">
          {children || (isMembers ? <OrgMembers /> : <GeneralSettings />)}
        </div>
      </main>
    </div>
  );
}

function GeneralSettings() {
  const { orgSlug } = useParams<{ orgSlug: string }>();

  const { data } = useQuery({
    queryKey: ["org", orgSlug],
    queryFn: () => orgsApi.get(orgSlug!),
    enabled: !!orgSlug,
  });

  const org = data?.data?.data?.organization as Organization | undefined;
  if (!org) return null;

  const effective = {
    cpu: org.custom_cpu_cores ?? org.plan?.max_cpu_cores ?? 1,
    ram: org.custom_ram_mb ?? org.plan?.max_ram_mb ?? 1024,
    disk: org.custom_disk_gb ?? org.plan?.max_disk_gb ?? 10,
    apps: org.custom_max_apps ?? org.plan?.max_apps ?? 5,
    dbs: org.custom_max_databases ?? org.plan?.max_databases ?? 3,
  };

  const priceLabel =
    org.billing_type === "paid" && org.price_monthly_cents
      ? formatPrice(org.price_monthly_cents, org.currency, org.billing_cycle)
      : null;

  return (
    <div className="space-y-6">
      <div>
        <h1 className="font-heading text-2xl font-semibold tracking-tight">
          General settings
        </h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Organization identity, resource quotas, and billing.
        </p>
      </div>

      {/* Identity */}
      <section className="overflow-hidden rounded-xl border border-border bg-card shadow-xs">
        <header className="border-b border-border bg-muted/30 px-5 py-3">
          <h2 className="text-sm font-semibold">Identity</h2>
        </header>
        <dl className="divide-y divide-border">
          <Row label="Name" value={org.name} />
          <Row label="Slug" value={<code className="font-mono text-xs">{org.slug}</code>} />
          <Row
            label="Description"
            value={org.description || <span className="text-muted-foreground/60">None</span>}
          />
          <Row
            label="Plan template"
            value={org.plan?.name || "None"}
          />
        </dl>
      </section>

      {/* Resource quota */}
      <section className="overflow-hidden rounded-xl border border-border bg-card shadow-xs">
        <header className="flex items-center justify-between border-b border-border bg-muted/30 px-5 py-3">
          <h2 className="text-sm font-semibold">Resource quota</h2>
          <Badge variant={hasOverrides(org) ? "brand" : "outline"}>
            {hasOverrides(org) ? "Custom" : "From plan"}
          </Badge>
        </header>
        <div className="grid grid-cols-2 gap-px bg-border sm:grid-cols-5">
          <Metric icon={Cpu} label="CPU" value={effective.cpu} unit="cores" />
          <Metric icon={MemoryStick} label="RAM" value={effective.ram} unit="MB" />
          <Metric icon={HardDrive} label="Disk" value={effective.disk} unit="GB" />
          <Metric icon={AppWindow} label="Apps" value={effective.apps} unit="max" />
          <Metric icon={Database} label="DBs" value={effective.dbs} unit="max" />
        </div>
        <div className="border-t border-border px-5 py-3 text-[11px] text-muted-foreground">
          Super admin can change these via{" "}
          <code className="font-mono">PUT /admin/orgs/{org.slug}/resources</code>.
        </div>
      </section>

      {/* Billing */}
      <section className="overflow-hidden rounded-xl border border-border bg-card shadow-xs">
        <header className="border-b border-border bg-muted/30 px-5 py-3">
          <h2 className="text-sm font-semibold">Billing</h2>
        </header>
        <div className="flex items-center gap-4 px-5 py-5">
          <div
            className={`flex h-12 w-12 items-center justify-center rounded-xl ${
              org.billing_type === "paid"
                ? "bg-brand/10 text-brand"
                : "bg-success/10 text-success"
            }`}
          >
            {org.billing_type === "paid" ? (
              <CreditCard className="h-5 w-5" />
            ) : (
              <Gift className="h-5 w-5" />
            )}
          </div>
          <div className="min-w-0 flex-1">
            <div className="flex items-center gap-2">
              <span className="font-heading text-lg font-semibold">
                {priceLabel || "Free"}
              </span>
              <Badge
                variant={org.billing_type === "paid" ? "brand" : "success"}
                className="capitalize"
              >
                {org.billing_type}
              </Badge>
            </div>
            <p className="mt-0.5 text-xs text-muted-foreground">
              {org.billing_type === "paid"
                ? `Charged ${org.billing_cycle.replace("_", " ")} in ${org.currency}`
                : "No charge — internal or complimentary tenant"}
            </p>
          </div>
        </div>
      </section>
    </div>
  );
}

function hasOverrides(org: Organization) {
  return (
    org.custom_cpu_cores !== null ||
    org.custom_ram_mb !== null ||
    org.custom_disk_gb !== null ||
    org.custom_max_apps !== null ||
    org.custom_max_databases !== null
  );
}

function formatPrice(cents: number, currency: string, cycle: string) {
  const amount = cents / 100;
  const formatter = new Intl.NumberFormat("en-US", {
    style: "currency",
    currency,
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  });
  const cycleLabel = cycle === "yearly" ? "/yr" : cycle === "one_time" ? "" : "/mo";
  try {
    return `${formatter.format(amount)}${cycleLabel}`;
  } catch {
    return `${amount} ${currency}${cycleLabel}`;
  }
}

function Row({
  label,
  value,
}: {
  label: string;
  value: React.ReactNode;
}) {
  return (
    <div className="flex items-center justify-between px-5 py-3 text-sm">
      <dt className="text-muted-foreground">{label}</dt>
      <dd className="font-medium">{value}</dd>
    </div>
  );
}

function Metric({
  icon: Icon,
  label,
  value,
  unit,
}: {
  icon: typeof Cpu;
  label: string;
  value: number;
  unit: string;
}) {
  return (
    <div className="flex flex-col gap-1 bg-card px-4 py-4">
      <div className="flex items-center gap-1.5 text-[11px] font-medium uppercase tracking-wider text-muted-foreground">
        <Icon className="h-3 w-3" />
        {label}
      </div>
      <div className="font-heading text-xl font-semibold">{value.toLocaleString()}</div>
      <div className="text-[10px] text-muted-foreground">{unit}</div>
    </div>
  );
}

export default OrgSettings;
