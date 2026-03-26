import apiClient from "./client";

export interface Application {
  id: string;
  environment_id: string;
  organization_id: string;
  name: string;
  source_type: string;
  source_config: { image?: string };
  status: string;
  docker_service_id: string | null;
  replicas: number;
  port: number | null;
  created_at: string;
}

export interface Deployment {
  id: string;
  app_id: string;
  version: number;
  image_ref: string;
  status: string;
  started_at: string | null;
  finished_at: string | null;
  triggered_by: string | null;
  trigger_type: string;
  error_message: string | null;
  created_at: string;
}

export const appsApi = {
  list: (orgSlug: string) =>
    apiClient.get<{ data: Application[] }>(`/orgs/${orgSlug}/apps`),

  create: (
    orgSlug: string,
    data: {
      name: string;
      environment_id: string;
      source_type: string;
      image: string;
      port?: number;
      replicas?: number;
    }
  ) => apiClient.post<{ data: Application }>(`/orgs/${orgSlug}/apps`, data),

  get: (orgSlug: string, appId: string) =>
    apiClient.get<{ data: Application }>(`/orgs/${orgSlug}/apps/${appId}`),

  update: (orgSlug: string, appId: string, data: Record<string, unknown>) =>
    apiClient.put<{ data: Application }>(
      `/orgs/${orgSlug}/apps/${appId}`,
      data
    ),

  delete: (orgSlug: string, appId: string) =>
    apiClient.delete(`/orgs/${orgSlug}/apps/${appId}`),

  deploy: (orgSlug: string, appId: string) =>
    apiClient.post<{ data: Deployment }>(
      `/orgs/${orgSlug}/apps/${appId}/deploy`
    ),

  rollback: (orgSlug: string, appId: string, deploymentId: string) =>
    apiClient.post(
      `/orgs/${orgSlug}/apps/${appId}/rollback/${deploymentId}`
    ),

  stop: (orgSlug: string, appId: string) =>
    apiClient.post(`/orgs/${orgSlug}/apps/${appId}/stop`),

  start: (orgSlug: string, appId: string) =>
    apiClient.post(`/orgs/${orgSlug}/apps/${appId}/start`),

  restart: (orgSlug: string, appId: string) =>
    apiClient.post(`/orgs/${orgSlug}/apps/${appId}/restart`),

  listDeployments: (orgSlug: string, appId: string) =>
    apiClient.get<{ data: Deployment[] }>(
      `/orgs/${orgSlug}/apps/${appId}/deployments`
    ),

  getStatus: (orgSlug: string, appId: string) =>
    apiClient.get<{ data: { status: string } }>(
      `/orgs/${orgSlug}/apps/${appId}/status`
    ),

  getLogs: (orgSlug: string, appId: string) =>
    apiClient.get<{ data: { logs: string } }>(
      `/orgs/${orgSlug}/apps/${appId}/logs`
    ),

  getMetrics: (orgSlug: string, appId: string) =>
    apiClient.get<{
      data: {
        cpu_percent: number;
        memory_usage: number;
        memory_limit: number;
        network_rx: number;
        network_tx: number;
        status: string;
      };
    }>(`/orgs/${orgSlug}/apps/${appId}/metrics`),
};
