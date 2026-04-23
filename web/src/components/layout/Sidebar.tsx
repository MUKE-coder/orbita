import { Link, useLocation, useNavigate } from "react-router-dom";
import { useQuery } from "@tanstack/react-query";
import {
  LayoutDashboard,
  FolderKanban,
  Store,
  Users,
  Settings,
  Server,
  Building2,
  ScrollText,
  Rocket,
  Database,
  Clock,
  GitBranch,
  PanelLeftClose,
  PanelLeftOpen,
  LogOut,
  User,
  ChevronsUpDown,
} from "lucide-react";
import type { LucideIcon } from "lucide-react";

import OrgSwitcher from "@/components/layout/OrgSwitcher";
import Logo from "@/components/layout/Logo";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { useOrgStore } from "@/stores/org";
import { useAuthStore } from "@/stores/auth";
import { useUIStore } from "@/stores/ui";
import { projectsApi } from "@/api/projects";
import { authApi } from "@/api/auth";
import { cn } from "@/lib/utils";

interface NavItemProps {
  to: string;
  icon: LucideIcon;
  label: string;
  exact?: boolean;
  badge?: string | number;
  collapsed?: boolean;
}

function NavItem({ to, icon: Icon, label, exact, badge, collapsed }: NavItemProps) {
  const { pathname } = useLocation();
  const active = exact ? pathname === to : pathname.startsWith(to);

  return (
    <Link
      to={to}
      title={collapsed ? label : undefined}
      className={cn(
        "group flex items-center gap-2.5 rounded-md px-2.5 py-1.5 text-[13px] font-medium transition-colors",
        collapsed && "justify-center px-2",
        active
          ? "bg-sidebar-accent text-sidebar-accent-foreground"
          : "text-sidebar-foreground/80 hover:bg-sidebar-accent/60 hover:text-sidebar-foreground"
      )}
    >
      <Icon
        className={cn(
          "h-4 w-4 shrink-0 transition-colors",
          active ? "text-brand" : "text-muted-foreground group-hover:text-foreground"
        )}
      />
      {!collapsed && (
        <>
          <span className="truncate">{label}</span>
          {badge !== undefined && (
            <span className="ml-auto rounded-md bg-muted px-1.5 py-0.5 text-[10px] font-medium text-muted-foreground">
              {badge}
            </span>
          )}
        </>
      )}
    </Link>
  );
}

function SectionLabel({
  children,
  collapsed,
}: {
  children: string;
  collapsed: boolean;
}) {
  if (collapsed) {
    return <div className="my-2 mx-auto h-px w-5 bg-sidebar-border" />;
  }
  return (
    <div className="px-2.5 pb-1.5 pt-4 text-[11px] font-medium uppercase tracking-wider text-muted-foreground">
      {children}
    </div>
  );
}

