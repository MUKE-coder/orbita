import apiClient from "./client";

export interface Project {
  id: string;
  organization_id: string;
  name: string;
  description: string | null;
  emoji: string;
  created_at: string;
  environments?: Environment[];
}

export interface Environment {
  id: string;
  project_id: string;
  name: string;
  type: string;
  created_at: string;
}

export const projectsApi = {
  list: (orgSlug: string) =>
    apiClient.get<{ data: Project[] }>(`/orgs/${orgSlug}/projects`),

  create: (
    orgSlug: string,
    data: { name: string; description?: string; emoji?: string }
  ) => apiClient.post<{ data: Project }>(`/orgs/${orgSlug}/projects`, data),

  get: (orgSlug: string, projectId: string) =>
    apiClient.get<{ data: Project }>(`/orgs/${orgSlug}/projects/${projectId}`),

  update: (
    orgSlug: string,
    projectId: string,
    data: { name?: string; description?: string; emoji?: string }
  ) =>
    apiClient.put<{ data: Project }>(
      `/orgs/${orgSlug}/projects/${projectId}`,
      data
    ),

  delete: (orgSlug: string, projectId: string) =>
    apiClient.delete(`/orgs/${orgSlug}/projects/${projectId}`),

  listEnvironments: (orgSlug: string, projectId: string) =>
    apiClient.get<{ data: Environment[] }>(
      `/orgs/${orgSlug}/projects/${projectId}/environments`
    ),

  createEnvironment: (
    orgSlug: string,
    projectId: string,
    data: { name: string; type?: string }
  ) =>
    apiClient.post<{ data: Environment }>(
      `/orgs/${orgSlug}/projects/${projectId}/environments`,
      data
    ),

  updateEnvironment: (
    orgSlug: string,
    projectId: string,
    envId: string,
    data: { name: string }
  ) =>
    apiClient.put<{ data: Environment }>(
      `/orgs/${orgSlug}/projects/${projectId}/environments/${envId}`,
      data
    ),

  deleteEnvironment: (
    orgSlug: string,
    projectId: string,
    envId: string
  ) =>
    apiClient.delete(
      `/orgs/${orgSlug}/projects/${projectId}/environments/${envId}`
    ),
};
