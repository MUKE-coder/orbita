import apiClient from "./client";

export interface Organization {
  id: string;
  name: string;
  slug: string;
  description: string | null;
  plan_id: string | null;
  plan: ResourcePlan | null;
  created_by: string;
  created_at: string;
}

export interface OrgMember {
  org_id: string;
  user_id: string;
  role: string;
  joined_at: string;
  user?: {
    id: string;
    name: string;
    email: string;
    avatar_url: string | null;
  };
}

export interface OrgInvite {
  id: string;
  org_id: string;
  email: string;
  role: string;
  expires_at: string;
  created_at: string;
  inviter?: { id: string; name: string; email: string };
}

export interface ResourcePlan {
  id: string;
  name: string;
  max_cpu_cores: number;
  max_ram_mb: number;
  max_disk_gb: number;
  max_apps: number;
  max_databases: number;
}

export const orgsApi = {
  list: () => apiClient.get<{ data: Organization[] }>("/orgs"),

  create: (data: { name: string; slug: string }) =>
    apiClient.post<{ data: Organization }>("/orgs", data),

  get: (slug: string) =>
    apiClient.get<{
      data: { organization: Organization; member_count: number };
    }>(`/orgs/${slug}`),

  update: (slug: string, data: { name?: string; description?: string }) =>
    apiClient.put<{ data: Organization }>(`/orgs/${slug}`, data),

  delete: (slug: string) => apiClient.delete(`/orgs/${slug}`),

  listMembers: (slug: string) =>
    apiClient.get<{ data: OrgMember[] }>(`/orgs/${slug}/members`),

  inviteMember: (slug: string, data: { email: string; role: string }) =>
    apiClient.post(`/orgs/${slug}/invites`, data),

  listInvites: (slug: string) =>
    apiClient.get<{ data: OrgInvite[] }>(`/orgs/${slug}/invites`),

  revokeInvite: (slug: string, id: string) =>
    apiClient.delete(`/orgs/${slug}/invites/${id}`),

  updateMemberRole: (slug: string, userId: string, role: string) =>
    apiClient.put(`/orgs/${slug}/members/${userId}/role`, { role }),

  removeMember: (slug: string, userId: string) =>
    apiClient.delete(`/orgs/${slug}/members/${userId}`),

  leave: (slug: string) => apiClient.post(`/orgs/${slug}/leave`),

  getInviteInfo: (token: string) =>
    apiClient.get<{
      data: {
        organization: string;
        org_slug: string;
        role: string;
        email: string;
      };
    }>(`/join?token=${token}`),

  acceptInvite: (token: string) =>
    apiClient.post(`/join?token=${token}`),
};
