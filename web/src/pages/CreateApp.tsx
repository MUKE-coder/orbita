import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useQuery, useMutation } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Loader2, Container, GitBranch } from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { appsApi } from "@/api/apps";
import { projectsApi } from "@/api/projects";
import { useOrgStore } from "@/stores/org";

const createAppSchema = z.object({
  name: z.string().min(2, "Name must be at least 2 characters"),
  environment_id: z.string().min(1, "Environment is required"),
  image: z.string().min(1, "Image is required"),
  port: z.string().optional().transform((v) => (v ? parseInt(v, 10) : undefined)),
  replicas: z.string().optional().transform((v) => (v ? parseInt(v, 10) : 1)),
});

type CreateAppForm = { name: string; environment_id: string; image: string; port?: number; replicas?: number };

function CreateApp() {
  const navigate = useNavigate();
  const currentOrg = useOrgStore((s) => s.currentOrg);
  const [sourceType] = useState<"docker-image" | "git">("docker-image");
  const [selectedProject, setSelectedProject] = useState<string>("");

  const slug = currentOrg?.slug || "";

  const { data: projectsData } = useQuery({
    queryKey: ["projects", slug],
    queryFn: () => projectsApi.list(slug),
    enabled: !!slug,
  });

  const projects = projectsData?.data?.data || [];
  const selectedProjectData = projects.find((p) => p.id === selectedProject);
  const environments = selectedProjectData?.environments || [];

  const {
    register,
    handleSubmit,
    setValue,
    formState: { errors },
  } = useForm({
    resolver: zodResolver(createAppSchema),
    defaultValues: { replicas: "1", name: "", environment_id: "", image: "" },
  });

  const createMutation = useMutation({
    mutationFn: (data: CreateAppForm) =>
      appsApi.create(slug, {
        name: data.name,
        environment_id: data.environment_id,
        image: data.image,
        source_type: sourceType,
        port: data.port,
        replicas: data.replicas || 1,
      }),
    onSuccess: (res) => {
      toast.success("App created");
      navigate(`/orgs/${slug}/apps/${res.data.data.id}`);
    },
    onError: () => toast.error("Failed to create app"),
  });

  if (!currentOrg) return null;

  return (
    <div className="mx-auto max-w-lg space-y-6">
      <h1 className="text-2xl font-bold">New Application</h1>

      {/* Source type selector */}
      <div className="grid grid-cols-2 gap-3">
        <Card
          className={`cursor-pointer transition-colors ${
            sourceType === "docker-image"
              ? "border-primary"
              : "hover:border-primary/50"
          }`}
        >
          <CardHeader className="pb-2">
            <CardTitle className="flex items-center gap-2 text-sm">
              <Container className="h-4 w-4" />
              Docker Image
            </CardTitle>
          </CardHeader>
          <CardContent>
            <CardDescription className="text-xs">
              Deploy from Docker Hub, GHCR, or any registry
            </CardDescription>
          </CardContent>
        </Card>
        <Card className="cursor-not-allowed opacity-50">
          <CardHeader className="pb-2">
            <CardTitle className="flex items-center gap-2 text-sm">
              <GitBranch className="h-4 w-4" />
              Git Repository
            </CardTitle>
          </CardHeader>
          <CardContent>
            <CardDescription className="text-xs">
              Coming in a future update
            </CardDescription>
          </CardContent>
        </Card>
      </div>

      <form
        onSubmit={handleSubmit((d) => createMutation.mutate(d as CreateAppForm))}
        className="space-y-4"
      >
        <div className="space-y-2">
          <Label>Project</Label>
          <Select onValueChange={(v) => setSelectedProject(String(v ?? ""))}>
            <SelectTrigger>
              <SelectValue placeholder="Select a project" />
            </SelectTrigger>
            <SelectContent>
              {projects.map((p) => (
                <SelectItem key={p.id} value={p.id}>
                  {p.emoji} {p.name}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        {selectedProject && (
          <div className="space-y-2">
            <Label>Environment</Label>
            <Select onValueChange={(v) => setValue("environment_id", String(v ?? ""))}>
              <SelectTrigger>
                <SelectValue placeholder="Select an environment" />
              </SelectTrigger>
              <SelectContent>
                {environments.map((env) => (
                  <SelectItem key={env.id} value={env.id}>
                    {env.name} ({env.type})
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            {errors.environment_id && (
              <p className="text-sm text-destructive">
                {errors.environment_id.message}
              </p>
            )}
          </div>
        )}

        <div className="space-y-2">
          <Label htmlFor="name">App Name</Label>
          <Input id="name" placeholder="my-api" {...register("name")} />
          {errors.name && (
            <p className="text-sm text-destructive">{errors.name.message}</p>
          )}
        </div>

        <div className="space-y-2">
          <Label htmlFor="image">Docker Image</Label>
          <Input
            id="image"
            placeholder="nginx:latest"
            {...register("image")}
          />
          {errors.image && (
            <p className="text-sm text-destructive">{errors.image.message}</p>
          )}
        </div>

        <div className="grid grid-cols-2 gap-4">
          <div className="space-y-2">
            <Label htmlFor="port">Port</Label>
            <Input
              id="port"
              type="number"
              placeholder="3000"
              {...register("port")}
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="replicas">Replicas</Label>
            <Input
              id="replicas"
              type="number"
              placeholder="1"
              {...register("replicas")}
            />
          </div>
        </div>

        <Button
          type="submit"
          className="w-full"
          disabled={createMutation.isPending}
        >
          {createMutation.isPending && (
            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
          )}
          Create App
        </Button>
      </form>
    </div>
  );
}

export default CreateApp;
