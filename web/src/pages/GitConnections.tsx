import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import {
  Plus,
  Loader2,
  Trash2,
  GitBranch,
  ExternalLink,
  Search,
  CheckCircle2,
  ShieldCheck,
  BookOpen,
  KeyRound,
} from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { PageHelp } from "@/components/layout/PageHelp";
import { gitApi, type GitConnection } from "@/api/git";
import { useOrgStore } from "@/stores/org";
import { cn } from "@/lib/utils";

const PROVIDERS = [
  {
    id: "github",
    name: "GitHub",
    icon: GitHubIcon,
    tokenUrl: "https://github.com/settings/tokens/new?scopes=repo,admin:repo_hook&description=Orbita",
    tokenLabel: "Personal Access Token (classic)",
    scopes: ["repo", "admin:repo_hook"],
    requiresBaseUrl: false,
    placeholder: "ghp_XxxXxxXxxXxxXxxXxxXxxXxxXxx",
  },
  {
    id: "gitlab",
    name: "GitLab",
    icon: GitLabIcon,
    tokenUrl: "https://gitlab.com/-/user_settings/personal_access_tokens",
    tokenLabel: "Personal Access Token",
    scopes: ["api", "read_repository"],
    requiresBaseUrl: true,
    placeholder: "glpat-XxxXxxXxxXxxXxxXxx",
  },
  {
    id: "gitea",
    name: "Gitea",
    icon: GiteaIcon,
    tokenUrl: "/user/settings/applications",
    tokenLabel: "Access Token",
    scopes: ["repository", "admin:repo_hook"],
    requiresBaseUrl: true,
    placeholder: "XxxXxxXxxXxxXxxXxxXxxXxx",
  },
] as const;

type ProviderId = (typeof PROVIDERS)[number]["id"];

const createSchema = z.object({
  provider: z.enum(["github", "gitlab", "gitea"]),
  access_token: z.string().min(10, "Token looks too short"),
  base_url: z.string().url().optional().or(z.literal("")),
});
type CreateForm = z.infer<typeof createSchema>;

export default function GitConnections() {
  const currentOrg = useOrgStore((s) => s.currentOrg);
  const slug = currentOrg?.slug || "";
  const qc = useQueryClient();
  const [dialogOpen, setDialogOpen] = useState(false);
  const [expanded, setExpanded] = useState<string | null>(null);

  const { data, isLoading } = useQuery({
    queryKey: ["git-connections", slug],
    queryFn: () => gitApi.listConnections(slug),
    enabled: !!slug,
  });

  const connections: GitConnection[] = data?.data?.data || [];

  const deleteMutation = useMutation({
    mutationFn: (id: string) => gitApi.deleteConnection(slug, id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["git-connections", slug] });
      toast.success("Connection removed");
    },
    onError: () => toast.error("Failed to remove connection"),
  });

  if (!currentOrg) {
    return (
      <div className="py-12 text-center text-sm text-muted-foreground">
        Select an organization.
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
              Git connections
            </h1>
            <PageHelp
              title="Git connections"
              summary="Link a GitHub, GitLab, or Gitea account so apps can deploy directly from a repository."
              steps={[
                {
                  title: "Create a Personal Access Token",
                  body: "Click Connect provider, then the 'Create token' link — it opens the provider with the correct scopes pre-filled (repo + admin:repo_hook for GitHub).",
                },
                {
                  title: "Paste the token",
                  body: "Orbita encrypts it at rest with your org's AES-256 key. It's never logged or shown in plaintext again.",
                },
                {
                  title: "Browse your repos",
                  body: "Each connection card has a Browse repos button to verify the token works and preview what's available.",
                },
                {
                  title: "Deploy from a repo",
                  body: "Go to New App → Git Repository tab. Connections + repos + branches all come from here.",
                },
              ]}
              nextLinks={[
                {
                  label: "New App",
                  to: `/orgs/${slug}/apps/new`,
                  description: "Deploy a connected repo",
                },
              ]}
            />
          </div>
          <p className="mt-1 text-sm text-muted-foreground">
            Connect your code provider so you can deploy apps straight from a
            repository. Pushes auto-deploy via webhooks.
          </p>
        </div>

        <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
          <DialogTrigger
            render={
              <Button variant="brand" size="sm">
                <Plus className="h-3.5 w-3.5" />
                Connect provider
              </Button>
            }
          />
          <ConnectDialog slug={slug} onClose={() => setDialogOpen(false)} />
        </Dialog>
      </div>

      {/* Feature strip */}
      <div className="grid gap-3 sm:grid-cols-3">
        <Feature
          icon={ShieldCheck}
          title="Encrypted at rest"
          body="Tokens are encrypted with your org's AES-256 key before storage."
        />
        <Feature
          icon={GitBranch}
          title="Auto-deploy on push"
          body="Every push to the connected branch triggers a new deployment."
        />
        <Feature
          icon={BookOpen}
          title="Any repo, private included"
          body="Private GitHub, GitLab, and self-hosted Gitea are all supported."
        />
      </div>

      {/* Connections */}
      {isLoading ? (
        <div className="flex justify-center py-16">
          <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
        </div>
      ) : connections.length === 0 ? (
        <EmptyConnections onConnect={() => setDialogOpen(true)} />
      ) : (
        <div className="space-y-3">
          {connections.map((conn) => (
            <ConnectionCard
              key={conn.id}
              conn={conn}
              slug={slug}
              expanded={expanded === conn.id}
              onToggle={() =>
                setExpanded((prev) => (prev === conn.id ? null : conn.id))
              }
              onDelete={() => {
                if (confirm(`Remove this ${conn.provider} connection?`)) {
                  deleteMutation.mutate(conn.id);
                }
              }}
            />
          ))}
        </div>
      )}
    </div>
  );
}

