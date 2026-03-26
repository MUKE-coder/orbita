import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Plus, Loader2, Trash2, Eye, EyeOff, Upload, Download } from "lucide-react";
import { toast } from "sonner";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
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
} from "@/components/ui/dialog";
import { envVarsApi, type EnvVar } from "@/api/envvars";

interface EnvVarEditorProps {
  orgSlug: string;
  appId: string;
}

function EnvVarEditor({ orgSlug, appId }: EnvVarEditorProps) {
  const queryClient = useQueryClient();
  const [newKey, setNewKey] = useState("");
  const [newValue, setNewValue] = useState("");
  const [newIsSecret, setNewIsSecret] = useState(false);
  const [importOpen, setImportOpen] = useState(false);
  const [importContent, setImportContent] = useState("");
  const [revealedIds, setRevealedIds] = useState<Set<string>>(new Set());

  const { data, isLoading } = useQuery({
    queryKey: ["env-vars", orgSlug, appId],
    queryFn: () => envVarsApi.list(orgSlug, appId),
  });

  const invalidate = () =>
    queryClient.invalidateQueries({ queryKey: ["env-vars", orgSlug, appId] });

  const setMut = useMutation({
    mutationFn: () =>
      envVarsApi.set(orgSlug, appId, { key: newKey, value: newValue, is_secret: newIsSecret }),
    onSuccess: () => {
      invalidate();
      toast.success("Variable added");
      setNewKey("");
      setNewValue("");
      setNewIsSecret(false);
    },
    onError: () => toast.error("Failed to add variable"),
  });

  const deleteMut = useMutation({
    mutationFn: (envId: string) => envVarsApi.delete(orgSlug, appId, envId),
    onSuccess: () => {
      invalidate();
      toast.success("Variable deleted");
    },
  });

  const importMut = useMutation({
    mutationFn: () => envVarsApi.importDotenv(orgSlug, appId, importContent),
    onSuccess: (res) => {
      invalidate();
      toast.success(`Imported ${res.data.data.imported} variables`);
      setImportOpen(false);
      setImportContent("");
    },
    onError: () => toast.error("Failed to import"),
  });

  const toggleReveal = (id: string) => {
    const next = new Set(revealedIds);
    if (next.has(id)) next.delete(id);
    else next.add(id);
    setRevealedIds(next);
  };

  const exportAsEnv = (vars: EnvVar[]) => {
    const content = vars
      .filter((v) => !v.is_secret)
      .map((v) => `${v.key}=${v.value}`)
      .join("\n");
    const blob = new Blob([content], { type: "text/plain" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = ".env";
    a.click();
    URL.revokeObjectURL(url);
  };

  const envVars = data?.data?.data || [];

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-medium">Environment Variables</h3>
        <div className="flex gap-2">
          <Button size="sm" variant="outline" onClick={() => setImportOpen(true)}>
            <Upload className="mr-2 h-3.5 w-3.5" />
            Import .env
          </Button>
          {envVars.length > 0 && (
            <Button size="sm" variant="outline" onClick={() => exportAsEnv(envVars)}>
              <Download className="mr-2 h-3.5 w-3.5" />
              Export
            </Button>
          )}
        </div>
      </div>

      {/* Add new variable */}
      <Card>
        <CardHeader className="pb-2">
          <CardTitle className="text-xs text-muted-foreground">Add Variable</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex gap-2 items-end">
            <div className="flex-1 space-y-1">
              <Label className="text-xs">Key</Label>
              <Input
                placeholder="DATABASE_URL"
                value={newKey}
                onChange={(e) => setNewKey(e.target.value)}
                className="font-mono text-xs"
              />
            </div>
            <div className="flex-1 space-y-1">
              <Label className="text-xs">Value</Label>
              <Input
                type={newIsSecret ? "password" : "text"}
                placeholder="value"
                value={newValue}
                onChange={(e) => setNewValue(e.target.value)}
                className="font-mono text-xs"
              />
            </div>
            <Button
              size="sm"
              variant={newIsSecret ? "default" : "outline"}
              className="h-9 px-2"
              onClick={() => setNewIsSecret(!newIsSecret)}
              title={newIsSecret ? "Marked as secret" : "Mark as secret"}
            >
              {newIsSecret ? <EyeOff className="h-3.5 w-3.5" /> : <Eye className="h-3.5 w-3.5" />}
            </Button>
            <Button
              size="sm"
              onClick={() => setMut.mutate()}
              disabled={!newKey || !newValue || setMut.isPending}
            >
              {setMut.isPending ? <Loader2 className="h-4 w-4 animate-spin" /> : <Plus className="h-4 w-4" />}
            </Button>
          </div>
        </CardContent>
      </Card>

      {/* Variable list */}
      {isLoading ? (
        <div className="flex justify-center py-4">
          <Loader2 className="h-5 w-5 animate-spin text-muted-foreground" />
        </div>
      ) : envVars.length === 0 ? (
        <p className="text-sm text-muted-foreground text-center py-6">
          No environment variables set.
        </p>
      ) : (
        <div className="space-y-1">
          {envVars.map((v) => (
            <div
              key={v.id}
              className="flex items-center gap-2 rounded border px-3 py-2 text-xs font-mono"
            >
              <span className="w-1/3 truncate font-semibold">{v.key}</span>
              <span className="flex-1 truncate text-muted-foreground">
                {v.is_secret && !revealedIds.has(v.id)
                  ? "••••••••"
                  : v.value}
              </span>
              <div className="flex gap-1">
                {v.is_secret && (
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-6 w-6"
                    onClick={() => toggleReveal(v.id)}
                  >
                    {revealedIds.has(v.id) ? (
                      <EyeOff className="h-3 w-3" />
                    ) : (
                      <Eye className="h-3 w-3" />
                    )}
                  </Button>
                )}
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-6 w-6 text-muted-foreground hover:text-destructive"
                  onClick={() => deleteMut.mutate(v.id)}
                >
                  <Trash2 className="h-3 w-3" />
                </Button>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Import Dialog */}
      <Dialog open={importOpen} onOpenChange={setImportOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Import from .env</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <textarea
              className="w-full rounded-md border bg-background px-3 py-2 text-xs font-mono h-48 resize-none"
              placeholder={"DATABASE_URL=postgres://...\nAPI_KEY=sk-...\n# Comments are ignored"}
              value={importContent}
              onChange={(e) => setImportContent(e.target.value)}
            />
            <Button
              className="w-full"
              onClick={() => importMut.mutate()}
              disabled={!importContent || importMut.isPending}
            >
              {importMut.isPending && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Import Variables
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}

export default EnvVarEditor;
