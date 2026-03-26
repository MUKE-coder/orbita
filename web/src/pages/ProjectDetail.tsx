import { useState } from "react";
import { useParams, Link } from "react-router-dom";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Plus, Loader2, ArrowLeft, Trash2 } from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import { Separator } from "@/components/ui/separator";
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
import { projectsApi } from "@/api/projects";
import { useOrgStore } from "@/stores/org";

const envSchema = z.object({
  name: z.string().min(1, "Name is required"),
});

type EnvForm = z.infer<typeof envSchema>;

const envTypeBadge: Record<string, string> = {
  production: "bg-green-500/10 text-green-500",
  staging: "bg-yellow-500/10 text-yellow-500",
  custom: "bg-blue-500/10 text-blue-500",
};

function ProjectDetail() {
  const { orgSlug, projectId } = useParams<{
    orgSlug: string;
    projectId: string;
  }>();
  const currentOrg = useOrgStore((s) => s.currentOrg);
  const queryClient = useQueryClient();
  const [envDialogOpen, setEnvDialogOpen] = useState(false);

  const slug = orgSlug || currentOrg?.slug || "";

  const { data, isLoading } = useQuery({
    queryKey: ["project", slug, projectId],
    queryFn: () => projectsApi.get(slug, projectId!),
    enabled: !!slug && !!projectId,
  });

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<EnvForm>({
    resolver: zodResolver(envSchema),
  });

  const createEnvMutation = useMutation({
    mutationFn: (data: EnvForm) =>
      projectsApi.createEnvironment(slug, projectId!, {
        name: data.name,
        type: "custom",
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["project", slug, projectId],
      });
      toast.success("Environment created");
      reset();
      setEnvDialogOpen(false);
    },
    onError: () => toast.error("Failed to create environment"),
  });

  const deleteEnvMutation = useMutation({
    mutationFn: (envId: string) =>
      projectsApi.deleteEnvironment(slug, projectId!, envId),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ["project", slug, projectId],
      });
      toast.success("Environment deleted");
    },
    onError: () => toast.error("Failed to delete environment"),
  });

  const project = data?.data?.data;

  if (isLoading) {
    return (
      <div className="flex justify-center py-12">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
      </div>
    );
  }

  if (!project) {
    return <p className="text-muted-foreground">Project not found</p>;
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3">
        <Link
          to="/"
          className="text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-5 w-5" />
        </Link>
        <span className="text-2xl">{project.emoji}</span>
        <div>
          <h1 className="text-xl font-semibold">{project.name}</h1>
          {project.description && (
            <p className="text-sm text-muted-foreground">
              {project.description}
            </p>
          )}
        </div>
      </div>

      <Separator />

      <div className="flex items-center justify-between">
        <h2 className="text-lg font-medium">Environments</h2>
        <Dialog open={envDialogOpen} onOpenChange={setEnvDialogOpen}>
          <DialogTrigger render={<Button size="sm" variant="outline"><Plus className="mr-2 h-4 w-4" />Add Environment</Button>} />
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Add Environment</DialogTitle>
            </DialogHeader>
            <form
              onSubmit={handleSubmit((d) => createEnvMutation.mutate(d))}
              className="space-y-4"
            >
              <div className="space-y-2">
                <Label htmlFor="envName">Name</Label>
                <Input
                  id="envName"
                  placeholder="Preview, QA, Dev..."
                  {...register("name")}
                />
                {errors.name && (
                  <p className="text-sm text-destructive">
                    {errors.name.message}
                  </p>
                )}
              </div>
              <Button
                type="submit"
                className="w-full"
                disabled={createEnvMutation.isPending}
              >
                {createEnvMutation.isPending && (
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                )}
                Create Environment
              </Button>
            </form>
          </DialogContent>
        </Dialog>
      </div>

      <div className="space-y-3">
        {project.environments?.map((env) => (
          <Card key={env.id}>
            <CardHeader className="py-3">
              <CardTitle className="flex items-center justify-between text-sm">
                <div className="flex items-center gap-2">
                  {env.name}
                  <Badge
                    variant="secondary"
                    className={envTypeBadge[env.type] || envTypeBadge.custom}
                  >
                    {env.type}
                  </Badge>
                </div>
                {env.type === "custom" && (
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-7 w-7 text-muted-foreground hover:text-destructive"
                    onClick={() => deleteEnvMutation.mutate(env.id)}
                  >
                    <Trash2 className="h-3.5 w-3.5" />
                  </Button>
                )}
              </CardTitle>
            </CardHeader>
            <CardContent className="py-2 text-xs text-muted-foreground">
              Resources will appear here once apps and databases are deployed
              (Phase 4+).
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}

export default ProjectDetail;