// -------------- subcomponents --------------

function Feature({
  icon: Icon,
  title,
  body,
}: {
  icon: typeof ShieldCheck;
  title: string;
  body: string;
}) {
  return (
    <div className="rounded-lg border border-border bg-card p-4">
      <div className="flex items-center gap-2">
        <div className="flex h-7 w-7 items-center justify-center rounded-md bg-brand/10 text-brand">
          <Icon className="h-3.5 w-3.5" />
        </div>
        <div className="text-[13px] font-medium">{title}</div>
      </div>
      <p className="mt-2 text-[12px] leading-relaxed text-muted-foreground">
        {body}
      </p>
    </div>
  );
}

function EmptyConnections({ onConnect }: { onConnect: () => void }) {
  return (
    <div className="flex flex-col items-center gap-4 rounded-xl border border-dashed border-border bg-card/50 px-6 py-16 text-center">
      <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-muted">
        <GitBranch className="h-5 w-5 text-muted-foreground" />
      </div>
      <div className="max-w-md">
        <h3 className="font-heading text-base font-semibold">
          No providers connected yet
        </h3>
        <p className="mt-1 text-sm text-muted-foreground">
          Connect GitHub, GitLab, or Gitea to deploy apps from a git repository
          with auto-deploy on push.
        </p>
      </div>
      <Button variant="brand" size="sm" onClick={onConnect}>
        <Plus className="h-3.5 w-3.5" />
        Connect your first provider
      </Button>
    </div>
  );
}

function ConnectionCard({
  conn,
  slug,
  expanded,
  onToggle,
  onDelete,
}: {
  conn: GitConnection;
  slug: string;
  expanded: boolean;
  onToggle: () => void;
  onDelete: () => void;
}) {
  const provider = PROVIDERS.find((p) => p.id === conn.provider);
  const Icon = provider?.icon || GitHubIcon;
  return (
    <div className="overflow-hidden rounded-xl border border-border bg-card shadow-xs">
      <div className="flex flex-wrap items-center gap-4 px-5 py-4">
        <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-muted">
          <Icon className="h-5 w-5" />
        </div>
        <div className="flex min-w-0 flex-1 flex-col">
          <div className="flex items-center gap-2">
            <span className="font-heading text-[15px] font-semibold capitalize">
              {provider?.name || conn.provider}
            </span>
            <Badge variant="success" className="gap-1">
              <CheckCircle2 className="h-2.5 w-2.5" />
              Connected
            </Badge>
          </div>
          <div className="mt-0.5 text-xs text-muted-foreground">
            {conn.metadata?.base_url || `${provider?.name}.com`} ·{" "}
            Added {new Date(conn.created_at).toLocaleDateString()}
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={onToggle}>
            {expanded ? "Hide repos" : "Browse repos"}
          </Button>
          <Button
            variant="ghost"
            size="icon-sm"
            onClick={onDelete}
            aria-label="Remove connection"
          >
            <Trash2 className="h-3.5 w-3.5 text-destructive" />
          </Button>
        </div>
      </div>
      {expanded && (
        <div className="border-t border-border bg-muted/20 px-5 py-4">
          <RepoBrowser slug={slug} connectionId={conn.id} />
        </div>
      )}
    </div>
  );
}

