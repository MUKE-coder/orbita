import { useQuery } from "@tanstack/react-query";
import { useNavigate } from "react-router-dom";
import { ChevronsUpDown, Plus, Building2 } from "lucide-react";

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { orgsApi } from "@/api/orgs";
import { useOrgStore } from "@/stores/org";
import { cn } from "@/lib/utils";

function OrgSwitcher() {
  const navigate = useNavigate();
  const { currentOrg, setCurrentOrg, setOrgs } = useOrgStore();

  const { data } = useQuery({
    queryKey: ["orgs"],
    queryFn: async () => {
      const res = await orgsApi.list();
      const orgs = res.data.data;
      setOrgs(orgs);
      if (!currentOrg && orgs.length > 0) {
        setCurrentOrg(orgs[0]);
      }
      return orgs;
    },
  });

  const orgs = data || [];
  const initials = currentOrg?.name?.slice(0, 2).toUpperCase() || "??";

  return (
    <DropdownMenu>
      <DropdownMenuTrigger className="group flex w-full items-center gap-2.5 rounded-lg border border-sidebar-border bg-sidebar px-2.5 py-2 text-sm text-sidebar-foreground outline-none transition-colors hover:bg-sidebar-accent focus-visible:ring-2 focus-visible:ring-ring/30">
        <div className="flex h-7 w-7 flex-shrink-0 items-center justify-center rounded-md bg-brand/10 text-[11px] font-semibold tracking-tight text-brand">
          {currentOrg ? initials : <Building2 className="h-3.5 w-3.5" />}
        </div>
        <div className="flex min-w-0 flex-1 flex-col items-start">
          <span className="truncate max-w-[140px] text-sm font-medium text-sidebar-foreground">
            {currentOrg?.name || "Select organization"}
          </span>
          {currentOrg && (
            <span className="truncate max-w-[140px] text-[11px] text-muted-foreground">
              {currentOrg.slug}
            </span>
          )}
        </div>
        <ChevronsUpDown className="h-3.5 w-3.5 flex-shrink-0 text-muted-foreground" />
      </DropdownMenuTrigger>
      <DropdownMenuContent className="w-64" align="start">
        <div className="px-2 py-1.5 text-[11px] font-medium uppercase tracking-wider text-muted-foreground">
          Organizations
        </div>
        {orgs.map((org) => (
          <DropdownMenuItem
            key={org.id}
            onClick={() => setCurrentOrg(org)}
            className={cn(
              "flex items-center gap-2.5",
              currentOrg?.id === org.id && "bg-accent"
            )}
          >
            <div className="flex h-6 w-6 flex-shrink-0 items-center justify-center rounded-md bg-brand/10 text-[10px] font-semibold text-brand">
              {org.name.slice(0, 2).toUpperCase()}
            </div>
            <div className="flex min-w-0 flex-col">
              <span className="truncate text-sm font-medium">{org.name}</span>
              <span className="truncate text-[11px] text-muted-foreground">
                {org.slug}
              </span>
            </div>
            {currentOrg?.id === org.id && (
              <span className="ml-auto text-xs text-brand">✓</span>
            )}
          </DropdownMenuItem>
        ))}
        <DropdownMenuSeparator />
        <DropdownMenuItem onClick={() => navigate("/orgs/new")}>
          <Plus className="mr-2 h-4 w-4" />
          Create organization
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

export default OrgSwitcher;
