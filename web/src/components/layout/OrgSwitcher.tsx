import { useQuery } from "@tanstack/react-query";
import { useNavigate } from "react-router-dom";
import { ChevronsUpDown, Plus } from "lucide-react";

import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { orgsApi } from "@/api/orgs";
import { useOrgStore } from "@/stores/org";

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

  return (
    <DropdownMenu>
      <DropdownMenuTrigger className="flex w-full items-center justify-between rounded-md border px-3 py-2 text-sm hover:bg-accent">
        <span className="truncate font-medium">
          {currentOrg?.name || "Select Organization"}
        </span>
        <ChevronsUpDown className="ml-2 h-4 w-4 shrink-0 opacity-50" />
      </DropdownMenuTrigger>
      <DropdownMenuContent className="w-56" align="start">
        {orgs.map((org) => (
          <DropdownMenuItem
            key={org.id}
            onClick={() => setCurrentOrg(org)}
            className={currentOrg?.id === org.id ? "bg-accent" : ""}
          >
            <span className="truncate">{org.name}</span>
          </DropdownMenuItem>
        ))}
        <DropdownMenuSeparator />
        <DropdownMenuItem onClick={() => navigate("/orgs/new")}>
          <Plus className="mr-2 h-4 w-4" />
          Create Organization
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

export default OrgSwitcher;
