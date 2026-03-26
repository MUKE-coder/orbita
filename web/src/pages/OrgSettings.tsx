import { ReactNode } from "react";
import { useParams, Link, useLocation } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { Settings, Users, ArrowLeft } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import { orgsApi } from "@/api/orgs";
import OrgMembers from "./OrgMembers";

function OrgSettings({ children }: { children?: ReactNode }) {
  const { orgSlug } = useParams<{ orgSlug: string }>();
  const location = useLocation();

  const { data } = useQuery({
    queryKey: ["org", orgSlug],
    queryFn: () => orgsApi.get(orgSlug!),
    enabled: !!orgSlug,
  });

  const org = data?.data?.data?.organization;
  const isMembers = location.pathname.includes("/members");

  return (
    <div className="flex min-h-screen bg-background">
      {/* Settings Sidebar */}
      <aside className="w-64 border-r p-4 space-y-4">
        <div>
          <Link
            to="/"
            className="flex items-center gap-2 text-sm text-muted-foreground hover:text-foreground"
          >
            <ArrowLeft className="h-4 w-4" />
            Back to dashboard
          </Link>
        </div>
        <div>
          <h2 className="text-lg font-semibold">{org?.name || "Settings"}</h2>
          <p className="text-xs text-muted-foreground">{orgSlug}</p>
        </div>
        <Separator />
        <nav className="space-y-1">
          <Link to={`/orgs/${orgSlug}/settings`}>
            <Button
              variant={!isMembers ? "secondary" : "ghost"}
              className="w-full justify-start"
              size="sm"
            >
              <Settings className="mr-2 h-4 w-4" />
              General
            </Button>
          </Link>
          <Link to={`/orgs/${orgSlug}/settings/members`}>
            <Button
              variant={isMembers ? "secondary" : "ghost"}
              className="w-full justify-start"
              size="sm"
            >
              <Users className="mr-2 h-4 w-4" />
              Members
            </Button>
          </Link>
        </nav>
      </aside>

      {/* Content */}
      <main className="flex-1 p-8">
        {children || (isMembers ? <OrgMembers /> : <GeneralSettings />)}
      </main>
    </div>
  );
}

function GeneralSettings() {
  const { orgSlug } = useParams<{ orgSlug: string }>();

  const { data } = useQuery({
    queryKey: ["org", orgSlug],
    queryFn: () => orgsApi.get(orgSlug!),
    enabled: !!orgSlug,
  });

  const org = data?.data?.data?.organization;
  const plan = org?.plan;

  return (
    <div className="max-w-lg space-y-6">
      <h2 className="text-xl font-semibold">General Settings</h2>

      <div className="space-y-2 rounded-lg border p-4">
        <div className="flex justify-between text-sm">
          <span className="text-muted-foreground">Name</span>
          <span>{org?.name}</span>
        </div>
        <div className="flex justify-between text-sm">
          <span className="text-muted-foreground">Slug</span>
          <span className="font-mono text-xs">{org?.slug}</span>
        </div>
        <div className="flex justify-between text-sm">
          <span className="text-muted-foreground">Plan</span>
          <span>{plan?.name || "None"}</span>
        </div>
        {plan && (
          <>
            <Separator />
            <div className="grid grid-cols-2 gap-2 text-xs text-muted-foreground">
              <span>CPU: {plan.max_cpu_cores} cores</span>
              <span>RAM: {plan.max_ram_mb} MB</span>
              <span>Disk: {plan.max_disk_gb} GB</span>
              <span>Apps: {plan.max_apps}</span>
              <span>Databases: {plan.max_databases}</span>
            </div>
          </>
        )}
      </div>
    </div>
  );
}

export default OrgSettings;
