import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import {
  Loader2,
  AppWindow,
  Database,
  Clock,
  Rocket,
  Cpu,
  HardDrive,
  MemoryStick,
  ArrowRight,
  Plus,
  Store,
  Sparkles,
} from "lucide-react";
import type { LucideIcon } from "lucide-react";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { PageHelp } from "@/components/layout/PageHelp";
import { dashboardApi } from "@/api/dashboard";
import { useOrgStore } from "@/stores/org";
import { cn } from "@/lib/utils";

const statusVariant: Record<string, "success" | "destructive" | "brand" | "warning" | "secondary"> = {
  success: "success",
  succeeded: "success",
  failed: "destructive",
  error: "destructive",
  running: "brand",
  deploying: "brand",
  pending: "warning",
  queued: "warning",
};

function formatBytes(bytes?: number) {
  if (!bytes && bytes !== 0) return "—";
  const gb = bytes / 1024 / 1024 / 1024;
  if (gb >= 1) return `${gb.toFixed(1)} GB`;
  const mb = bytes / 1024 / 1024;
  return `${mb.toFixed(0)} MB`;
}

function DashboardOverview() {
  const currentOrg = useOrgStore((s) => s.currentOrg);
  const slug = currentOrg?.slug || "";

  const { data: dashData, isLoading } = useQuery({
    queryKey: ["dashboard", slug],
    queryFn: () => dashboardApi.getDashboard(slug),
    enabled: !!slug,
    refetchInterval: 30000,
  });

  const { data: metricsData } = useQuery({
    queryKey: ["metrics-overview", slug],
    queryFn: () => dashboardApi.getMetricsOverview(slug),
    enabled: !!slug,
    refetchInterval: 10000,
  });

  const dash = dashData?.data?.data;
  const metrics = metricsData?.data?.data;

  if (!currentOrg) {
    return (
      <div className="mx-auto flex max-w-lg flex-col items-center gap-5 py-20 text-center">
        <div className="flex h-14 w-14 items-center justify-center rounded-2xl border border-border bg-card shadow-sm">
          <Sparkles className="h-6 w-6 text-brand" />
        </div>
        <div>
          <h1 className="font-heading text-2xl font-semibold tracking-tight">
            Welcome to Orbita
          </h1>
          <p className="mt-2 text-sm text-muted-foreground">
            Create your first organization to begin deploying apps, databases,
            and scheduled jobs.
          </p>
        </div>
        <Link to="/orgs/new">
          <Button variant="brand" size="lg">
            <Plus className="h-4 w-4" />
            Create organization
          </Button>
        </Link>
      </div>
    );
  }

  if (isLoading) {
    return (
      <div className="flex justify-center py-20">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
      </div>
    );
  }

  return (
    <div className="space-y-8">
      {/* Page header */}
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="font-heading text-2xl font-semibold tracking-tight">
              Overview
            </h1>
            <PageHelp
              title="Dashboard overview"
              summary="Live snapshot of this organization's apps, databases, cron jobs, and resource usage."
              steps={[
                {
                  title: "Check the stat cards",
                  body: "Running apps, databases, active cron jobs, and CPU load. Numbers refresh every 30s.",
                },
                {
                  title: "Watch memory + disk gauges",
                  body: "Totals come from this org's quota. Used values are summed across this org's running containers.",
                },
                {
                  title: "Scan recent deployments",
                  body: "Every deploy and rollback shows up here. Click View All to jump into Projects.",
                },
              ]}
              nextLinks={[
                {
                  label: "Projects",
                  to: `/orgs/${slug}/projects`,
                  description: "Create projects + environments for your apps",
                },
                {
                  label: "Deploy a new app",
                  to: `/orgs/${slug}/apps/new`,
                  description: "From a Docker image or a git repository",
                },
                {
                  label: "Git connections",
                  to: `/orgs/${slug}/git`,
                  description: "Connect GitHub, GitLab, or Gitea for build-from-push",
                },
              ]}
            />
          </div>
          <p className="mt-1 text-sm text-muted-foreground">
            Platform health and recent activity for{" "}
            <span className="font-medium text-foreground">{currentOrg.name}</span>.
          </p>
        </div>
        <div className="flex flex-wrap gap-2">
          <Link to={`/orgs/${slug}/apps/new`}>
            <Button variant="outline" size="sm">
              <Rocket className="h-3.5 w-3.5" />
              Deploy app
            </Button>
          </Link>
          <Link to={`/orgs/${slug}/databases/new`}>
            <Button variant="outline" size="sm">
              <Database className="h-3.5 w-3.5" />
              New database
            </Button>
          </Link>
          <Link to={`/orgs/${slug}/services`}>
            <Button variant="brand" size="sm">
              <Store className="h-3.5 w-3.5" />
              Marketplace
            </Button>
          </Link>
        </div>
      </div>

      {/* Stat cards */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard
          icon={AppWindow}
          label="Running apps"
          value={dash?.running_apps ?? 0}
          sub={`${dash?.total_apps ?? 0} total`}
          tone="brand"
        />
        <StatCard
          icon={Database}
          label="Databases"
          value={dash?.running_databases ?? 0}
          sub={`${dash?.total_databases ?? 0} total`}
          tone="success"
        />
        <StatCard
          icon={Clock}
          label="Active cron"
          value={dash?.active_cron_jobs ?? 0}
          sub={`${dash?.total_cron_jobs ?? 0} jobs`}
          tone="warning"
        />
        <StatCard
          icon={Cpu}
          label="CPU load"
          value={`${(metrics?.cpu_percent ?? 0).toFixed(1)}%`}
          bar={metrics?.cpu_percent ?? 0}
          tone="brand"
        />
      </div>

      {/* Resources */}
      <div className="grid gap-4 lg:grid-cols-2">
        <ResourceCard
          icon={MemoryStick}
          label="Memory usage"
          used={metrics?.memory_used}
          total={metrics?.memory_total}
        />
        <ResourceCard
          icon={HardDrive}
          label="Disk usage"
          used={metrics?.disk_used}
          total={metrics?.disk_total}
        />
      </div>

      {/* Recent deployments */}
      <section className="overflow-hidden rounded-xl border border-border bg-card shadow-xs">
        <div className="flex items-center justify-between border-b border-border bg-muted/30 px-5 py-3">
          <div className="flex items-center gap-2">
            <Rocket className="h-4 w-4 text-muted-foreground" />
            <h2 className="text-sm font-semibold">Recent deployments</h2>
          </div>
          <Link
            to={`/orgs/${slug}/projects`}
            className="inline-flex items-center gap-1 text-xs font-medium text-muted-foreground transition-colors hover:text-foreground"
          >
            View all
            <ArrowRight className="h-3 w-3" />
          </Link>
        </div>

        {!dash?.recent_deploys?.length ? (
          <div className="flex flex-col items-center gap-2 px-6 py-12 text-center">
            <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-muted">
              <Rocket className="h-4 w-4 text-muted-foreground" />
            </div>
            <p className="text-sm text-muted-foreground">No deployments yet</p>
            <Link to={`/orgs/${slug}/apps/new`}>
              <Button variant="brand" size="sm" className="mt-2">
                Deploy your first app
              </Button>
            </Link>
          </div>
        ) : (
          <div className="divide-y divide-border">
            {dash.recent_deploys.map((d) => {
              const variant = statusVariant[d.status?.toLowerCase()] || "secondary";
              return (
                <div
                  key={d.id}
                  className="flex items-center justify-between px-5 py-3 transition-colors hover:bg-accent/30"
                >
                  <div className="flex min-w-0 items-center gap-3">
                    <div className="flex h-8 w-8 flex-shrink-0 items-center justify-center rounded-md bg-brand/10 text-brand">
                      <Rocket className="h-3.5 w-3.5" />
                    </div>
                    <div className="min-w-0">
                      <div className="flex items-center gap-2">
                        <span className="truncate text-sm font-medium">
                          {d.app_name}
                        </span>
                        <span className="font-mono text-[11px] text-muted-foreground">
                          v{d.version}
                        </span>
                      </div>
                      {d.started_at && (
                        <div className="text-[11px] text-muted-foreground">
                          {new Date(d.started_at).toLocaleString()}
                        </div>
                      )}
                    </div>
                  </div>
                  <Badge variant={variant}>{d.status}</Badge>
                </div>
              );
            })}
          </div>
        )}
      </section>
    </div>
  );
}

function StatCard({
  icon: Icon,
  label,
  value,
  sub,
  bar,
  tone = "brand",
}: {
  icon: LucideIcon;
  label: string;
  value: string | number;
  sub?: string;
  bar?: number;
  tone?: "brand" | "success" | "warning" | "destructive";
}) {
  const toneClasses = {
    brand: "bg-brand/10 text-brand ring-brand/20",
    success: "bg-success/10 text-success ring-success/20",
    warning: "bg-warning/10 text-warning-foreground ring-warning/30",
    destructive: "bg-destructive/10 text-destructive ring-destructive/20",
  };
  return (
    <div className="relative overflow-hidden rounded-xl border border-border bg-card p-5 shadow-xs transition-shadow hover:shadow-sm">
      <div className="flex items-start justify-between">
        <span className="text-xs font-medium uppercase tracking-wider text-muted-foreground">
          {label}
        </span>
        <div className={cn("flex h-8 w-8 items-center justify-center rounded-lg ring-1", toneClasses[tone])}>
          <Icon className="h-4 w-4" />
        </div>
      </div>
      <div className="mt-4 font-heading text-3xl font-semibold tracking-tight">
        {value}
      </div>
      {sub && <div className="mt-1 text-xs text-muted-foreground">{sub}</div>}
      {bar !== undefined && (
        <div className="mt-3 h-1.5 w-full overflow-hidden rounded-full bg-muted">
          <div
            className="h-full rounded-full bg-brand transition-all"
            style={{ width: `${Math.min(100, bar)}%` }}
          />
        </div>
      )}
    </div>
  );
}

function ResourceCard({
  icon: Icon,
  label,
  used,
  total,
}: {
  icon: LucideIcon;
  label: string;
  used?: number;
  total?: number;
}) {
  const pct = used && total ? (used / total) * 100 : 0;
  const barColor =
    pct > 90 ? "bg-destructive" : pct > 75 ? "bg-warning" : "bg-brand";

  return (
    <div className="rounded-xl border border-border bg-card p-5 shadow-xs">
      <div className="flex items-center gap-2">
        <Icon className="h-4 w-4 text-muted-foreground" />
        <span className="text-sm font-medium">{label}</span>
        <span className="ml-auto text-xs font-medium text-muted-foreground">
          {pct.toFixed(0)}%
        </span>
      </div>

      <div className="mt-4 flex items-baseline gap-2">
        <span className="font-heading text-2xl font-semibold tracking-tight">
          {formatBytes(used)}
        </span>
        <span className="text-sm text-muted-foreground">
          of {formatBytes(total)}
        </span>
      </div>

      <div className="mt-3 h-2 w-full overflow-hidden rounded-full bg-muted">
        <div
          className={cn("h-full rounded-full transition-all", barColor)}
          style={{ width: `${Math.min(100, pct)}%` }}
        />
      </div>
    </div>
  );
}

export default DashboardOverview;
