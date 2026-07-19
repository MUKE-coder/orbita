import apiClient from "./client";
import type { Organization } from "./orgs";

export interface Node {
  id: string;
  name: string;
  ip: string;
  ssh_port: number;
  role: string;
  status: string;
  labels: Record<string, string>;
  cpu_cores: number;
  ram_mb: number;
  disk_gb: number;
  docker_version: string | null;
  created_at: string;
}

export interface NodeMetrics {
  cpu_percent: number;
  memory_used: number;
  memory_total: number;
  disk_used: number;
  disk_total: number;
  container_count: number;
  uptime_seconds: number;
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

export const adminApi = {
  // Nodes
  listNodes: () => apiClient.get<{ data: Node[] }>("/admin/nodes"),

  addNode: (data: { name: string; ip: string; ssh_port?: number; ssh_private_key: string }) =>
    apiClient.post<{ data: Node }>("/admin/nodes", data),

  getNode: (id: string) => apiClient.get<{ data: Node }>(`/admin/nodes/${id}`),

  getNodeMetrics: (id: string) =>
    apiClient.get<{ data: NodeMetrics }>(`/admin/nodes/${id}/metrics`),

  drainNode: (id: string) => apiClient.post(`/admin/nodes/${id}/drain`),

  removeNode: (id: string) => apiClient.delete(`/admin/nodes/${id}`),

  getPlatformMetrics: () =>
    apiClient.get<{ data: { total_nodes: number; online_nodes: number } }>("/admin/platform/metrics"),

  getPlatformCapacity: () =>
    apiClient.get<{
      data: {
        host: { cpu_cores: number; ram_mb: number; disk_gb: number };
        allocated: { cpu_cores: number; ram_mb: number; disk_gb: number };
        available: { cpu_cores: number; ram_mb: number; disk_gb: number };
        orgs: Array<{
          slug: string;
          name: string;
          cpu_cores: number;
          ram_mb: number;
          disk_gb: number;
        }>;
      };
    }>("/admin/platform/capacity"),

  // Plans
  listPlans: () => apiClient.get<{ data: ResourcePlan[] }>("/admin/plans"),

  createPlan: (data: Omit<ResourcePlan, "id">) =>
    apiClient.post<{ data: ResourcePlan }>("/admin/plans", data),

  updatePlan: (id: string, data: Partial<ResourcePlan>) =>
    apiClient.put<{ data: ResourcePlan }>(`/admin/plans/${id}`, data),

  deletePlan: (id: string) => apiClient.delete(`/admin/plans/${id}`),

  // Orgs
  listAllOrgs: () => apiClient.get<{ data: Organization[] }>("/admin/orgs"),

  assignPlan: (orgSlug: string, planId: string) =>
    apiClient.put(`/admin/orgs/${orgSlug}/plan`, { plan_id: planId }),

  // Onboard a tenant: the org plus the account its admin signs in with.
  provisionOrg: (data: {
    name: string;
    slug?: string;
    plan_id?: string;
    custom_cpu_cores?: number;
    custom_ram_mb?: number;
    custom_disk_gb?: number;
    custom_max_apps?: number;
    custom_max_databases?: number;
    billing_type?: string;
    price_monthly_cents?: number;
    currency?: string;
    billing_cycle?: string;
    admin_email: string;
    admin_name?: string;
    admin_password?: string;
    admin_role?: string;
  }) => apiClient.post<{ data: ProvisionResult }>("/admin/orgs", data),

  // Email settings
  getEmailSettings: () =>
    apiClient.get<{ data: EmailSettings }>("/admin/settings/email"),

  updateEmailSettings: (data: {
    api_key?: string | null;
    email_from: string;
    email_from_name: string;
  }) => apiClient.put<{ data: EmailSettings }>("/admin/settings/email", data),

  sendTestEmail: (to: string) =>
    apiClient.post<{ data: { message: string } }>("/admin/settings/email/test", { to }),
};

export interface EmailSettings {
  configured: boolean;
  has_api_key: boolean;
  email_from: string;
  email_from_name: string;
  source: "dashboard" | "environment" | "unset";
}

export interface ProvisionResult {
  organization: { id: string; name: string; slug: string };
  user: { id: string; email: string; name: string };
  password: string;
  generated: boolean;
}