export function Sidebar() {
  const currentOrg = useOrgStore((s) => s.currentOrg);
  const collapsed = useUIStore((s) => s.sidebarCollapsed);
  const toggle = useUIStore((s) => s.toggleSidebar);

  const { data: projectsData } = useQuery({
    queryKey: ["projects", currentOrg?.slug],
    queryFn: () => projectsApi.list(currentOrg!.slug),
    enabled: !!currentOrg,
  });

  const projects = projectsData?.data?.data || [];

  return (
    <aside
      className={cn(
        "flex flex-col border-r border-sidebar-border bg-sidebar transition-[width] duration-200",
        collapsed ? "w-16" : "w-64"
      )}
    >
      {/* Brand header + collapse toggle */}
      <div
        className={cn(
          "flex h-14 items-center border-b border-sidebar-border",
          collapsed ? "justify-center px-2" : "px-3"
        )}
      >
        {collapsed ? (
          <button
            onClick={toggle}
            className="inline-flex h-9 w-9 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-sidebar-accent hover:text-foreground"
            aria-label="Expand sidebar"
            title="Expand sidebar"
          >
            <PanelLeftOpen className="h-4 w-4" />
          </button>
        ) : (
          <>
            <Logo size="md" />
            <div className="ml-auto flex items-center gap-1">
              <span className="rounded-md bg-brand/10 px-1.5 py-0.5 text-[10px] font-medium text-brand">
                v0.1.0
              </span>
              <button
                onClick={toggle}
                className="inline-flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-sidebar-accent hover:text-foreground"
                aria-label="Collapse sidebar"
                title="Collapse sidebar"
              >
                <PanelLeftClose className="h-4 w-4" />
              </button>
            </div>
          </>
        )}
      </div>

      {/* Org switcher (hidden when collapsed) */}
      {!collapsed && (
        <div className="p-3">
          <OrgSwitcher />
        </div>
      )}

      {/* Primary nav */}
      <nav
        className={cn(
          "flex-1 overflow-y-auto pb-4",
          collapsed ? "px-2 pt-3" : "px-3"
        )}
      >
        {currentOrg && (
          <>
            <div className="space-y-0.5">
              <NavItem
                to={`/dashboard`}
                icon={LayoutDashboard}
                label="Overview"
                exact
                collapsed={collapsed}
              />
              <NavItem
                to={`/orgs/${currentOrg.slug}/projects`}
                icon={FolderKanban}
                label="Projects"
                badge={collapsed ? undefined : projects.length || undefined}
                collapsed={collapsed}
              />
              <NavItem
                to={`/orgs/${currentOrg.slug}/services`}
                icon={Store}
                label="Marketplace"
                collapsed={collapsed}
              />
            </div>

            <SectionLabel collapsed={collapsed}>Create</SectionLabel>
            <div className="space-y-0.5">
              <NavItem
                to={`/orgs/${currentOrg.slug}/apps/new`}
                icon={Rocket}
                label="New App"
                collapsed={collapsed}
              />
              <NavItem
                to={`/orgs/${currentOrg.slug}/databases/new`}
                icon={Database}
                label="New Database"
                collapsed={collapsed}
              />
              <NavItem
                to={`/orgs/${currentOrg.slug}/cron-jobs/new`}
                icon={Clock}
                label="New Cron Job"
                collapsed={collapsed}
              />
            </div>

            {!collapsed && projects.length > 0 && (
              <>
                <SectionLabel collapsed={collapsed}>Projects</SectionLabel>
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

            <SectionLabel collapsed={collapsed}>Organization</SectionLabel>
            <div className="space-y-0.5">
              <NavItem
                to={`/orgs/${currentOrg.slug}/git`}
                icon={GitBranch}
                label="Git connections"
                collapsed={collapsed}
              />
              <NavItem
                to={`/orgs/${currentOrg.slug}/settings/members`}
                icon={Users}
                label="Members"
                collapsed={collapsed}
              />
              <NavItem
                to={`/orgs/${currentOrg.slug}/settings`}
                icon={Settings}
                label="Settings"
                collapsed={collapsed}
              />
              <NavItem
                to={`/orgs/${currentOrg.slug}/audit-logs`}
                icon={ScrollText}
                label="Audit log"
                collapsed={collapsed}
              />
            </div>

            <SectionLabel collapsed={collapsed}>Admin</SectionLabel>
            <div className="space-y-0.5">
              <NavItem
                to="/admin/orgs"
                icon={Building2}
                label="Organizations"
                collapsed={collapsed}
              />
              <NavItem
                to="/admin/nodes"
                icon={Server}
                label="Nodes"
                collapsed={collapsed}
              />
            </div>
          </>
        )}
      </nav>

      {/* Account menu at bottom */}
      <div className="border-t border-sidebar-border p-3">
        <AccountMenu collapsed={collapsed} />
      </div>
    </aside>
  );
}

function AccountMenu({ collapsed }: { collapsed: boolean }) {
  const navigate = useNavigate();
  const logout = useAuthStore((s) => s.logout);

  // Fetch the authenticated user's profile for display.
  const { data } = useQuery({
    queryKey: ["me"],
    queryFn: () => authApi.getProfile(),
    staleTime: 60_000,
  });
  const user = data?.data?.data;

  const handleLogout = () => {
    logout();
    navigate("/login");
  };

  const displayName = user?.name || "Account";
  const displayEmail = user?.email || "";
  const initials = user?.name
    ? user.name
        .split(/\s+/)
        .slice(0, 2)
        .map((p) => p[0]?.toUpperCase() || "")
        .join("") || user.name.slice(0, 2).toUpperCase()
    : "";

  return (
    <DropdownMenu>
      <DropdownMenuTrigger
        className={cn(
          "group flex w-full items-center gap-2.5 rounded-lg border border-transparent px-2 py-2 text-left text-sm text-sidebar-foreground outline-none transition-colors hover:bg-sidebar-accent focus-visible:ring-2 focus-visible:ring-ring/30",
          collapsed && "justify-center px-0"
        )}
      >
        <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-linear-to-br from-brand/70 to-brand text-[11px] font-semibold text-brand-foreground">
          {initials || <User className="h-3.5 w-3.5" />}
        </div>
        {!collapsed && (
          <>
            <div className="flex min-w-0 flex-1 flex-col">
              <span className="truncate text-[13px] font-medium">
                {displayName}
              </span>
              {displayEmail && (
                <span className="truncate text-[11px] text-muted-foreground">
                  {displayEmail}
                </span>
              )}
            </div>
            <ChevronsUpDown className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
          </>
        )}
      </DropdownMenuTrigger>
      <DropdownMenuContent
        align={collapsed ? "end" : "start"}
        side={collapsed ? "right" : "top"}
        className="w-64"
      >
        {user && (
          <>
            <div className="px-2 py-2">
              <div className="truncate text-sm font-medium text-foreground">
                {user.name}
              </div>
              <div className="truncate text-xs text-muted-foreground">
                {user.email}
              </div>
              <div className="mt-1.5 flex flex-wrap gap-1">
                {user.is_super_admin && (
                  <span className="rounded-sm bg-brand/10 px-1.5 py-0.5 text-[10px] font-medium text-brand">
                    Super admin
                  </span>
                )}
                {user.is_email_verified ? (
                  <span className="rounded-sm bg-success/10 px-1.5 py-0.5 text-[10px] font-medium text-success">
                    Verified
                  </span>
                ) : (
                  <span className="rounded-sm bg-warning/10 px-1.5 py-0.5 text-[10px] font-medium text-warning-foreground">
                    Email unverified
                  </span>
                )}
              </div>
            </div>
            <DropdownMenuSeparator />
          </>
        )}
        <DropdownMenuLabel className="text-[11px] font-medium uppercase tracking-wider text-muted-foreground">
          Account
        </DropdownMenuLabel>
        <DropdownMenuItem onClick={handleLogout} className="text-destructive">
          <LogOut className="mr-2 h-4 w-4" />
          Sign out
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

export default Sidebar;
