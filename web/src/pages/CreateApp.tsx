import { useState, useMemo } from "react";
import { useNavigate } from "react-router-dom";
import { useQuery, useMutation } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import {
  Loader2,
  Container,
  GitBranch,
  Rocket,
  Cpu,
  MemoryStick,
  FolderCog,
  Settings2,
  ArrowRight,
  Info,
  ArrowLeft,
} from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { appsApi } from "@/api/apps";
import { projectsApi } from "@/api/projects";
import { gitApi } from "@/api/git";
import { useOrgStore } from "@/stores/org";
import { cn } from "@/lib/utils";

// -------- schemas (two branches; we pick one at submit time) --------

const baseSchema = {
  name: z.string().min(2, "Name must be at least 2 characters"),
  environment_id: z.string().min(1, "Environment is required"),
  port: z.number().int().min(1).max(65535).optional(),
  replicas: z.number().int().min(1).max(100),
  memory_mb: z.number().int().min(0),
  cpu_shares: z.number().int().min(0),
};

const dockerSchema = z.object({
  ...baseSchema,
  image: z.string().min(1, "Image is required (e.g. nginx:alpine)"),
});

const gitSchema = z.object({
  ...baseSchema,
  git_connection_id: z.string().min(1, "Select a git connection"),
  repo_full_name: z.string().min(1, "Select a repository"),
  branch: z.string().min(1, "Select a branch"),
  dockerfile_path: z.string(),
  build_context: z.string(),
});

type DockerForm = z.infer<typeof dockerSchema>;
type GitForm = z.infer<typeof gitSchema>;

