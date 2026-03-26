import { Link, useNavigate } from "react-router-dom";
import { LogOut, Settings, Rocket } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";
import OrgSwitcher from "@/components/layout/OrgSwitcher";
import { useAuthStore } from "@/stores/auth";
import { useOrgStore } from "@/stores/org";

function Dashboard() {
  const navigate = useNavigate();
  const logout = useAuthStore((s) => s.logout);
  const currentOrg = useOrgStore((s) => s.currentOrg);

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

        <nav className="flex-1 p-2 space-y-1">
          {currentOrg && (
            <>
              <Link to={`/orgs/${currentOrg.slug}/settings`}>
                <Button variant="ghost" className="w-full justify-start" size="sm">
                  <Settings className="mr-2 h-4 w-4" />
                  Org Settings
                </Button>
              </Link>
              <Link to={`/orgs/${currentOrg.slug}/settings/members`}>
                <Button variant="ghost" className="w-full justify-start" size="sm">
                  Members
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
      <main className="flex-1 p-8">
        <h1 className="text-3xl font-bold">Dashboard</h1>
        {currentOrg ? (
          <p className="mt-4 text-muted-foreground">
            Organization: <strong>{currentOrg.name}</strong> — Dashboard content
            will be built in Phase 14.
          </p>
        ) : (
          <div className="mt-8 flex flex-col items-center gap-4 text-center">
            <p className="text-muted-foreground">
              You don't belong to any organization yet.
            </p>
            <Link to="/orgs/new">
              <Button>Create your first organization</Button>
            </Link>
          </div>
        )}
      </main>
    </div>
  );
}

export default Dashboard;
