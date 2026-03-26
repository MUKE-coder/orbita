import apiClient from "./client";

export interface Domain {
  id: string;
  resource_id: string;
  resource_type: string;
  organization_id: string;
  domain: string;
  ssl_enabled: boolean;
  status: string;
  verified: boolean;
  created_at: string;
}

export const domainsApi = {
  listByOrg: (orgSlug: string) =>
    apiClient.get<{ data: Domain[] }>(`/orgs/${orgSlug}/domains`),

  listByApp: (orgSlug: string, appId: string) =>
    apiClient.get<{ data: Domain[] }>(
      `/orgs/${orgSlug}/apps/${appId}/domains`
    ),

  addToApp: (
    orgSlug: string,
    appId: string,
    data: { domain: string; ssl_enabled?: boolean; port?: number }
  ) =>
    apiClient.post<{ data: Domain }>(
      `/orgs/${orgSlug}/apps/${appId}/domains`,
      data
    ),

  remove: (orgSlug: string, domainId: string) =>
    apiClient.delete(`/orgs/${orgSlug}/domains/${domainId}`),

  verify: (orgSlug: string, domain: string) =>
    apiClient.get<{ data: { domain: string; verified: boolean } }>(
      `/orgs/${orgSlug}/domains/verify?domain=${domain}`
    ),
};