function RepoBrowser({ slug, connectionId }: { slug: string; connectionId: string }) {
  const [query, setQuery] = useState("");

  const { data, isLoading, error } = useQuery({
    queryKey: ["git-repos", slug, connectionId],
    queryFn: () => gitApi.listRepos(slug, connectionId),
    staleTime: 60_000,
  });

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-10 text-sm text-muted-foreground">
        <Loader2 className="mr-2 h-4 w-4 animate-spin" />
        Loading repos...
      </div>
    );
  }

  if (error) {
    return (
      <div className="py-6 text-center text-sm text-destructive">
        Couldn't load repos. Check the token has <code>repo</code> scope.
      </div>
    );
  }

  const repos = data?.data?.data || [];
  const filtered = query
    ? repos.filter((r) => r.full_name.toLowerCase().includes(query.toLowerCase()))
    : repos;

  if (repos.length === 0) {
    return (
      <div className="py-6 text-center text-sm text-muted-foreground">
        No repositories visible with the token's scopes.
      </div>
    );
  }

  return (
    <div className="space-y-3">
      <div className="relative">
        <Search className="pointer-events-none absolute left-3 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-muted-foreground" />
        <Input
          placeholder={`Search ${repos.length} repositories...`}
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          className="pl-9"
        />
      </div>

      <div className="max-h-80 overflow-y-auto rounded-lg border border-border bg-background">
        {filtered.slice(0, 100).map((repo) => (
          <div
            key={repo.full_name}
            className="flex items-center justify-between border-b border-border px-4 py-2.5 text-sm last:border-0 hover:bg-accent/40"
          >
            <div className="flex min-w-0 items-center gap-2.5">
              <BookOpen className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
              <span className="truncate font-mono text-[13px]">
                {repo.full_name}
              </span>
              <Badge variant="outline" className="h-5 shrink-0 px-1.5 text-[10px]">
                {repo.default_branch}
              </Badge>
            </div>
            <a
              href={repo.clone_url.replace(/\.git$/, "")}
              target="_blank"
              rel="noreferrer"
              className="shrink-0 text-muted-foreground hover:text-foreground"
            >
              <ExternalLink className="h-3.5 w-3.5" />
            </a>
          </div>
        ))}
        {filtered.length === 0 && (
          <div className="px-4 py-6 text-center text-xs text-muted-foreground">
            No repos match "{query}"
          </div>
        )}
      </div>

      <div className="flex items-start gap-2 rounded-md bg-muted/50 px-3 py-2 text-xs text-muted-foreground">
        <ShieldCheck className="mt-0.5 h-3.5 w-3.5 shrink-0 text-brand" />
        <span>
          Repos listed here can be deployed via{" "}
          <strong className="text-foreground">Create App → Git Repository</strong>.
          Only your token's scopes control what's visible.
        </span>
      </div>
    </div>
  );
}

function ConnectDialog({ slug, onClose }: { slug: string; onClose: () => void }) {
  const qc = useQueryClient();
  const [provider, setProvider] = useState<ProviderId>("github");
  const providerInfo = PROVIDERS.find((p) => p.id === provider)!;

  const {
    register,
    handleSubmit,
    reset,
    setValue,
    formState: { errors },
  } = useForm<CreateForm>({
    resolver: zodResolver(createSchema),
    defaultValues: { provider: "github" },
  });

  const createMutation = useMutation({
    mutationFn: (data: CreateForm) =>
      gitApi.createConnection(slug, {
        provider: data.provider,
        access_token: data.access_token,
        base_url: data.base_url || undefined,
      }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ["git-connections", slug] });
      toast.success(`${providerInfo.name} connected`);
      reset();
      onClose();
    },
    onError: (err: unknown) => {
      const msg =
        (err as { response?: { data?: { error?: { message?: string } } } })
          ?.response?.data?.error?.message || "Failed to connect";
      toast.error(msg);
    },
  });

  return (
    <DialogContent className="max-w-lg">
      <DialogHeader>
        <DialogTitle>Connect a provider</DialogTitle>
        <DialogDescription>
          Paste a personal access token. We encrypt it with your org's AES-256
          key before storing.
        </DialogDescription>
      </DialogHeader>

      {/* Provider tabs */}
      <div className="grid grid-cols-3 gap-2">
        {PROVIDERS.map((p) => {
          const Icon = p.icon;
          const active = p.id === provider;
          return (
            <button
              key={p.id}
              type="button"
              onClick={() => {
                setProvider(p.id);
                setValue("provider", p.id);
              }}
              className={cn(
                "flex flex-col items-center gap-2 rounded-lg border px-4 py-4 text-sm transition-colors",
                active
                  ? "border-brand bg-brand/5 text-foreground ring-1 ring-brand/20"
                  : "border-border text-muted-foreground hover:border-foreground/20 hover:bg-accent/50"
              )}
            >
              <Icon className="h-6 w-6" />
              <span className="font-medium">{p.name}</span>
            </button>
          );
        })}
      </div>

      <form
        onSubmit={handleSubmit((d) => createMutation.mutate({ ...d, provider }))}
        className="mt-2 space-y-4"
      >
        <input type="hidden" {...register("provider")} value={provider} />

        {providerInfo.requiresBaseUrl && (
          <div className="space-y-1.5">
            <Label htmlFor="base_url">Server URL</Label>
            <Input
              id="base_url"
              placeholder="https://git.yourcompany.com"
              {...register("base_url")}
            />
            {errors.base_url && (
              <p className="text-xs text-destructive">{errors.base_url.message}</p>
            )}
          </div>
        )}

        <div className="space-y-1.5">
          <div className="flex items-center justify-between">
            <Label htmlFor="access_token">{providerInfo.tokenLabel}</Label>
            <a
              href={providerInfo.tokenUrl}
              target="_blank"
              rel="noreferrer"
              className="inline-flex items-center gap-1 text-xs font-medium text-brand hover:underline"
            >
              <KeyRound className="h-3 w-3" />
              Create token
            </a>
          </div>
          <Input
            id="access_token"
            type="password"
            placeholder={providerInfo.placeholder}
            autoComplete="new-password"
            {...register("access_token")}
          />
          {errors.access_token && (
            <p className="text-xs text-destructive">
              {errors.access_token.message}
            </p>
          )}
          <div className="flex flex-wrap gap-1.5 pt-1">
            <span className="text-[11px] text-muted-foreground">
              Required scopes:
            </span>
            {providerInfo.scopes.map((s) => (
              <code
                key={s}
                className="rounded-sm bg-muted px-1.5 py-0.5 font-mono text-[10px] text-foreground"
              >
                {s}
              </code>
            ))}
          </div>
        </div>

        <div className="flex justify-end gap-2 pt-2">
          <Button
            type="button"
            variant="ghost"
            onClick={onClose}
            disabled={createMutation.isPending}
          >
            Cancel
          </Button>
          <Button
            type="submit"
            variant="brand"
            disabled={createMutation.isPending}
          >
            {createMutation.isPending && (
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            )}
            Connect {providerInfo.name}
          </Button>
        </div>
      </form>
    </DialogContent>
  );
}

