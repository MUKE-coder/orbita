import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Loader2, Plus, Server, Trash2, PauseCircle } from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { PageHelp } from "@/components/layout/PageHelp";
import { adminApi } from "@/api/admin";

const nodeStatusColor: Record<string, string> = {
  online: "bg-green-500/10 text-green-500",
  offline: "bg-red-500/10 text-red-500",
  pending: "bg-yellow-500/10 text-yellow-500",
  draining: "bg-orange-500/10 text-orange-500",
};

const addNodeSchema = z.object({
  name: z.string().min(1),
  ip: z.string().min(1),
  ssh_port: z.string().optional().transform((v) => (v ? parseInt(v) : 22)),
  ssh_private_key: z.string().min(1, "SSH key is required"),
});

function AdminNodes() {
  const queryClient = useQueryClient();
  const [dialogOpen, setDialogOpen] = useState(false);

  const { data, isLoading } = useQuery({
    queryKey: ["admin-nodes"],
    queryFn: () => adminApi.listNodes(),
  });

  const { data: platformData } = useQuery({
    queryKey: ["platform-metrics"],
    queryFn: () => adminApi.getPlatformMetrics(),
  });

  const { register, handleSubmit, reset, formState: { errors } } = useForm({
    resolver: zodResolver(addNodeSchema),
    defaultValues: { name: "", ip: "", ssh_port: "22", ssh_private_key: "" },
  });

  const addMut = useMutation({
    mutationFn: (data: { name: string; ip: string; ssh_port: number; ssh_private_key: string }) =>
      adminApi.addNode(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin-nodes"] });
      toast.success("Node added");
      reset();
      setDialogOpen(false);
    },
    onError: () => toast.error("Failed to add node"),
  });

  const drainMut = useMutation({
    mutationFn: (id: string) => adminApi.drainNode(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin-nodes"] });
      toast.success("Node draining");
    },
  });

  const removeMut = useMutation({
    mutationFn: (id: string) => adminApi.removeNode(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["admin-nodes"] });
      toast.success("Node removed");
    },
  });

  const nodes = data?.data?.data || [];
  const platform = platformData?.data?.data;

  return (
    <div className="space-y-6">
      <div className="flex items-start justify-between gap-4">
        <div>
          <div className="flex items-center gap-2">
            <h2 className="font-heading text-2xl font-semibold tracking-tight">
              Node Management
            </h2>
            <PageHelp
              title="Node management"
              summary="Add more servers to your Orbita cluster so apps can be spread across multiple machines."
              steps={[
                {
                  title: "Why add a node?",
                  body: "Your VPS is a single machine. When it runs out of RAM or CPU, you add another server as a 'worker node'. Orbita's Docker Swarm spreads apps across all healthy nodes.",
                },
                {
                  title: "Prepare the target server",
                  body: "Fresh Ubuntu/Debian VPS with Docker installed. SSH access via key-based auth (not password).",
                },
                {
                  title: "Click Add Node",
                  body: "Give it a name, paste its public IP, SSH port (default 22), and the SSH private key contents. Orbita SSHes in and runs `docker swarm join` with the cluster token.",
                },
                {
                  title: "Drain vs Remove",
                  body: "Drain: stops scheduling new tasks, gracefully moves existing containers to other nodes (do this before maintenance). Remove: hard removes the node from the cluster.",
                },
                {
                  title: "Single-server is fine too",
                  body: "If you only have one VPS, you don't need to add any nodes. Orbita runs everything on the manager (your current server).",
                },
              ]}
              nextLinks={[
                {
                  label: "Admin → Organizations",
                  to: `/admin/orgs`,
                  description: "Manage per-org resource quotas",
                },
              ]}
            />
          </div>
          {platform && (
            <p className="mt-1 text-sm text-muted-foreground">
              {platform.online_nodes}/{platform.total_nodes} nodes online
            </p>
          )}
        </div>
        <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
          <DialogTrigger render={<Button size="sm"><Plus className="mr-2 h-4 w-4" />Add Node</Button>} />
          <DialogContent>
            <DialogHeader><DialogTitle>Add Worker Node</DialogTitle></DialogHeader>
            <form onSubmit={handleSubmit((d) => addMut.mutate(d as { name: string; ip: string; ssh_port: number; ssh_private_key: string }))} className="space-y-4">
              <div className="space-y-2">
                <Label>Name</Label>
                <Input placeholder="worker-01" {...register("name")} />
                {errors.name && <p className="text-xs text-destructive">{errors.name.message}</p>}
              </div>
              <div className="grid grid-cols-3 gap-3">
                <div className="col-span-2 space-y-2">
                  <Label>IP Address</Label>
                  <Input placeholder="192.168.1.100" {...register("ip")} />
                </div>
                <div className="space-y-2">
                  <Label>SSH Port</Label>
                  <Input placeholder="22" {...register("ssh_port")} />
                </div>
              </div>
              <div className="space-y-2">
                <Label>SSH Private Key</Label>
                <textarea
                  className="w-full rounded-md border bg-background px-3 py-2 text-xs font-mono h-24 resize-none"
                  placeholder="-----BEGIN OPENSSH PRIVATE KEY-----"
                  {...register("ssh_private_key")}
                />
                {errors.ssh_private_key && <p className="text-xs text-destructive">{errors.ssh_private_key.message}</p>}
              </div>
              <Button type="submit" className="w-full" disabled={addMut.isPending}>
                {addMut.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                Add Node
              </Button>
            </form>
          </DialogContent>
        </Dialog>
      </div>

      {isLoading ? (
        <div className="flex justify-center py-12"><Loader2 className="h-6 w-6 animate-spin text-muted-foreground" /></div>
      ) : nodes.length === 0 ? (
        <p className="text-center text-muted-foreground py-12">No worker nodes added yet.</p>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {nodes.map((node) => (
            <Card key={node.id}>
              <CardHeader className="pb-2">
                <CardTitle className="flex items-center justify-between text-sm">
                  <div className="flex items-center gap-2">
                    <Server className="h-4 w-4 text-muted-foreground" />
                    {node.name}
                  </div>
                  <Badge className={nodeStatusColor[node.status] || ""} variant="secondary">
                    {node.status}
                  </Badge>
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-2 text-xs text-muted-foreground">
                <div className="flex justify-between">
                  <span>IP:</span> <span className="font-mono">{node.ip}</span>
                </div>
                <div className="flex justify-between">
                  <span>Role:</span> <span>{node.role}</span>
                </div>
                <div className="flex gap-2 pt-2">
                  {node.status === "online" && (
                    <Button size="sm" variant="outline" className="flex-1 text-xs" onClick={() => drainMut.mutate(node.id)}>
                      <PauseCircle className="mr-1 h-3 w-3" />Drain
                    </Button>
                  )}
                  <Button size="sm" variant="outline" className="flex-1 text-xs text-destructive" onClick={() => removeMut.mutate(node.id)}>
                    <Trash2 className="mr-1 h-3 w-3" />Remove
                  </Button>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}

export default AdminNodes;
