import { ReactNode } from "react";
import { Link, useNavigate } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import { LogOut, Settings, Users, Rocket, FolderKanban, Plus, Store, Shield } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import OrgSwitcher from "@/components/layout/OrgSwitcher";
import NotificationBell from "@/components/layout/NotificationBell";
import { useAuthStore } from "@/stores/auth";
import { useOrgStore } from "@/stores/org";
import { projectsApi } from "@/api/projects";
import Projects from "./Projects";

function Dashboard({ children }: { children?: ReactNode }) {
  const navigate = useNavigate();
  const logout = useAuthStore((s) => s.logout);
  const currentOrg = useOrgStore((s) => s.currentOrg);

  const { data: projectsData } = useQuery({
    queryKey: ["projects", currentOrg?.slug],
    queryFn: () => projectsApi.list(currentOrg!.slug),
    enabled: !!currentOrg,
  });

  const projects = projectsData?.data?.data || [];

  const handleLogout = () => {
    logout();
    navigate("/login");
  };

  return (
    <div className="flex min-h-screen bg-background">
      {/* Sidebar */}
      <aside className="flex w-64 flex-col border-r">
        <div className="p-4">
          <div className="flex items-center gap-2 text-lg font-semibold">
            <Rocket className="h-5 w-5" />
            Orbita
          </div>
          <p className="text-xs text-muted-foreground">v0.1.0</p>
        </div>

        <Separator />

        <div className="p-2">
          <OrgSwitcher />
        </div>

        <Separator />

        <nav className="flex-1 overflow-y-auto p-2 space-y-1">
          {currentOrg && (
            <>
              <Link to="/">
                <Button variant="ghost" className="w-full justify-start" size="sm">
                  <FolderKanban className="mr-2 h-4 w-4" />
                  Projects
                </Button>
              </Link>

              {/* Project list in sidebar */}
              {projects.length > 0 && (
                <div className="ml-4 space-y-0.5">
                  {projects.map((p) => (
                    <Link
                      key={p.id}
                      to={`/orgs/${currentOrg.slug}/projects/${p.id}`}
                    >
                      <Button
                        variant="ghost"
                        className="w-full justify-start text-xs h-7"
                        size="sm"
                      >
                        <span className="mr-1.5">{p.emoji || "🚀"}</span>
                        <span className="truncate">{p.name}</span>
                      </Button>
                    </Link>
                  ))}
                </div>
              )}

              <Separator className="my-2" />

              <Link to={`/orgs/${currentOrg.slug}/apps/new`}>
                <Button variant="ghost" className="w-full justify-start" size="sm">
                  <Plus className="mr-2 h-4 w-4" />
                  New App
                </Button>
              </Link>
              <Link to={`/orgs/${currentOrg.slug}/databases/new`}>
                <Button variant="ghost" className="w-full justify-start" size="sm">
                  <Plus className="mr-2 h-4 w-4" />
                  New Database
                </Button>
              </Link>
              <Link to={`/orgs/${currentOrg.slug}/cron-jobs/new`}>
                <Button variant="ghost" className="w-full justify-start" size="sm">
                  <Plus className="mr-2 h-4 w-4" />
                  New Cron Job
                </Button>
              </Link>

              <Link to={`/orgs/${currentOrg.slug}/services`}>
                <Button variant="ghost" className="w-full justify-start" size="sm">
                  <Store className="mr-2 h-4 w-4" />
                  Services
                </Button>
              </Link>

              <Separator className="my-2" />

              <Link to={`/orgs/${currentOrg.slug}/settings/members`}>
                <Button variant="ghost" className="w-full justify-start" size="sm">
                  <Users className="mr-2 h-4 w-4" />
                  Members
                </Button>
              </Link>
              <Link to={`/orgs/${currentOrg.slug}/settings`}>
                <Button variant="ghost" className="w-full justify-start" size="sm">
                  <Settings className="mr-2 h-4 w-4" />
                  Org Settings
                </Button>
              </Link>
              <Separator className="my-2" />

              <Link to={`/orgs/${currentOrg.slug}/audit-logs`}>
                <Button variant="ghost" className="w-full justify-start" size="sm">
                  <Shield className="mr-2 h-4 w-4" />
                  Audit Log
                </Button>
              </Link>

              <Link to="/admin/nodes">
                <Button variant="ghost" className="w-full justify-start" size="sm">
                  <Settings className="mr-2 h-4 w-4" />
                  Admin: Nodes
                </Button>
              </Link>
            </>
          )}
        </nav>

        <Separator />

        <div className="p-2">
          <Button
            variant="ghost"
            className="w-full justify-start text-muted-foreground"
            size="sm"
            onClick={handleLogout}
          >
            <LogOut className="mr-2 h-4 w-4" />
            Sign Out
          </Button>
        </div>
      </aside>

      {/* Main content */}
      <div className="flex-1 flex flex-col">
        {/* Top bar */}
        <header className="flex items-center justify-end border-b px-6 py-2 gap-2">
          {currentOrg && <NotificationBell />}
        </header>

        <main className="flex-1 p-8">
        {children || (
          currentOrg ? (
            <Projects />
          ) : (
            <div className="flex flex-col items-center gap-4 py-12 text-center">
              <h1 className="text-3xl font-bold">Welcome to Orbita</h1>
              <p className="text-muted-foreground">
                Create or select an organization to get started.
              </p>
              <Link to="/orgs/new">
                <Button>Create your first organization</Button>
              </Link>
            </div>
          )
        )}
      </main>
      </div>
    </div>
  );
}

export default Dashboard;
