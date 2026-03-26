import apiClient from "./client";

export interface GitConnection {
  id: string;
  organization_id: string;
  provider: string;
  metadata: { base_url?: string };
  created_at: string;
}

export const gitApi = {
  listConnections: (orgSlug: string) =>
    apiClient.get<{ data: GitConnection[] }>(
      `/orgs/${orgSlug}/git-connections`
    ),

  createConnection: (
    orgSlug: string,
    data: {
      provider: string;
      access_token: string;
      refresh_token?: string;
      base_url?: string;
    }
  ) =>
    apiClient.post<{ data: GitConnection }>(
      `/orgs/${orgSlug}/git-connections`,
      data
    ),

  deleteConnection: (orgSlug: string, id: string) =>
    apiClient.delete(`/orgs/${orgSlug}/git-connections/${id}`),

  listRepos: (orgSlug: string, connId: string) =>
    apiClient.get<{
      data: Array<{
        full_name: string;
        clone_url: string;
        default_branch: string;
      }>;
    }>(`/orgs/${orgSlug}/git-connections/${connId}/repos`),

  listBranches: (
    orgSlug: string,
    connId: string,
    owner: string,
    repo: string
  ) =>
    apiClient.get<{ data: string[] }>(
      `/orgs/${orgSlug}/git-connections/${connId}/repos/${owner}/${repo}/branches`
    ),
};
