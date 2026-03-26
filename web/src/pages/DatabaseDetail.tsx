import { useState } from "react";
import { useParams, Link } from "react-router-dom";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  Loader2,
  Play,
  Square,
  RotateCcw,
  ArrowLeft,
  Copy,
  Eye,
  EyeOff,
  Download,
  Undo2,
} from "lucide-react";
import { toast } from "sonner";

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
import { databasesApi } from "@/api/databases";
import { useOrgStore } from "@/stores/org";

const engineIcons: Record<string, string> = {
  postgres: "🐘",
  mysql: "🐬",
  mariadb: "🦭",
  mongodb: "🍃",
  redis: "🔴",
};

const statusColor: Record<string, string> = {
  running: "bg-green-500/10 text-green-500",
  stopped: "bg-gray-500/10 text-gray-400",
  creating: "bg-blue-500/10 text-blue-500",
  failed: "bg-red-500/10 text-red-500",
};

function DatabaseDetail() {
  const { orgSlug, dbId } = useParams<{ orgSlug: string; dbId: string }>();
  const currentOrg = useOrgStore((s) => s.currentOrg);
  const queryClient = useQueryClient();
  const slug = orgSlug || currentOrg?.slug || "";
  const [showConn, setShowConn] = useState(false);

  const { data: dbData, isLoading } = useQuery({
    queryKey: ["database", slug, dbId, showConn],
    queryFn: () => databasesApi.get(slug, dbId!, showConn),
    enabled: !!slug && !!dbId,
  });

  const { data: backupsData } = useQuery({
    queryKey: ["db-backups", slug, dbId],
    queryFn: () => databasesApi.listBackups(slug, dbId!),
    enabled: !!slug && !!dbId,
  });

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: ["database", slug, dbId] });
    queryClient.invalidateQueries({ queryKey: ["db-backups", slug, dbId] });
  };

  const restartMut = useMutation({
    mutationFn: () => databasesApi.restart(slug, dbId!),
    onSuccess: () => { invalidate(); toast.success("Database restarted"); },
  });
  const stopMut = useMutation({
    mutationFn: () => databasesApi.stop(slug, dbId!),
    onSuccess: () => { invalidate(); toast.success("Database stopped"); },
  });
  const startMut = useMutation({
    mutationFn: () => databasesApi.start(slug, dbId!),
    onSuccess: () => { invalidate(); toast.success("Database started"); },
  });
  const backupMut = useMutation({
    mutationFn: () => databasesApi.createBackup(slug, dbId!),
    onSuccess: () => { invalidate(); toast.success("Backup created"); },
  });
  const restoreMut = useMutation({
    mutationFn: (backupId: string) => databasesApi.restoreBackup(slug, dbId!, backupId),
    onSuccess: () => { toast.success("Restore initiated"); },
  });

  const mdb = dbData?.data?.data?.database;
  const connStr = dbData?.data?.data?.connection_string;
  const backups = backupsData?.data?.data || [];
  const anyPending = restartMut.isPending || stopMut.isPending || startMut.isPending;

  if (isLoading) {
    return <div className="flex justify-center py-12"><Loader2 className="h-6 w-6 animate-spin text-muted-foreground" /></div>;
  }

  if (!mdb) return <p className="text-muted-foreground">Database not found</p>;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Link to="/dashboard" className="text-muted-foreground hover:text-foreground"><ArrowLeft className="h-5 w-5" /></Link>
          <span className="text-2xl">{engineIcons[mdb.engine] || "🗄️"}</span>
          <div>
            <h1 className="text-xl font-semibold">{mdb.name}</h1>
            <p className="text-sm text-muted-foreground">{mdb.engine} {mdb.version}</p>
          </div>
          <Badge className={statusColor[mdb.status] || ""} variant="secondary">{mdb.status}</Badge>
        </div>
        <div className="flex gap-2">
          {mdb.status === "running" ? (
            <>
              <Button size="sm" variant="outline" onClick={() => restartMut.mutate()} disabled={anyPending}>
                <RotateCcw className="mr-2 h-4 w-4" />Restart
              </Button>
              <Button size="sm" variant="outline" onClick={() => stopMut.mutate()} disabled={anyPending}>
                <Square className="mr-2 h-4 w-4" />Stop
              </Button>
            </>
          ) : (
            <Button size="sm" variant="outline" onClick={() => startMut.mutate()} disabled={anyPending}>
              <Play className="mr-2 h-4 w-4" />Start
            </Button>
          )}
        </div>
      </div>

      <Separator />

      <Tabs defaultValue="connection">
        <TabsList>
          <TabsTrigger value="connection">Connection</TabsTrigger>
          <TabsTrigger value="backups">Backups</TabsTrigger>
          <TabsTrigger value="settings">Settings</TabsTrigger>
        </TabsList>

        <TabsContent value="connection" className="space-y-4">
          <Card>
            <CardHeader><CardTitle className="text-sm">Connection Details</CardTitle></CardHeader>
            <CardContent className="space-y-3">
              <div className="flex items-center gap-2">
                <code className="flex-1 rounded bg-muted p-2 text-xs font-mono break-all">
                  {showConn && connStr ? connStr : "••••••••••••••••••••••••••"}
                </code>
                <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => setShowConn(!showConn)}>
                  {showConn ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                </Button>
                {showConn && connStr && (
                  <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => { navigator.clipboard.writeText(connStr); toast.success("Copied"); }}>
                    <Copy className="h-4 w-4" />
                  </Button>
                )}
              </div>
              <div className="grid grid-cols-2 gap-2 text-sm">
                <div><span className="text-muted-foreground">Port:</span> {mdb.port || "—"}</div>
                <div><span className="text-muted-foreground">Volume:</span> {mdb.volume_name || "—"}</div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="backups" className="space-y-4">
          <div className="flex justify-between items-center">
            <h3 className="text-sm font-medium">Backup History</h3>
            <Button size="sm" onClick={() => backupMut.mutate()} disabled={backupMut.isPending}>
              {backupMut.isPending ? <Loader2 className="mr-2 h-4 w-4 animate-spin" /> : <Download className="mr-2 h-4 w-4" />}
              Create Backup
            </Button>
          </div>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Date</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>Size</TableHead>
                <TableHead></TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {backups.map((b) => (
                <TableRow key={b.id}>
                  <TableCell className="text-xs">{new Date(b.created_at).toLocaleString()}</TableCell>
                  <TableCell><Badge className={statusColor[b.status] || ""} variant="secondary">{b.status}</Badge></TableCell>
                  <TableCell className="text-xs">{(b.size_bytes / 1024 / 1024).toFixed(1)} MB</TableCell>
                  <TableCell>
                    {b.status === "completed" && (
                      <Button size="sm" variant="ghost" onClick={() => restoreMut.mutate(b.id)} disabled={restoreMut.isPending}>
                        <Undo2 className="mr-1 h-3 w-3" />Restore
                      </Button>
                    )}
                  </TableCell>
                </TableRow>
              ))}
              {backups.length === 0 && (
                <TableRow><TableCell colSpan={4} className="text-center text-muted-foreground py-8">No backups yet</TableCell></TableRow>
              )}
            </TableBody>
          </Table>
        </TabsContent>

        <TabsContent value="settings">
          <Card>
            <CardHeader><CardTitle className="text-sm">Database Info</CardTitle></CardHeader>
            <CardContent className="grid grid-cols-2 gap-3 text-sm">
              <div><span className="text-muted-foreground">Engine:</span> {mdb.engine}</div>
              <div><span className="text-muted-foreground">Version:</span> {mdb.version}</div>
              <div><span className="text-muted-foreground">Status:</span> {mdb.status}</div>
              <div><span className="text-muted-foreground">Port:</span> {mdb.port || "—"}</div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  );
}

export default DatabaseDetail;
