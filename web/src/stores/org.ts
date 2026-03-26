import { create } from "zustand";
import type { Organization } from "@/api/orgs";

interface OrgState {
  currentOrg: Organization | null;
  orgs: Organization[];
  setCurrentOrg: (org: Organization | null) => void;
  setOrgs: (orgs: Organization[]) => void;
}

export const useOrgStore = create<OrgState>((set) => ({
  currentOrg: null,
  orgs: [],
  setCurrentOrg: (org) => set({ currentOrg: org }),
  setOrgs: (orgs) => set({ orgs }),
}));
