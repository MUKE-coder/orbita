import { useEffect, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { useQuery, useMutation } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Loader2 } from "lucide-react";
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
import { cronApi } from "@/api/cron";
import { projectsApi } from "@/api/projects";
import { useOrgStore } from "@/stores/org";

const schedulePresets = [
  { label: "Every minute", value: "* * * * *" },
  { label: "Every hour", value: "@hourly" },
  { label: "Every day at midnight", value: "@daily" },
  { label: "Every week", value: "@weekly" },
  { label: "Every month", value: "@monthly" },
];

const schema = z.object({
  name: z.string().min(2),
  schedule: z.string().min(1, "Schedule is required"),
  image: z.string().min(1, "Image is required"),
  command: z.string().optional(),
  environment_id: z.string().min(1, "Environment is required"),
});

function CreateCron() {
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const currentOrg = useOrgStore((s) => s.currentOrg);
  const slug = currentOrg?.slug || "";

  const lockedProjectId = searchParams.get("project") || "";
  const lockedEnvId = searchParams.get("env") || "";
  const isQuickCreate = !!lockedProjectId;

  const [selectedProject, setSelectedProject] = useState(lockedProjectId);
  const [concurrency, setConcurrency] = useState("forbid");

  const { data: projectsData } = useQuery({
    queryKey: ["projects", slug],
    queryFn: () => projectsApi.list(slug),
    enabled: !!slug,
  });

  const projects = projectsData?.data?.data || [];
  const selectedProjectData = projects.find((p) => p.id === selectedProject);
  const environments = selectedProjectData?.environments || [];

  const { register, handleSubmit, setValue, formState: { errors } } = useForm({
    resolver: zodResolver(schema),
    defaultValues: { name: "", schedule: "", image: "", command: "", environment_id: lockedEnvId || "" },
  });

  useEffect(() => {
    if (lockedEnvId) setValue("environment_id", lockedEnvId);
  }, [lockedEnvId, setValue]);

  const createMut = useMutation({
    mutationFn: (data: { name: string; schedule: string; image: string; command?: string; environment_id: string }) =>
      cronApi.create(slug, { ...data, concurrency_policy: concurrency }),
    onSuccess: (res) => {
      toast.success("Cron job created");
      navigate(`/orgs/${slug}/cron-jobs/${res.data.data.id}`);
    },
    onError: () => toast.error("Failed to create cron job"),
  });

  if (!currentOrg) return null;

  return (
    <div className="mx-auto max-w-lg space-y-6">
      <h1 className="text-2xl font-bold">New Cron Job</h1>

      <form onSubmit={handleSubmit((d) => createMut.mutate(d))} className="space-y-4">
        <div className="space-y-2">
          <Label htmlFor="name">Name</Label>
          <Input id="name" placeholder="nightly-cleanup" {...register("name")} />
          {errors.name && <p className="text-sm text-destructive">{errors.name.message}</p>}
        </div>

        <div className="space-y-2">
          <Label>Schedule</Label>
          <div className="flex flex-wrap gap-2 mb-2">
            {schedulePresets.map((preset) => (
              <Button
                key={preset.value}
                type="button"
                variant="outline"
                size="sm"
                className="text-xs"
                onClick={() => setValue("schedule", preset.value)}
              >
                {preset.label}
              </Button>
            ))}
          </div>
          <Input placeholder="* * * * * or @daily" {...register("schedule")} />
          {errors.schedule && <p className="text-sm text-destructive">{errors.schedule.message}</p>}
        </div>

        {isQuickCreate && selectedProjectData ? (
          <div className="flex items-center gap-3 rounded-lg border border-brand/20 bg-brand/5 px-3 py-2 text-sm">
            <span className="text-base">{selectedProjectData.emoji || "📦"}</span>
            <div className="min-w-0 flex-1 truncate">
              <span className="text-muted-foreground">In</span>{" "}
              <span className="font-medium text-foreground">{selectedProjectData.name}</span>
            </div>
          </div>
        ) : (
          <div className="space-y-2">
            <Label>Project</Label>
            <Select onValueChange={(v) => v && setSelectedProject(String(v))}>
              <SelectTrigger><SelectValue placeholder="Select a project" /></SelectTrigger>
              <SelectContent>
                {projects.map((p) => (
                  <SelectItem key={p.id} value={p.id}>{p.emoji} {p.name}</SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        )}

        {!isQuickCreate && selectedProject && (
          <div className="space-y-2">
            <Label>Environment</Label>
            <Select onValueChange={(v) => v && setValue("environment_id", String(v))}>
              <SelectTrigger><SelectValue placeholder="Select environment" /></SelectTrigger>
              <SelectContent>
                {environments.map((env) => (
                  <SelectItem key={env.id} value={env.id}>{env.name}</SelectItem>
                ))}
              </SelectContent>
            </Select>
            {errors.environment_id && <p className="text-sm text-destructive">{errors.environment_id.message}</p>}
          </div>
        )}

        <div className="space-y-2">
          <Label htmlFor="image">Docker Image</Label>
          <Input id="image" placeholder="alpine:latest" {...register("image")} />
          {errors.image && <p className="text-sm text-destructive">{errors.image.message}</p>}
        </div>

        <div className="space-y-2">
          <Label htmlFor="command">Command (optional)</Label>
          <Input id="command" placeholder="sh -c 'echo hello'" {...register("command")} />
        </div>

        <div className="space-y-2">
          <Label>Concurrency Policy</Label>
          <Select value={concurrency} onValueChange={(v) => v && setConcurrency(String(v))}>
            <SelectTrigger><SelectValue /></SelectTrigger>
            <SelectContent>
              <SelectItem value="forbid">Forbid (skip if running)</SelectItem>
              <SelectItem value="allow">Allow (run in parallel)</SelectItem>
              <SelectItem value="replace">Replace (kill previous)</SelectItem>
            </SelectContent>
          </Select>
        </div>

        <Button type="submit" className="w-full" disabled={createMut.isPending}>
          {createMut.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
          Create Cron Job
        </Button>
      </form>
    </div>
  );
}

export default CreateCron;
