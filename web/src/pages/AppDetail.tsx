import { useParams } from "react-router-dom";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  Loader2,
  Play,
  Square,
  RotateCcw,
  Rocket,
  ArrowLeft,
  Undo2,
} from "lucide-react";
import { toast } from "sonner";
import { Link } from "react-router-dom";

import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
import {
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from "@/components/ui/tabs";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { appsApi } from "@/api/apps";
import { useOrgStore } from "@/stores/org";
import DomainList from "@/components/domains/DomainList";
import EnvVarEditor from "@/components/env/EnvVarEditor";
import XtermTerminal from "@/components/terminal/XtermTerminal";
import { useAuthStore } from "@/stores/auth";

const statusColor: Record<string, string> = {
  running: "bg-green-500/10 text-green-500",
  stopped: "bg-gray-500/10 text-gray-400",
  deploying: "bg-blue-500/10 text-blue-500",
  created: "bg-yellow-500/10 text-yellow-500",
  failed: "bg-red-500/10 text-red-500",
};

function AppDetail() {
  const { orgSlug, appId } = useParams<{ orgSlug: string; appId: string }>();
  const currentOrg = useOrgStore((s) => s.currentOrg);
  const queryClient = useQueryClient();
  const slug = orgSlug || currentOrg?.slug || "";
  const accessToken = useAuthStore((s) => s.accessToken);

  const { data: appData, isLoading } = useQuery({
    queryKey: ["app", slug, appId],
    queryFn: () => appsApi.get(slug, appId!),
    enabled: !!slug && !!appId,
  });

  const { data: deploymentsData } = useQuery({
    queryKey: ["deployments", slug, appId],
    queryFn: () => appsApi.listDeployments(slug, appId!),
    enabled: !!slug && !!appId,
  });

  const { data: logsData } = useQuery({
    queryKey: ["logs", slug, appId],
    queryFn: () => appsApi.getLogs(slug, appId!),
    enabled: !!slug && !!appId,
    refetchInterval: 5000,
  });

  const { data: metricsData } = useQuery({
    queryKey: ["metrics", slug, appId],
    queryFn: () => appsApi.getMetrics(slug, appId!),
    enabled: !!slug && !!appId,
    refetchInterval: 10000,
  });

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: ["app", slug, appId] });
    queryClient.invalidateQueries({ queryKey: ["deployments", slug, appId] });
  };

  const deployMutation = useMutation({
    mutationFn: () => appsApi.deploy(slug, appId!),
    onSuccess: () => { invalidate(); toast.success("Deploy triggered"); },
    onError: () => toast.error("Deploy failed"),
  });

  const stopMutation = useMutation({
    mutationFn: () => appsApi.stop(slug, appId!),
    onSuccess: () => { invalidate(); toast.success("App stopped"); },
    onError: () => toast.error("Failed to stop"),
  });

  const startMutation = useMutation({
    mutationFn: () => appsApi.start(slug, appId!),
    onSuccess: () => { invalidate(); toast.success("App started"); },
    onError: () => toast.error("Failed to start"),
  });

  const restartMutation = useMutation({
    mutationFn: () => appsApi.restart(slug, appId!),
    onSuccess: () => { invalidate(); toast.success("App restarted"); },
    onError: () => toast.error("Failed to restart"),
  });

  const rollbackMutation = useMutation({
    mutationFn: (deployId: string) => appsApi.rollback(slug, appId!, deployId),
    onSuccess: () => { invalidate(); toast.success("Rollback triggered"); },
    onError: () => toast.error("Rollback failed"),
  });

  const app = appData?.data?.data;
  const deployments = deploymentsData?.data?.data || [];
  const logs = logsData?.data?.data?.logs || "";
  const metrics = metricsData?.data?.data;
  const anyPending = deployMutation.isPending || stopMutation.isPending || startMutation.isPending || restartMutation.isPending;

  if (isLoading) {
    return (
      <div className="flex justify-center py-12">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
      </div>
    );
  }

  if (!app) return <p className="text-muted-foreground">App not found</p>;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Link to="/dashboard" className="text-muted-foreground hover:text-foreground">
            <ArrowLeft className="h-5 w-5" />
          </Link>
          <div>
            <h1 className="text-xl font-semibold">{app.name}</h1>
            <p className="text-sm text-muted-foreground">
              {app.source_config?.image || app.source_type}
            </p>
          </div>
          <Badge className={statusColor[app.status] || ""} variant="secondary">
            {app.status}
          </Badge>
        </div>
        <div className="flex gap-2">
          <Button
            size="sm"
            onClick={() => deployMutation.mutate()}
            disabled={anyPending}
          >
            {deployMutation.isPending ? (
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            ) : (
              <Rocket className="mr-2 h-4 w-4" />
            )}
            Deploy
          </Button>
          {app.status === "running" ? (
            <>
              <Button size="sm" variant="outline" onClick={() => restartMutation.mutate()} disabled={anyPending}>
                <RotateCcw className="mr-2 h-4 w-4" />Restart
              </Button>
              <Button size="sm" variant="outline" onClick={() => stopMutation.mutate()} disabled={anyPending}>
                <Square className="mr-2 h-4 w-4" />Stop
              </Button>
            </>
          ) : (
            <Button size="sm" variant="outline" onClick={() => startMutation.mutate()} disabled={anyPending}>
              <Play className="mr-2 h-4 w-4" />Start
            </Button>
          )}
        </div>
      </div>

      <Separator />

      {/* Tabs */}
      <Tabs defaultValue="overview">
        <TabsList>
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="deployments">Deployments</TabsTrigger>
          <TabsTrigger value="logs">Logs</TabsTrigger>
          <TabsTrigger value="metrics">Metrics</TabsTrigger>
          <TabsTrigger value="env">Environment</TabsTrigger>
          <TabsTrigger value="terminal">Terminal</TabsTrigger>
          <TabsTrigger value="domains">Domains</TabsTrigger>
        </TabsList>

        {/* Overview */}
        <TabsContent value="overview" className="space-y-4">
          <Card>
            <CardHeader><CardTitle className="text-sm">Application Info</CardTitle></CardHeader>
            <CardContent className="grid grid-cols-2 gap-3 text-sm">
              <div><span className="text-muted-foreground">Status:</span> {app.status}</div>
              <div><span className="text-muted-foreground">Replicas:</span> {app.replicas}</div>
              <div><span className="text-muted-foreground">Port:</span> {app.port || "—"}</div>
              <div><span className="text-muted-foreground">Source:</span> {app.source_type}</div>
              <div className="col-span-2"><span className="text-muted-foreground">Image:</span> {app.source_config?.image || "—"}</div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Deployments */}
        <TabsContent value="deployments">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Version</TableHead>
                <TableHead>Image</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Trigger</TableHead>
                <TableHead>Started</TableHead>
                <TableHead></TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {deployments.map((d) => (
                <TableRow key={d.id}>
                  <TableCell className="font-mono">v{d.version}</TableCell>
                  <TableCell className="font-mono text-xs max-w-48 truncate">{d.image_ref}</TableCell>
                  <TableCell>
                    <Badge className={statusColor[d.status] || ""} variant="secondary">{d.status}</Badge>
                  </TableCell>
                  <TableCell>{d.trigger_type}</TableCell>
                  <TableCell className="text-xs">
                    {d.started_at ? new Date(d.started_at).toLocaleString() : "—"}
                  </TableCell>
                  <TableCell>
                    {d.status === "success" && (
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => rollbackMutation.mutate(d.id)}
                        disabled={rollbackMutation.isPending}
                      >
                        <Undo2 className="mr-1 h-3 w-3" />Rollback
                      </Button>
                    )}
                  </TableCell>
                </TableRow>
              ))}
              {deployments.length === 0 && (
                <TableRow>
                  <TableCell colSpan={6} className="text-center text-muted-foreground py-8">
                    No deployments yet. Click Deploy to get started.
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        </TabsContent>

        {/* Logs */}
        <TabsContent value="logs">
          <Card>
            <CardContent className="p-0">
              <pre className="max-h-96 overflow-auto rounded-lg bg-gray-950 p-4 text-xs text-green-400 font-mono">
                {logs || "No logs available. Deploy the app first."}
              </pre>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Metrics */}
        <TabsContent value="metrics">
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <Card>
              <CardHeader className="pb-1"><CardTitle className="text-xs text-muted-foreground">CPU</CardTitle></CardHeader>
              <CardContent><p className="text-2xl font-bold">{metrics?.cpu_percent?.toFixed(1) ?? 0}%</p></CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-1"><CardTitle className="text-xs text-muted-foreground">Memory</CardTitle></CardHeader>
              <CardContent>
                <p className="text-2xl font-bold">
                  {metrics ? `${(metrics.memory_usage / 1024 / 1024).toFixed(0)} MB` : "0 MB"}
                </p>
                <p className="text-xs text-muted-foreground">
                  / {metrics ? `${(metrics.memory_limit / 1024 / 1024).toFixed(0)} MB` : "—"}
                </p>
              </CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-1"><CardTitle className="text-xs text-muted-foreground">Network In</CardTitle></CardHeader>
              <CardContent><p className="text-2xl font-bold">{metrics ? `${(metrics.network_rx / 1024).toFixed(0)} KB` : "0"}</p></CardContent>
            </Card>
            <Card>
              <CardHeader className="pb-1"><CardTitle className="text-xs text-muted-foreground">Network Out</CardTitle></CardHeader>
              <CardContent><p className="text-2xl font-bold">{metrics ? `${(metrics.network_tx / 1024).toFixed(0)} KB` : "0"}</p></CardContent>
            </Card>
          </div>
        </TabsContent>

        {/* Domains */}
        {/* Environment Variables */}
        <TabsContent value="env">
          <EnvVarEditor orgSlug={slug} appId={appId!} />
        </TabsContent>

        {/* Terminal */}
        <TabsContent value="terminal">
          {accessToken && appId ? (
            <XtermTerminal
              wsUrl={`${window.location.protocol === "https:" ? "wss:" : "ws:"}//${window.location.host}/api/v1/orgs/${slug}/apps/${appId}/terminal?token=${accessToken}`}
            />
          ) : (
            <p className="text-muted-foreground text-center py-8">
              Authentication required for terminal access.
            </p>
          )}
        </TabsContent>

        <TabsContent value="domains">
          <DomainList orgSlug={slug} appId={appId!} />
        </TabsContent>
      </Tabs>
    </div>
  );
}

export default AppDetail;
