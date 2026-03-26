import { useState } from "react";
import { useParams, Link } from "react-router-dom";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  Loader2,
  ArrowLeft,
  Play,
  Pause,
  Clock,
  FileText,
} from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
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
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { cronApi } from "@/api/cron";
import { useOrgStore } from "@/stores/org";

const runStatusColor: Record<string, string> = {
  running: "bg-blue-500/10 text-blue-500",
  success: "bg-green-500/10 text-green-500",
  failed: "bg-red-500/10 text-red-500",
  timeout: "bg-orange-500/10 text-orange-500",
  skipped: "bg-gray-500/10 text-gray-400",
};

function CronDetail() {
  const { orgSlug, cronId } = useParams<{ orgSlug: string; cronId: string }>();
  const currentOrg = useOrgStore((s) => s.currentOrg);
  const queryClient = useQueryClient();
  const slug = orgSlug || currentOrg?.slug || "";
  const [logDialogOpen, setLogDialogOpen] = useState(false);
  const [selectedRunLogs, setSelectedRunLogs] = useState("");

  const { data: jobData, isLoading } = useQuery({
    queryKey: ["cron-job", slug, cronId],
    queryFn: () => cronApi.get(slug, cronId!),
    enabled: !!slug && !!cronId,
  });

  const { data: runsData } = useQuery({
    queryKey: ["cron-runs", slug, cronId],
    queryFn: () => cronApi.listRuns(slug, cronId!),
    enabled: !!slug && !!cronId,
    refetchInterval: 5000,
  });

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: ["cron-job", slug, cronId] });
    queryClient.invalidateQueries({ queryKey: ["cron-runs", slug, cronId] });
  };

  const toggleMut = useMutation({
    mutationFn: () => cronApi.toggle(slug, cronId!),
    onSuccess: () => { invalidate(); toast.success("Toggled"); },
  });

  const triggerMut = useMutation({
    mutationFn: () => cronApi.trigger(slug, cronId!),
    onSuccess: () => { invalidate(); toast.success("Job triggered"); },
  });

  const viewLogs = async (runId: string) => {
    try {
      const res = await cronApi.getRunLogs(slug, cronId!, runId);
      setSelectedRunLogs(res.data.data.logs || "No logs available.");
      setLogDialogOpen(true);
    } catch {
      toast.error("Failed to load logs");
    }
  };

  const job = jobData?.data?.data;
  const runs = runsData?.data?.data || [];

  if (isLoading) {
    return <div className="flex justify-center py-12"><Loader2 className="h-6 w-6 animate-spin text-muted-foreground" /></div>;
  }

  if (!job) return <p className="text-muted-foreground">Cron job not found</p>;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Link to="/" className="text-muted-foreground hover:text-foreground">
            <ArrowLeft className="h-5 w-5" />
          </Link>
          <Clock className="h-5 w-5 text-muted-foreground" />
          <div>
            <h1 className="text-xl font-semibold">{job.name}</h1>
            <p className="text-sm text-muted-foreground font-mono">{job.schedule}</p>
          </div>
          <Badge variant={job.enabled ? "default" : "secondary"}>
            {job.enabled ? "Enabled" : "Disabled"}
          </Badge>
        </div>
        <div className="flex gap-2">
          <Button size="sm" onClick={() => triggerMut.mutate()} disabled={triggerMut.isPending}>
            {triggerMut.isPending ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Play className="mr-2 h-4 w-4" />}
            Run Now
          </Button>
          <Button size="sm" variant="outline" onClick={() => toggleMut.mutate()} disabled={toggleMut.isPending}>
            <Pause className="mr-2 h-4 w-4" />
            {job.enabled ? "Disable" : "Enable"}
          </Button>
        </div>
      </div>

      <Separator />

      {/* Info */}
      <Card>
        <CardHeader><CardTitle className="text-sm">Configuration</CardTitle></CardHeader>
        <CardContent className="grid grid-cols-2 gap-3 text-sm">
          <div><span className="text-muted-foreground">Image:</span> <span className="font-mono text-xs">{job.image}</span></div>
          <div><span className="text-muted-foreground">Command:</span> {job.command || "—"}</div>
          <div><span className="text-muted-foreground">Timeout:</span> {job.timeout}s</div>
          <div><span className="text-muted-foreground">Concurrency:</span> {job.concurrency_policy}</div>
          <div><span className="text-muted-foreground">Max Retries:</span> {job.max_retries}</div>
          <div><span className="text-muted-foreground">Next Run:</span> {job.next_run_at ? new Date(job.next_run_at).toLocaleString() : "—"}</div>
          <div><span className="text-muted-foreground">Last Run:</span> {job.last_run_at ? new Date(job.last_run_at).toLocaleString() : "Never"}</div>
        </CardContent>
      </Card>

      {/* Run History */}
      <div>
        <h2 className="text-lg font-medium mb-3">Run History</h2>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Started</TableHead>
              <TableHead>Duration</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Exit Code</TableHead>
              <TableHead></TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {runs.map((run) => (
              <TableRow key={run.id}>
                <TableCell className="text-xs">{new Date(run.started_at).toLocaleString()}</TableCell>
                <TableCell className="text-xs">
                  {run.duration_ms != null ? `${(run.duration_ms / 1000).toFixed(1)}s` : "—"}
                </TableCell>
                <TableCell>
                  <Badge className={runStatusColor[run.status] || ""} variant="secondary">{run.status}</Badge>
                </TableCell>
                <TableCell className="font-mono text-xs">{run.exit_code ?? "—"}</TableCell>
                <TableCell>
                  <Button size="sm" variant="ghost" onClick={() => viewLogs(run.id)}>
                    <FileText className="mr-1 h-3 w-3" />Logs
                  </Button>
                </TableCell>
              </TableRow>
            ))}
            {runs.length === 0 && (
              <TableRow>
                <TableCell colSpan={5} className="text-center text-muted-foreground py-8">
                  No runs yet. Click "Run Now" to trigger manually.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>

      {/* Log Dialog */}
      <Dialog open={logDialogOpen} onOpenChange={setLogDialogOpen}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>Run Logs</DialogTitle>
          </DialogHeader>
          <pre className="max-h-96 overflow-auto rounded-lg bg-gray-950 p-4 text-xs text-green-400 font-mono">
            {selectedRunLogs}
          </pre>
        </DialogContent>
      </Dialog>
    </div>
  );
}

export default CronDetail;
