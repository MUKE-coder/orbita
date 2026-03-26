import apiClient from "./client";

export interface Template {
  id: string;
  name: string;
  description: string | null;
  category: string;
  compose_template: string;
  params_schema: Array<{
    name: string;
    type: string;
    default: string;
    required: boolean;
  }>;
  icon_url: string | null;
  is_active: boolean;
}

export interface DeployedService {
  id: string;
  environment_id: string;
  organization_id: string;
  template_id: string | null;
  name: string;
  config: Record<string, unknown>;
  status: string;
  docker_service_ids: string[];
  created_at: string;
  template?: Template;
}

export const servicesApi = {
  listTemplates: () =>
    apiClient.get<{ data: Template[] }>("/templates"),

  getTemplate: (id: string) =>
    apiClient.get<{ data: Template }>(`/templates/${id}`),

  listServices: (orgSlug: string) =>
    apiClient.get<{ data: DeployedService[] }>(`/orgs/${orgSlug}/services`),

  deployService: (
    orgSlug: string,
    data: {
      template_id: string;
      name: string;
      environment_id: string;
      params: Record<string, string>;
    }
  ) =>
    apiClient.post<{ data: DeployedService }>(
      `/orgs/${orgSlug}/services`,
      data
    ),

  getService: (orgSlug: string, serviceId: string) =>
    apiClient.get<{ data: DeployedService }>(
      `/orgs/${orgSlug}/services/${serviceId}`
    ),

  deleteService: (orgSlug: string, serviceId: string) =>
    apiClient.delete(`/orgs/${orgSlug}/services/${serviceId}`),
};
