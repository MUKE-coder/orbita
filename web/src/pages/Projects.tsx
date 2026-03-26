import { useState } from "react";
import { Link } from "react-router-dom";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Plus, Loader2, FolderOpen } from "lucide-react";
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
import { projectsApi } from "@/api/projects";
import { useOrgStore } from "@/stores/org";

const createProjectSchema = z.object({
  name: z.string().min(2, "Name must be at least 2 characters"),
  description: z.string().optional(),
  emoji: z.string().optional(),
});

type CreateProjectForm = z.infer<typeof createProjectSchema>;

function Projects() {
  const currentOrg = useOrgStore((s) => s.currentOrg);
  const queryClient = useQueryClient();
  const [dialogOpen, setDialogOpen] = useState(false);

  const { data, isLoading } = useQuery({
    queryKey: ["projects", currentOrg?.slug],
    queryFn: () => projectsApi.list(currentOrg!.slug),
    enabled: !!currentOrg,
  });

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<CreateProjectForm>({
    resolver: zodResolver(createProjectSchema),
  });

  const createMutation = useMutation({
    mutationFn: (data: CreateProjectForm) =>
      projectsApi.create(currentOrg!.slug, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["projects", currentOrg?.slug] });
      toast.success("Project created");
      reset();
      setDialogOpen(false);
    },
    onError: () => toast.error("Failed to create project"),
  });

  const projects = data?.data?.data || [];

  if (!currentOrg) {
    return (
      <div className="text-center text-muted-foreground py-12">
        Select an organization to view projects.
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold">Projects</h2>
        <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
          <DialogTrigger render={<Button size="sm"><Plus className="mr-2 h-4 w-4" />New Project</Button>} />
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Create Project</DialogTitle>
            </DialogHeader>
            <form
              onSubmit={handleSubmit((d) => createMutation.mutate(d))}
              className="space-y-4"
            >
              <div className="space-y-2">
                <Label htmlFor="name">Project Name</Label>
                <Input id="name" placeholder="My App" {...register("name")} />
                {errors.name && (
                  <p className="text-sm text-destructive">{errors.name.message}</p>
                )}
              </div>
              <div className="space-y-2">
                <Label htmlFor="description">Description (optional)</Label>
                <Input
                  id="description"
                  placeholder="A brief description"
                  {...register("description")}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="emoji">Emoji</Label>
                <Input
                  id="emoji"
                  placeholder="🚀"
                  className="w-20"
                  {...register("emoji")}
                />
              </div>
              <Button
                type="submit"
                className="w-full"
                disabled={createMutation.isPending}
              >
                {createMutation.isPending && (
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                )}
                Create Project
              </Button>
            </form>
          </DialogContent>
        </Dialog>
      </div>

      {isLoading ? (
        <div className="flex justify-center py-12">
          <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
        </div>
      ) : projects.length === 0 ? (
        <div className="flex flex-col items-center gap-3 py-12 text-center">
          <FolderOpen className="h-12 w-12 text-muted-foreground" />
          <p className="text-muted-foreground">No projects yet</p>
          <Button size="sm" onClick={() => setDialogOpen(true)}>
            <Plus className="mr-2 h-4 w-4" />
            Create your first project
          </Button>
        </div>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {projects.map((project) => (
            <Link
              key={project.id}
              to={`/orgs/${currentOrg.slug}/projects/${project.id}`}
            >
              <Card className="transition-colors hover:border-primary/50">
                <CardHeader className="pb-2">
                  <CardTitle className="flex items-center gap-2 text-base">
                    <span className="text-xl">{project.emoji || "🚀"}</span>
                    {project.name}
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  {project.description && (
                    <p className="text-sm text-muted-foreground line-clamp-2 mb-2">
                      {project.description}
                    </p>
                  )}
                  <div className="flex gap-1">
                    {project.environments?.map((env) => (
                      <Badge key={env.id} variant="secondary" className="text-xs">
                        {env.name}
                      </Badge>
                    ))}
                  </div>
                </CardContent>
              </Card>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}

export default Projects;
