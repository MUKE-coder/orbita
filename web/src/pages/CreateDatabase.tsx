import { useState } from "react";
import { useNavigate } from "react-router-dom";
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
import { databasesApi } from "@/api/databases";
import { projectsApi } from "@/api/projects";
import { useOrgStore } from "@/stores/org";

const engines = [
  { value: "postgres", label: "PostgreSQL", icon: "🐘", versions: ["16", "15"] },
  { value: "mysql", label: "MySQL", icon: "🐬", versions: ["8"] },
  { value: "mariadb", label: "MariaDB", icon: "🦭", versions: ["11", "10"] },
  { value: "mongodb", label: "MongoDB", icon: "🍃", versions: ["7", "6"] },
  { value: "redis", label: "Redis", icon: "🔴", versions: ["7"] },
];

const schema = z.object({
  name: z.string().min(2),
  environment_id: z.string().min(1, "Environment is required"),
});

function CreateDatabase() {
  const navigate = useNavigate();
  const currentOrg = useOrgStore((s) => s.currentOrg);
  const slug = currentOrg?.slug || "";

  const [selectedEngine, setSelectedEngine] = useState("postgres");
  const [selectedVersion, setSelectedVersion] = useState("16");
  const [selectedProject, setSelectedProject] = useState("");

  const { data: projectsData } = useQuery({
    queryKey: ["projects", slug],
    queryFn: () => projectsApi.list(slug),
    enabled: !!slug,
  });

  const projects = projectsData?.data?.data || [];
  const environments = projects.find((p) => p.id === selectedProject)?.environments || [];

  const { register, handleSubmit, setValue, formState: { errors } } = useForm({
    resolver: zodResolver(schema),
    defaultValues: { name: "", environment_id: "" },
  });

  const createMut = useMutation({
    mutationFn: (data: { name: string; environment_id: string }) =>
      databasesApi.create(slug, {
        ...data,
        engine: selectedEngine,
        version: selectedVersion,
      }),
    onSuccess: (res) => {
      toast.success("Database created");
      navigate(`/orgs/${slug}/databases/${res.data.data.id}`);
    },
    onError: () => toast.error("Failed to create database"),
  });

  const engineData = engines.find((e) => e.value === selectedEngine)!;

  if (!currentOrg) return null;

  return (
    <div className="mx-auto max-w-lg space-y-6">
      <h1 className="text-2xl font-bold">New Database</h1>

      {/* Engine selector */}
      <div className="grid grid-cols-5 gap-2">
        {engines.map((eng) => (
          <button
            key={eng.value}
            type="button"
            onClick={() => { setSelectedEngine(eng.value); setSelectedVersion(eng.versions[0]); }}
            className={`flex flex-col items-center gap-1 rounded-lg border p-3 text-xs transition-colors ${
              selectedEngine === eng.value ? "border-primary bg-primary/5" : "hover:border-primary/50"
            }`}
          >
            <span className="text-xl">{eng.icon}</span>
            {eng.label}
          </button>
        ))}
      </div>

      <form onSubmit={handleSubmit((d) => createMut.mutate(d))} className="space-y-4">
        <div className="space-y-2">
          <Label>Version</Label>
          <Select value={selectedVersion} onValueChange={(v) => v && setSelectedVersion(String(v))}>
            <SelectTrigger><SelectValue /></SelectTrigger>
            <SelectContent>
              {engineData.versions.map((v) => (
                <SelectItem key={v} value={v}>{engineData.label} {v}</SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

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

        {selectedProject && (
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
          <Label htmlFor="name">Database Name</Label>
          <Input id="name" placeholder="my-database" {...register("name")} />
          {errors.name && <p className="text-sm text-destructive">{errors.name.message}</p>}
        </div>

        <Button type="submit" className="w-full" disabled={createMut.isPending}>
          {createMut.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
          Create {engineData.label} Database
        </Button>
      </form>
    </div>
  );
}

export default CreateDatabase;
