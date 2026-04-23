import { useState } from "react";
import { Link } from "react-router-dom";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import {
  Plus,
  Loader2,
  FolderKanban,
  ArrowUpRight,
  Search,
} from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { PageHelp } from "@/components/layout/PageHelp";
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
  const [query, setQuery] = useState("");

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
  const filtered = query
    ? projects.filter(
        (p) =>
          p.name.toLowerCase().includes(query.toLowerCase()) ||
          p.description?.toLowerCase().includes(query.toLowerCase())
      )
    : projects;

  if (!currentOrg) {
    return (
      <div className="py-12 text-center text-sm text-muted-foreground">
        Select an organization to view projects.
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="font-heading text-2xl font-semibold tracking-tight">
              Projects
            </h1>
            <PageHelp
              title="Projects"
              summary="Group apps, databases, and cron jobs by logical unit (e.g., one per website or client)."
              steps={[
                {
                  title: "Create a project",
                  body: "Give it a name and an emoji for quick scanning.",
                },
                {
                  title: "Environments auto-create",
                  body: "Every new project gets Production + Staging by default. You can rename or add more.",
                },
                {
                  title: "Deploy apps into it",
                  body: "From Projects, add apps + databases + cron jobs. They stay scoped to the project's environments.",
                },
              ]}
              nextLinks={[
                {
                  label: "Deploy a new app",
                  to: `/orgs/${currentOrg.slug}/apps/new`,
                  description: "From a Docker image or git repo",
                },
                {
                  label: "Provision a database",
                  to: `/orgs/${currentOrg.slug}/databases/new`,
                  description: "PostgreSQL, MySQL, MongoDB, Redis, etc.",
                },
              ]}
            />
          </div>
          <p className="mt-1 text-sm text-muted-foreground">
            Group apps, databases, and cron jobs by project.
          </p>
        </div>

        <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
          <DialogTrigger
            render={
              <Button variant="brand" size="sm">
                <Plus className="h-3.5 w-3.5" />
                New project
              </Button>
            }
          />
          <DialogContent>
            <DialogHeader>
              <DialogTitle>Create project</DialogTitle>
            </DialogHeader>
            <form
              onSubmit={handleSubmit((d) => createMutation.mutate(d))}
              className="space-y-4"
            >
              <div className="space-y-1.5">
                <Label htmlFor="name">Project name</Label>
                <Input
                  id="name"
                  placeholder="my-saas"
                  autoFocus
                  {...register("name")}
                />
                {errors.name && (
                  <p className="text-xs text-destructive">{errors.name.message}</p>
                )}
              </div>
              <div className="space-y-1.5">
                <Label htmlFor="description">Description</Label>
                <Input
                  id="description"
                  placeholder="Optional — a brief description"
                  {...register("description")}
                />
              </div>
              <div className="space-y-1.5">
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
                variant="brand"
                className="w-full"
                disabled={createMutation.isPending}
              >
                {createMutation.isPending && (
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                )}
                Create project
              </Button>
            </form>
          </DialogContent>
        </Dialog>
      </div>

      {/* Search */}
      {projects.length > 0 && (
        <div className="relative max-w-md">
          <Search className="pointer-events-none absolute left-3 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Filter projects..."
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            className="pl-9"
          />
        </div>
      )}

      {/* Projects */}
      {isLoading ? (
        <div className="flex justify-center py-16">
          <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
        </div>
      ) : projects.length === 0 ? (
        <EmptyState onCreate={() => setDialogOpen(true)} />
      ) : filtered.length === 0 ? (
        <div className="rounded-xl border border-dashed border-border bg-card/50 px-6 py-16 text-center">
          <p className="text-sm text-muted-foreground">
            No projects match "{query}"
          </p>
        </div>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {filtered.map((project) => (
            <Link
              key={project.id}
              to={`/orgs/${currentOrg.slug}/projects/${project.id}`}
              className="group"
            >
              <div className="relative flex h-full flex-col overflow-hidden rounded-xl border border-border bg-card p-5 shadow-xs transition-all hover:border-brand/30 hover:shadow-md">
                <div className="flex items-start justify-between">
                  <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-muted text-xl">
                    {project.emoji || "📦"}
                  </div>
                  <ArrowUpRight className="h-4 w-4 text-muted-foreground transition-colors group-hover:text-brand" />
                </div>

                <h3 className="mt-4 font-heading text-base font-semibold tracking-tight">
                  {project.name}
                </h3>

                {project.description ? (
                  <p className="mt-1 line-clamp-2 text-sm leading-relaxed text-muted-foreground">
                    {project.description}
                  </p>
                ) : (
                  <p className="mt-1 text-sm text-muted-foreground/60">
                    No description
                  </p>
                )}

                <div className="mt-auto flex items-center gap-1.5 pt-4">
                  {project.environments?.length ? (
                    project.environments.map((env) => (
                      <Badge
                        key={env.id}
                        variant="outline"
                        className="h-5 px-1.5 text-[10px] font-medium"
                      >
                        {env.name}
                      </Badge>
                    ))
                  ) : (
                    <span className="text-[11px] text-muted-foreground/60">
                      No environments
                    </span>
                  )}
                </div>
              </div>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}

function EmptyState({ onCreate }: { onCreate: () => void }) {
  return (
    <div className="flex flex-col items-center gap-4 rounded-xl border border-dashed border-border bg-card/50 px-6 py-16 text-center">
      <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-muted">
        <FolderKanban className="h-5 w-5 text-muted-foreground" />
      </div>
      <div>
        <h3 className="font-heading text-base font-semibold">
          No projects yet
        </h3>
        <p className="mt-1 text-sm text-muted-foreground">
          Projects group your apps, databases, and cron jobs. Each project
          auto-creates Production + Staging environments.
        </p>
      </div>
      <Button variant="brand" size="sm" onClick={onCreate}>
        <Plus className="h-3.5 w-3.5" />
        Create your first project
      </Button>
    </div>
  );
}

export default Projects;
