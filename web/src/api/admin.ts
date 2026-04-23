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
};
