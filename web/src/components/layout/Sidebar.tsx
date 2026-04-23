import { Link, useLocation } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import {
  LayoutDashboard,
  FolderKanban,
  Store,
  Users,
  Settings,
  Server,
  ScrollText,
  Rocket,
  Database,
  Clock,
} from "lucide-react";
import type { LucideIcon } from "lucide-react";

import OrgSwitcher from "@/components/layout/OrgSwitcher";
import Logo from "@/components/layout/Logo";
import { useOrgStore } from "@/stores/org";
import { projectsApi } from "@/api/projects";
import { cn } from "@/lib/utils";

interface NavItemProps {
  to: string;
  icon: LucideIcon;
  label: string;
  exact?: boolean;
  badge?: string | number;
}

function NavItem({ to, icon: Icon, label, exact, badge }: NavItemProps) {
  const { pathname } = useLocation();
  const active = exact ? pathname === to : pathname.startsWith(to);

  return (
    <Link
      to={to}
      className={cn(
        "group flex items-center gap-2.5 rounded-md px-2.5 py-1.5 text-[13px] font-medium transition-colors",
        active
          ? "bg-sidebar-accent text-sidebar-accent-foreground"
          : "text-sidebar-foreground/80 hover:bg-sidebar-accent/60 hover:text-sidebar-foreground"
      )}
    >
      <Icon
        className={cn(
          "h-4 w-4 flex-shrink-0 transition-colors",
          active ? "text-brand" : "text-muted-foreground group-hover:text-foreground"
        )}
      />
      <span className="truncate">{label}</span>
      {badge !== undefined && (
        <span className="ml-auto rounded-md bg-muted px-1.5 py-0.5 text-[10px] font-medium text-muted-foreground">
          {badge}
        </span>
      )}
    </Link>
  );
}

function SectionLabel({ children }: { children: string }) {
  return (
    <div className="px-2.5 pb-1.5 pt-4 text-[11px] font-medium uppercase tracking-wider text-muted-foreground">
      {children}
    </div>
  );
}

export function Sidebar() {
  const currentOrg = useOrgStore((s) => s.currentOrg);

  const { data: projectsData } = useQuery({
    queryKey: ["projects", currentOrg?.slug],
    queryFn: () => projectsApi.list(currentOrg!.slug),
    enabled: !!currentOrg,
  });

  const projects = projectsData?.data?.data || [];

  return (
    <aside className="flex w-64 flex-col border-r border-sidebar-border bg-sidebar">
      {/* Brand header */}
      <div className="flex h-14 items-center border-b border-sidebar-border px-4">
        <Logo size="md" />
        <span className="ml-auto rounded-md bg-brand/10 px-1.5 py-0.5 text-[10px] font-medium text-brand">
          v0.1.0
        </span>
      </div>

      {/* Org switcher */}
      <div className="p-3">
        <OrgSwitcher />
      </div>

      {/* Primary nav */}
      <nav className="flex-1 overflow-y-auto px-3 pb-4">
        {currentOrg && (
          <>
            <div className="space-y-0.5">
              <NavItem
                to={`/orgs/${currentOrg.slug}/projects`}
                icon={LayoutDashboard}
                label="Overview"
                exact
              />
              <NavItem
                to={`/orgs/${currentOrg.slug}/projects`}
                icon={FolderKanban}
                label="Projects"
                badge={projects.length || undefined}
              />
              <NavItem
                to={`/orgs/${currentOrg.slug}/services`}
                icon={Store}
                label="Marketplace"
              />
            </div>

            <SectionLabel>Create</SectionLabel>
            <div className="space-y-0.5">
              <NavItem
                to={`/orgs/${currentOrg.slug}/apps/new`}
                icon={Rocket}
                label="New App"
              />
              <NavItem
                to={`/orgs/${currentOrg.slug}/databases/new`}
                icon={Database}
                label="New Database"
              />
              <NavItem
                to={`/orgs/${currentOrg.slug}/cron-jobs/new`}
                icon={Clock}
                label="New Cron Job"
              />
            </div>

            {projects.length > 0 && (
              <>
                <SectionLabel>Projects</SectionLabel>
                <div className="space-y-0.5">
                  {projects.slice(0, 8).map((p) => (
                    <Link
                      key={p.id}
                      to={`/orgs/${currentOrg.slug}/projects/${p.id}`}
                      className="group flex items-center gap-2.5 rounded-md px-2.5 py-1.5 text-[13px] text-sidebar-foreground/80 transition-colors hover:bg-sidebar-accent/60 hover:text-sidebar-foreground"
                    >
                      <span className="text-sm leading-none">{p.emoji || "📦"}</span>
                      <span className="truncate">{p.name}</span>
                    </Link>
                  ))}
                </div>
              </>
            )}

            <SectionLabel>Organization</SectionLabel>
            <div className="space-y-0.5">
              <NavItem
                to={`/orgs/${currentOrg.slug}/settings/members`}
                icon={Users}
                label="Members"
              />
              <NavItem
                to={`/orgs/${currentOrg.slug}/settings`}
                icon={Settings}
                label="Settings"
              />
              <NavItem
                to={`/orgs/${currentOrg.slug}/audit-logs`}
                icon={ScrollText}
                label="Audit log"
              />
            </div>

            <SectionLabel>Admin</SectionLabel>
            <div className="space-y-0.5">
              <NavItem to="/admin/nodes" icon={Server} label="Nodes" />
            </div>
          </>
        )}
      </nav>
    </aside>
  );
}

export default Sidebar;