export default function CreateApp() {
  const navigate = useNavigate();
  const currentOrg = useOrgStore((s) => s.currentOrg);
  const slug = currentOrg?.slug || "";

  const [source, setSource] = useState<"docker-image" | "git">("docker-image");
  const [selectedProject, setSelectedProject] = useState<string>("");

  const { data: projectsData } = useQuery({
    queryKey: ["projects", slug],
    queryFn: () => projectsApi.list(slug),
    enabled: !!slug,
  });

  const projects = projectsData?.data?.data || [];
  const selectedProjectData = projects.find((p) => p.id === selectedProject);
  const environments = selectedProjectData?.environments || [];

  const createMutation = useMutation({
    mutationFn: (data: Parameters<typeof appsApi.create>[1]) =>
      appsApi.create(slug, data),
    onSuccess: (res) => {
      toast.success("App created. Click Deploy to launch it.");
      navigate(`/orgs/${slug}/apps/${res.data.data.id}`);
    },
    onError: (err: unknown) => {
      const msg =
        (err as { response?: { data?: { error?: { message?: string } } } })
          ?.response?.data?.error?.message || "Failed to create app";
      toast.error(msg);
    },
  });

  if (!currentOrg) return null;

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <button
            onClick={() => navigate(-1)}
            className="mb-2 inline-flex items-center gap-1 text-xs font-medium text-muted-foreground hover:text-foreground"
          >
            <ArrowLeft className="h-3 w-3" />
            Back
          </button>
          <h1 className="font-heading text-2xl font-semibold tracking-tight">
            New application
          </h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Deploy a container from a Docker image or build directly from a git
            repository.
          </p>
        </div>
      </div>

      {/* Source type tabs */}
      <div className="grid grid-cols-2 gap-3">
        <SourceCard
          active={source === "docker-image"}
          icon={Container}
          title="Docker image"
          desc="Deploy from Docker Hub, GHCR, or any registry"
          onClick={() => setSource("docker-image")}
        />
        <SourceCard
          active={source === "git"}
          icon={GitBranch}
          title="Git repository"
          desc="Build on push from GitHub, GitLab, or Gitea"
          onClick={() => setSource("git")}
        />
      </div>

      {/* Project + Environment (shared) */}
      <Section title="Target" description="Where this app will live.">
        <div className="grid gap-4 sm:grid-cols-2">
          <div className="space-y-1.5">
            <Label>Project</Label>
            <Select onValueChange={(v) => setSelectedProject(String(v ?? ""))}>
              <SelectTrigger>
                <SelectValue placeholder="Select a project" />
              </SelectTrigger>
              <SelectContent>
                {projects.map((p) => (
                  <SelectItem key={p.id} value={p.id}>
                    {p.emoji || "📦"} {p.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-1.5">
            <Label>Environment</Label>
            <Select disabled={!selectedProject}>
              <SelectTrigger>
                <SelectValue
                  placeholder={
                    selectedProject ? "Select an environment" : "Pick a project first"
                  }
                />
              </SelectTrigger>
              <SelectContent>
                {environments.map((e) => (
                  <SelectItem key={e.id} value={e.id}>
                    {e.name} ({e.type})
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </div>
      </Section>

      {/* Source-specific forms */}
      {source === "docker-image" ? (
        <DockerForm
          slug={slug}
          environments={environments}
          onSubmit={(d) =>
            createMutation.mutate({
              ...d,
              source_type: "docker-image",
            })
          }
          isSubmitting={createMutation.isPending}
        />
      ) : (
        <GitSourceForm
          slug={slug}
          environments={environments}
          onSubmit={(d) =>
            createMutation.mutate({
              ...d,
              source_type: "git",
            })
          }
          isSubmitting={createMutation.isPending}
        />
      )}
    </div>
  );
}

// -------- Docker image form --------

function DockerForm({
  environments,
  onSubmit,
  isSubmitting,
}: {
  slug: string;
  environments: { id: string; name: string; type: string }[];
  onSubmit: (data: DockerForm) => void;
  isSubmitting: boolean;
}) {
  const {
    register,
    handleSubmit,
    setValue,
    formState: { errors },
  } = useForm<DockerForm>({
    resolver: zodResolver(dockerSchema),
    defaultValues: { replicas: 1, memory_mb: 0, cpu_shares: 0 },
  });

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
      <Section title="Image" description="The container image to run.">
        <div className="space-y-1.5">
          <Label htmlFor="name">App name</Label>
          <Input id="name" placeholder="my-api" {...register("name")} />
          {errors.name && <Err msg={errors.name.message} />}
        </div>

        <div className="space-y-1.5">
          <Label htmlFor="image">Docker image</Label>
          <Input
            id="image"
            placeholder="nginx:alpine or ghcr.io/you/app:v1.2.3"
            className="font-mono"
            {...register("image")}
          />
          {errors.image && <Err msg={errors.image.message} />}
        </div>

        <input type="hidden" {...register("environment_id")} />
        <EnvSelect
          environments={environments}
          onChange={(v) => setValue("environment_id", v)}
        />
        {errors.environment_id && <Err msg={errors.environment_id.message} />}
      </Section>

      <RuntimeSection register={register} errors={errors} />

      <Submit isSubmitting={isSubmitting} />
    </form>
  );
}

// -------- Git source form --------

function GitSourceForm({
  slug,
  environments,
  onSubmit,
  isSubmitting,
}: {
  slug: string;
  environments: { id: string; name: string; type: string }[];
  onSubmit: (data: GitForm & { repo_url?: string }) => void;
  isSubmitting: boolean;
}) {
  const {
    register,
    handleSubmit,
    setValue,
    watch,
    formState: { errors },
  } = useForm<GitForm>({
    resolver: zodResolver(gitSchema),
    defaultValues: {
      replicas: 1,
      memory_mb: 0,
      cpu_shares: 0,
      dockerfile_path: "Dockerfile",
      build_context: "",
    },
  });

  const connId = watch("git_connection_id");
  const repoFullName = watch("repo_full_name");

  const { data: connsData } = useQuery({
    queryKey: ["git-connections", slug],
    queryFn: () => gitApi.listConnections(slug),
    enabled: !!slug,
  });
  const conns = connsData?.data?.data || [];

  const { data: reposData, isLoading: loadingRepos } = useQuery({
    queryKey: ["git-repos", slug, connId],
    queryFn: () => gitApi.listRepos(slug, connId!),
    enabled: !!connId,
  });
  const repos = reposData?.data?.data || [];

  const [owner, repo] = useMemo(() => {
    if (!repoFullName) return ["", ""];
    const [o, r] = repoFullName.split("/");
    return [o || "", r || ""];
  }, [repoFullName]);

  const { data: branchesData, isLoading: loadingBranches } = useQuery({
    queryKey: ["git-branches", slug, connId, owner, repo],
    queryFn: () => gitApi.listBranches(slug, connId!, owner, repo),
    enabled: !!connId && !!owner && !!repo,
  });
  const branches = branchesData?.data?.data || [];

  const selectedRepo = repos.find((r) => r.full_name === repoFullName);

  const submit = handleSubmit((d) => {
    const data = d as GitForm;
    const cloneURL = selectedRepo?.clone_url || `https://github.com/${data.repo_full_name}.git`;
    onSubmit({ ...data, repo_url: cloneURL });
  });

  if (conns.length === 0) {
    return (
      <div className="flex flex-col items-center gap-3 rounded-xl border border-dashed border-border bg-card/50 px-6 py-12 text-center">
        <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-muted">
          <GitBranch className="h-5 w-5 text-muted-foreground" />
        </div>
        <p className="text-sm text-muted-foreground">
          No git connections yet. Add one first.
        </p>
        <a href={`/orgs/${slug}/git`}>
          <Button variant="brand" size="sm">
            Connect a provider
            <ArrowRight className="h-3 w-3" />
          </Button>
        </a>
      </div>
    );
  }

  return (
    <form onSubmit={submit} className="space-y-6">
      <Section
        title="Repository"
        description="Pick the git connection, repo, and branch to build from."
      >
        <div className="space-y-1.5">
          <Label>Git connection</Label>
          <Select onValueChange={(v) => setValue("git_connection_id", String(v))}>
            <SelectTrigger>
              <SelectValue placeholder="Select a connection" />
            </SelectTrigger>
            <SelectContent>
              {conns.map((c) => (
                <SelectItem key={c.id} value={c.id}>
                  <span className="capitalize">{c.provider}</span>
                  {" · "}
                  {new Date(c.created_at).toLocaleDateString()}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          {errors.git_connection_id && <Err msg={errors.git_connection_id.message} />}
        </div>

        <div className="space-y-1.5">
          <Label>Repository</Label>
          <Select
            disabled={!connId}
            onValueChange={(v) => setValue("repo_full_name", String(v))}
          >
            <SelectTrigger>
              <SelectValue
                placeholder={
                  loadingRepos
                    ? "Loading repositories..."
                    : !connId
                    ? "Pick a connection first"
                    : repos.length === 0
                    ? "No repos visible with this token"
                    : "Select a repository"
                }
              />
            </SelectTrigger>
            <SelectContent>
              {repos.slice(0, 200).map((r) => (
                <SelectItem key={r.full_name} value={r.full_name}>
                  {r.full_name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          {errors.repo_full_name && <Err msg={errors.repo_full_name.message} />}
        </div>

        <div className="space-y-1.5">
          <Label>Branch</Label>
          <Select
            disabled={!repoFullName}
            onValueChange={(v) => setValue("branch", String(v))}
          >
            <SelectTrigger>
              <SelectValue
                placeholder={
                  loadingBranches
                    ? "Loading branches..."
                    : !repoFullName
                    ? "Pick a repo first"
                    : "Select a branch"
                }
              />
            </SelectTrigger>
            <SelectContent>
              {branches.map((b) => (
                <SelectItem key={b} value={b}>
                  {b}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
          {errors.branch && <Err msg={errors.branch.message} />}
        </div>

        <div className="grid gap-4 sm:grid-cols-2">
          <div className="space-y-1.5">
            <Label htmlFor="dockerfile_path">Dockerfile path</Label>
            <Input
              id="dockerfile_path"
              placeholder="Dockerfile"
              className="font-mono"
              {...register("dockerfile_path")}
            />
            <p className="text-[11px] text-muted-foreground">
              Relative to repo root or build context.
            </p>
          </div>
          <div className="space-y-1.5">
            <Label htmlFor="build_context">Build context</Label>
            <Input
              id="build_context"
              placeholder="(repo root)"
              className="font-mono"
              {...register("build_context")}
            />
            <p className="text-[11px] text-muted-foreground">
              Subdirectory for monorepos (optional, e.g. <code>backend</code>).
            </p>
          </div>
        </div>

        <div className="flex items-start gap-2 rounded-md border border-border bg-muted/40 px-3 py-2 text-xs text-muted-foreground">
          <Info className="mt-0.5 h-3.5 w-3.5 shrink-0 text-brand" />
          <span>
            Orbita clones + builds via Docker BuildKit. Your PAT stays encrypted
            at rest and is injected into the clone URL only at build time.
          </span>
        </div>
      </Section>

      <Section title="App" description="How it's identified.">
        <div className="space-y-1.5">
          <Label htmlFor="name">App name</Label>
          <Input id="name" placeholder="my-api" {...register("name")} />
          {errors.name && <Err msg={errors.name.message} />}
        </div>

        <input type="hidden" {...register("environment_id")} />
        <EnvSelect
          environments={environments}
          onChange={(v) => setValue("environment_id", v)}
        />
        {errors.environment_id && <Err msg={errors.environment_id.message} />}
      </Section>

      <RuntimeSection register={register} errors={errors} />

      <Submit isSubmitting={isSubmitting} />
    </form>
  );
}

// -------- shared bits --------

function Section({
  title,
  description,
  children,
}: {
  title: string;
  description?: string;
  children: React.ReactNode;
}) {
  return (
    <section className="overflow-hidden rounded-xl border border-border bg-card shadow-xs">
      <header className="border-b border-border bg-muted/30 px-5 py-3">
        <h2 className="font-heading text-sm font-semibold tracking-tight">{title}</h2>
        {description && (
          <p className="mt-0.5 text-xs text-muted-foreground">{description}</p>
        )}
      </header>
      <div className="space-y-4 px-5 py-4">{children}</div>
    </section>
  );
}

function SourceCard({
  active,
  icon: Icon,
  title,
  desc,
  onClick,
}: {
  active: boolean;
  icon: typeof Container;
  title: string;
  desc: string;
  onClick: () => void;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        "flex flex-col gap-2 rounded-xl border p-5 text-left transition-colors",
        active
          ? "border-brand bg-brand/5 ring-1 ring-brand/20"
          : "border-border bg-card hover:border-foreground/20 hover:bg-accent/50"
      )}
    >
      <div
        className={cn(
          "flex h-9 w-9 items-center justify-center rounded-lg",
          active ? "bg-brand/15 text-brand" : "bg-muted text-muted-foreground"
        )}
      >
        <Icon className="h-4 w-4" />
      </div>
      <div className="font-heading text-[15px] font-semibold tracking-tight">
        {title}
      </div>
      <div className="text-xs text-muted-foreground">{desc}</div>
    </button>
  );
}

function EnvSelect({
  environments,
  onChange,
}: {
  environments: { id: string; name: string; type: string }[];
  onChange: (v: string) => void;
}) {
  return (
    <div className="space-y-1.5">
      <Label>Environment</Label>
      <Select
        disabled={environments.length === 0}
        onValueChange={(v) => onChange(String(v))}
      >
        <SelectTrigger>
          <SelectValue
            placeholder={
              environments.length === 0 ? "Pick a project above first" : "Select an environment"
            }
          />
        </SelectTrigger>
        <SelectContent>
          {environments.map((e) => (
            <SelectItem key={e.id} value={e.id}>
              {e.name} ({e.type})
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
    </div>
  );
}

/* eslint-disable @typescript-eslint/no-explicit-any */
function RuntimeSection({
  register,
  errors,
}: {
  register: any;
  errors: any;
}) {
  return (
    <Section title="Runtime" description="Network + resource limits per container.">
      <div className="grid gap-4 sm:grid-cols-2">
        <div className="space-y-1.5">
          <Label htmlFor="port">
            <FolderCog className="mr-1 inline h-3 w-3 text-muted-foreground" /> Port
          </Label>
          <Input
            id="port"
            type="number"
            placeholder="3000"
            {...register("port", { valueAsNumber: true })}
          />
          <p className="text-[11px] text-muted-foreground">
            The port your container listens on.
          </p>
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="replicas">
            <Settings2 className="mr-1 inline h-3 w-3 text-muted-foreground" /> Replicas
          </Label>
          <Input
            id="replicas"
            type="number"
            placeholder="1"
            {...register("replicas", { valueAsNumber: true })}
          />
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="memory_mb">
            <MemoryStick className="mr-1 inline h-3 w-3 text-muted-foreground" /> Memory limit
          </Label>
          <div className="relative">
            <Input
              id="memory_mb"
              type="number"
              placeholder="0 = unlimited"
              className="pr-12"
              {...register("memory_mb", { valueAsNumber: true })}
            />
            <span className="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-xs text-muted-foreground">
              MB
            </span>
          </div>
        </div>
        <div className="space-y-1.5">
          <Label htmlFor="cpu_shares">
            <Cpu className="mr-1 inline h-3 w-3 text-muted-foreground" /> CPU limit
          </Label>
          <div className="relative">
            <Input
              id="cpu_shares"
              type="number"
              placeholder="0 = unlimited"
              className="pr-16"
              {...register("cpu_shares", { valueAsNumber: true })}
            />
            <span className="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-[10px] text-muted-foreground">
              ×1/1000 cores
            </span>
          </div>
          <p className="text-[11px] text-muted-foreground">1000 = 1 core</p>
        </div>
      </div>
      {errors.replicas && <Err msg={errors.replicas.message} />}
    </Section>
  );
}

function Submit({ isSubmitting }: { isSubmitting: boolean }) {
  return (
    <div className="flex justify-end pt-2">
      <Button type="submit" variant="brand" size="lg" disabled={isSubmitting}>
        {isSubmitting && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
        <Rocket className="h-4 w-4" />
        Create app
      </Button>
    </div>
  );
}

function Err({ msg }: { msg?: string }) {
  if (!msg) return null;
  return <p className="text-xs text-destructive">{msg}</p>;
}
