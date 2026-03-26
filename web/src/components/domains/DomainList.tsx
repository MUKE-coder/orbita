import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Plus, Loader2, Globe, Trash2, ShieldCheck, ShieldAlert } from "lucide-react";
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
import { domainsApi } from "@/api/domains";

const domainSchema = z.object({
  domain: z.string().min(3, "Domain is required"),
});

const statusColor: Record<string, string> = {
  active: "bg-green-500/10 text-green-500",
  pending: "bg-yellow-500/10 text-yellow-500",
  error: "bg-red-500/10 text-red-500",
};

interface DomainListProps {
  orgSlug: string;
  appId: string;
}

function DomainList({ orgSlug, appId }: DomainListProps) {
  const queryClient = useQueryClient();
  const [showForm, setShowForm] = useState(false);

  const { data, isLoading } = useQuery({
    queryKey: ["app-domains", orgSlug, appId],
    queryFn: () => domainsApi.listByApp(orgSlug, appId),
  });

  const { register, handleSubmit, reset, formState: { errors } } = useForm({
    resolver: zodResolver(domainSchema),
    defaultValues: { domain: "" },
  });

  const addMut = useMutation({
    mutationFn: (data: { domain: string }) =>
      domainsApi.addToApp(orgSlug, appId, { domain: data.domain, ssl_enabled: true }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["app-domains", orgSlug, appId] });
      toast.success("Domain added");
      reset();
      setShowForm(false);
    },
    onError: (err: unknown) => {
      const msg = (err as { response?: { data?: { error?: { message?: string } } } })
        ?.response?.data?.error?.message || "Failed to add domain";
      toast.error(msg);
    },
  });

  const removeMut = useMutation({
    mutationFn: (domainId: string) => domainsApi.remove(orgSlug, domainId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["app-domains", orgSlug, appId] });
      toast.success("Domain removed");
    },
  });

  const domains = data?.data?.data || [];

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-medium">Custom Domains</h3>
        <Button size="sm" variant="outline" onClick={() => setShowForm(!showForm)}>
          <Plus className="mr-2 h-4 w-4" />
          Add Domain
        </Button>
      </div>

      {showForm && (
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm">Add Custom Domain</CardTitle>
          </CardHeader>
          <CardContent>
            <form onSubmit={handleSubmit((d) => addMut.mutate(d))} className="flex gap-3 items-end">
              <div className="flex-1 space-y-1">
                <Label htmlFor="domain">Domain</Label>
                <Input id="domain" placeholder="app.example.com" {...register("domain")} />
                {errors.domain && <p className="text-xs text-destructive">{errors.domain.message}</p>}
              </div>
              <Button type="submit" size="sm" disabled={addMut.isPending}>
                {addMut.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
                Add
              </Button>
            </form>
            <p className="mt-2 text-xs text-muted-foreground">
              Point a CNAME record to your server IP or configure an A record.
              SSL will be automatically provisioned via Let's Encrypt.
            </p>
          </CardContent>
        </Card>
      )}

      {isLoading ? (
        <div className="flex justify-center py-4">
          <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
        </div>
      ) : domains.length === 0 ? (
        <p className="text-sm text-muted-foreground py-4 text-center">
          No custom domains configured.
        </p>
      ) : (
        <div className="space-y-2">
          {domains.map((d) => (
            <div key={d.id} className="flex items-center justify-between rounded-lg border p-3">
              <div className="flex items-center gap-3">
                <Globe className="h-4 w-4 text-muted-foreground" />
                <div>
                  <p className="text-sm font-medium">{d.domain}</p>
                  <div className="flex items-center gap-2 mt-0.5">
                    <Badge className={statusColor[d.status] || ""} variant="secondary">
                      {d.status}
                    </Badge>
                    {d.ssl_enabled && (
                      d.verified ? (
                        <span className="flex items-center gap-1 text-xs text-green-500">
                          <ShieldCheck className="h-3 w-3" /> SSL Active
                        </span>
                      ) : (
                        <span className="flex items-center gap-1 text-xs text-yellow-500">
                          <ShieldAlert className="h-3 w-3" /> SSL Pending
                        </span>
                      )
                    )}
                  </div>
                </div>
              </div>
              <Button
                variant="ghost"
                size="icon"
                className="h-8 w-8 text-muted-foreground hover:text-destructive"
                onClick={() => removeMut.mutate(d.id)}
              >
                <Trash2 className="h-4 w-4" />
              </Button>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

export default DomainList;
