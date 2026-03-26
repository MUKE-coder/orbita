import { useState } from "react";
import { useQuery, useMutation } from "@tanstack/react-query";
import { useNavigate } from "react-router-dom";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Loader2, Rocket } from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Badge } from "@/components/ui/badge";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { servicesApi, type Template } from "@/api/services";
import { projectsApi } from "@/api/projects";
import { useOrgStore } from "@/stores/org";

const categoryLabels: Record<string, string> = {
  cms: "CMS",
  analytics: "Analytics",
  monitoring: "Monitoring",
  automation: "Automation",
  storage: "Storage",
  devtools: "Dev Tools",
  security: "Security",
  other: "Other",
};

const nameSchema = z.object({
  name: z.string().min(2),
  environment_id: z.string().min(1, "Required"),
});

function Marketplace() {
  const currentOrg = useOrgStore((s) => s.currentOrg);
  const navigate = useNavigate();
  const slug = currentOrg?.slug || "";

  const [selectedTemplate, setSelectedTemplate] = useState<Template | null>(null);
  const [deployOpen, setDeployOpen] = useState(false);
  const [selectedProject, setSelectedProject] = useState("");
  const [params, setParams] = useState<Record<string, string>>({});
  const [categoryFilter, setCategoryFilter] = useState<string>("all");

  const { data: templatesData, isLoading } = useQuery({
    queryKey: ["templates"],
    queryFn: () => servicesApi.listTemplates(),
  });

  const { data: projectsData } = useQuery({
    queryKey: ["projects", slug],
    queryFn: () => projectsApi.list(slug),
    enabled: !!slug,
  });

  const templates = templatesData?.data?.data || [];
  const projects = projectsData?.data?.data || [];
  const environments = projects.find((p) => p.id === selectedProject)?.environments || [];

  const categories = [...new Set(templates.map((t) => t.category))];
  const filtered = categoryFilter === "all" ? templates : templates.filter((t) => t.category === categoryFilter);

  const { register, handleSubmit, setValue, formState: { errors } } = useForm({
    resolver: zodResolver(nameSchema),
    defaultValues: { name: "", environment_id: "" },
  });

  const deployMut = useMutation({
    mutationFn: (data: { name: string; environment_id: string }) =>
      servicesApi.deployService(slug, {
        template_id: selectedTemplate!.id,
        name: data.name,
        environment_id: data.environment_id,
        params,
      }),
    onSuccess: (res) => {
      toast.success("Service deployed!");
      setDeployOpen(false);
      navigate(`/orgs/${slug}/services/${res.data.data.id}`);
    },
    onError: () => toast.error("Failed to deploy service"),
  });

  const openDeploy = (tmpl: Template) => {
    setSelectedTemplate(tmpl);
    const defaults: Record<string, string> = {};
    tmpl.params_schema?.forEach((p) => {
      if (p.default) defaults[p.name] = p.default;
    });
    setParams(defaults);
    setValue("name", tmpl.name.toLowerCase().replace(/\s+/g, "-"));
    setDeployOpen(true);
  };

  if (!currentOrg) {
    return <p className="text-muted-foreground text-center py-12">Select an organization first.</p>;
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-xl font-semibold">Service Marketplace</h2>
          <p className="text-sm text-muted-foreground">Deploy pre-configured services with one click</p>
        </div>
      </div>

      {/* Category filter */}
      <div className="flex gap-2 flex-wrap">
        <Button size="sm" variant={categoryFilter === "all" ? "default" : "outline"} onClick={() => setCategoryFilter("all")}>
          All
        </Button>
        {categories.map((cat) => (
          <Button key={cat} size="sm" variant={categoryFilter === cat ? "default" : "outline"} onClick={() => setCategoryFilter(cat)}>
            {categoryLabels[cat] || cat}
          </Button>
        ))}
      </div>

      {isLoading ? (
        <div className="flex justify-center py-12"><Loader2 className="h-6 w-6 animate-spin text-muted-foreground" /></div>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {filtered.map((tmpl) => (
            <Card key={tmpl.id} className="flex flex-col">
              <CardHeader className="pb-2">
                <CardTitle className="flex items-center gap-2 text-base">
                  {tmpl.name}
                </CardTitle>
                <Badge variant="secondary" className="w-fit text-xs">{categoryLabels[tmpl.category] || tmpl.category}</Badge>
              </CardHeader>
              <CardContent className="flex-1 flex flex-col justify-between gap-3">
                <CardDescription className="text-xs line-clamp-2">{tmpl.description}</CardDescription>
                <Button size="sm" onClick={() => openDeploy(tmpl)}>
                  <Rocket className="mr-2 h-4 w-4" />
                  Deploy
                </Button>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {/* Deploy dialog */}
      <Dialog open={deployOpen} onOpenChange={setDeployOpen}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>Deploy {selectedTemplate?.name}</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleSubmit((d) => deployMut.mutate(d))} className="space-y-4">
            <div className="space-y-2">
              <Label>Service Name</Label>
              <Input {...register("name")} />
              {errors.name && <p className="text-xs text-destructive">{errors.name.message}</p>}
            </div>

            <div className="space-y-2">
              <Label>Project</Label>
              <Select onValueChange={(v) => v && setSelectedProject(String(v))}>
                <SelectTrigger><SelectValue placeholder="Select project" /></SelectTrigger>
                <SelectContent>
                  {projects.map((p) => <SelectItem key={p.id} value={p.id}>{p.emoji} {p.name}</SelectItem>)}
                </SelectContent>
              </Select>
            </div>

            {selectedProject && (
              <div className="space-y-2">
                <Label>Environment</Label>
                <Select onValueChange={(v) => v && setValue("environment_id", String(v))}>
                  <SelectTrigger><SelectValue placeholder="Select environment" /></SelectTrigger>
                  <SelectContent>
                    {environments.map((e) => <SelectItem key={e.id} value={e.id}>{e.name}</SelectItem>)}
                  </SelectContent>
                </Select>
                {errors.environment_id && <p className="text-xs text-destructive">{errors.environment_id.message}</p>}
              </div>
            )}

            {/* Template params */}
            {selectedTemplate?.params_schema?.map((param) => (
              <div key={param.name} className="space-y-1">
                <Label className="text-xs">{param.name}</Label>
                <Input
                  type={param.type === "password" ? "password" : "text"}
                  value={params[param.name] || ""}
                  onChange={(e) => setParams({ ...params, [param.name]: e.target.value })}
                  placeholder={param.default || `Enter ${param.name}`}
                />
              </div>
            ))}

            <Button type="submit" className="w-full" disabled={deployMut.isPending}>
              {deployMut.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Deploy Service
            </Button>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}

export default Marketplace;
