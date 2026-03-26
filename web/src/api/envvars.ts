import apiClient from "./client";

export interface EnvVar {
  id: string;
  key: string;
  value: string;
  is_secret: boolean;
  resource_id: string;
  resource_type: string;
  created_at: string;
  updated_at: string;
}

export const envVarsApi = {
  list: (orgSlug: string, appId: string) =>
    apiClient.get<{ data: EnvVar[] }>(`/orgs/${orgSlug}/apps/${appId}/env`),

  set: (orgSlug: string, appId: string, data: { key: string; value: string; is_secret: boolean }) =>
    apiClient.post(`/orgs/${orgSlug}/apps/${appId}/env`, data),

  bulkSet: (orgSlug: string, appId: string, variables: Array<{ key: string; value: string; is_secret: boolean }>) =>
    apiClient.put(`/orgs/${orgSlug}/apps/${appId}/env/bulk`, { variables }),

  importDotenv: (orgSlug: string, appId: string, content: string) =>
    apiClient.post<{ data: { message: string; imported: number } }>(
      `/orgs/${orgSlug}/apps/${appId}/env/import`,
      { content }
    ),

  delete: (orgSlug: string, appId: string, envId: string) =>
    apiClient.delete(`/orgs/${orgSlug}/apps/${appId}/env/${envId}`),
};