// -------------- provider SVG icons --------------

function GitHubIcon(props: React.SVGProps<SVGSVGElement>) {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" {...props}>
      <path d="M12 .3a12 12 0 0 0-3.79 23.4c.6.11.82-.26.82-.58l-.02-2.24c-3.34.72-4.04-1.58-4.04-1.58-.55-1.38-1.33-1.75-1.33-1.75-1.09-.74.08-.73.08-.73 1.2.08 1.83 1.23 1.83 1.23 1.07 1.83 2.8 1.3 3.49 1 .1-.77.42-1.3.76-1.6-2.66-.3-5.47-1.33-5.47-5.93 0-1.31.47-2.38 1.23-3.22-.12-.3-.53-1.52.12-3.17 0 0 1-.32 3.3 1.23a11.48 11.48 0 0 1 6 0c2.3-1.55 3.3-1.23 3.3-1.23.65 1.65.24 2.87.12 3.17.77.84 1.23 1.91 1.23 3.22 0 4.61-2.81 5.62-5.48 5.92.43.37.81 1.1.81 2.22l-.01 3.29c0 .32.22.69.83.58A12 12 0 0 0 12 .3" />
    </svg>
  );
}

function GitLabIcon(props: React.SVGProps<SVGSVGElement>) {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" {...props}>
      <path d="M23.955 13.587l-1.342-4.135-2.664-8.189a.459.459 0 0 0-.873 0L16.41 9.452H7.59L4.919 1.263a.459.459 0 0 0-.872 0L1.378 9.452.045 13.587a.916.916 0 0 0 .335 1.026L12 23.054l11.625-8.44a.918.918 0 0 0 .33-1.027" />
    </svg>
  );
}

function GiteaIcon(props: React.SVGProps<SVGSVGElement>) {
  return (
    <svg viewBox="0 0 24 24" fill="currentColor" {...props}>
      <path d="M4.09 0A3.16 3.16 0 0 0 .94 3.16v17.68A3.16 3.16 0 0 0 4.09 24h15.82a3.16 3.16 0 0 0 3.15-3.16V3.16A3.16 3.16 0 0 0 19.91 0zm7.9 5.79h2.94l1.87 3.93h3.3c.42 0 .75.34.75.75v8.56a.75.75 0 0 1-.75.75H4.9a.75.75 0 0 1-.75-.75v-8.56c0-.41.33-.75.75-.75h3.3l1.87-3.93zm1.47 5.29c-1.46 0-2.65 1.19-2.65 2.65s1.19 2.65 2.65 2.65 2.65-1.19 2.65-2.65-1.19-2.65-2.65-2.65z" />
    </svg>
  );
}
