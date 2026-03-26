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
} from "lucide-react";

import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { dashboardApi } from "@/api/dashboard";
import { useOrgStore } from "@/stores/org";

const statusColor: Record<string, string> = {
  success: "bg-green-500/10 text-green-500",
  failed: "bg-red-500/10 text-red-500",
  running: "bg-blue-500/10 text-blue-500",
  pending: "bg-yellow-500/10 text-yellow-500",
};

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
      <div className="flex flex-col items-center gap-4 py-12 text-center">
        <h1 className="text-3xl font-bold">Welcome to Orbita</h1>
        <p className="text-muted-foreground">Create or select an organization to get started.</p>
        <Link to="/orgs/new"><Button>Create your first organization</Button></Link>
      </div>
    );
  }

  if (isLoading) {
    return <div className="flex justify-center py-12"><Loader2 className="h-6 w-6 animate-spin text-muted-foreground" /></div>;
  }

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Dashboard</h1>

      {/* Overview cards */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Running Apps</CardTitle>
            <AppWindow className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <p className="text-3xl font-bold">{dash?.running_apps || 0}</p>
            <p className="text-xs text-muted-foreground">{dash?.total_apps || 0} total</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Databases</CardTitle>
            <Database className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <p className="text-3xl font-bold">{dash?.running_databases || 0}</p>
            <p className="text-xs text-muted-foreground">{dash?.total_databases || 0} total</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">Active Cron Jobs</CardTitle>
            <Clock className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <p className="text-3xl font-bold">{dash?.active_cron_jobs || 0}</p>
            <p className="text-xs text-muted-foreground">{dash?.total_cron_jobs || 0} total</p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium text-muted-foreground">CPU Usage</CardTitle>
            <Cpu className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            <p className="text-3xl font-bold">{metrics?.cpu_percent?.toFixed(1) || 0}%</p>
            <div className="mt-1 h-2 w-full rounded-full bg-muted">
              <div className="h-2 rounded-full bg-primary" style={{ width: `${metrics?.cpu_percent || 0}%` }} />
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Resource usage */}
      <div className="grid gap-4 sm:grid-cols-2">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm flex items-center gap-2">
              <HardDrive className="h-4 w-4" /> Memory Usage
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-baseline gap-2">
              <p className="text-2xl font-bold">
                {metrics ? `${(metrics.memory_used / 1024 / 1024 / 1024).toFixed(1)} GB` : "—"}
              </p>
              <p className="text-sm text-muted-foreground">
                / {metrics ? `${(metrics.memory_total / 1024 / 1024 / 1024).toFixed(1)} GB` : "—"}
              </p>
            </div>
            <div className="mt-2 h-2 w-full rounded-full bg-muted">
              <div
                className="h-2 rounded-full bg-primary"
                style={{ width: `${metrics ? (metrics.memory_used / metrics.memory_total * 100) : 0}%` }}
              />
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm flex items-center gap-2">
              <HardDrive className="h-4 w-4" /> Disk Usage
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="flex items-baseline gap-2">
              <p className="text-2xl font-bold">
                {metrics ? `${(metrics.disk_used / 1024 / 1024 / 1024).toFixed(1)} GB` : "—"}
              </p>
              <p className="text-sm text-muted-foreground">
                / {metrics ? `${(metrics.disk_total / 1024 / 1024 / 1024).toFixed(1)} GB` : "—"}
              </p>
            </div>
            <div className="mt-2 h-2 w-full rounded-full bg-muted">
              <div
                className="h-2 rounded-full bg-primary"
                style={{ width: `${metrics ? (metrics.disk_used / metrics.disk_total * 100) : 0}%` }}
              />
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Recent deploys */}
      <Card>
        <CardHeader>
          <CardTitle className="text-sm flex items-center gap-2">
            <Rocket className="h-4 w-4" /> Recent Deployments
          </CardTitle>
        </CardHeader>
        <CardContent>
          {!dash?.recent_deploys?.length ? (
            <p className="text-sm text-muted-foreground text-center py-4">No recent deployments</p>
          ) : (
            <div className="space-y-2">
              {dash.recent_deploys.map((d) => (
                <div key={d.id} className="flex items-center justify-between rounded border px-3 py-2 text-sm">
                  <div className="flex items-center gap-3">
                    <span className="font-medium">{d.app_name}</span>
                    <span className="font-mono text-xs text-muted-foreground">v{d.version}</span>
                  </div>
                  <div className="flex items-center gap-2">
                    <Badge className={statusColor[d.status] || ""} variant="secondary">{d.status}</Badge>
                    <span className="text-xs text-muted-foreground">
                      {d.started_at ? new Date(d.started_at).toLocaleString() : "—"}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Quick actions */}
      <div className="flex gap-3">
        <Link to={`/orgs/${slug}/apps/new`}>
          <Button size="sm" variant="outline">Deploy New App</Button>
        </Link>
        <Link to={`/orgs/${slug}/databases/new`}>
          <Button size="sm" variant="outline">New Database</Button>
        </Link>
        <Link to={`/orgs/${slug}/services`}>
          <Button size="sm" variant="outline">Add Service</Button>
        </Link>
      </div>
    </div>
  );
}

export default DashboardOverview;
