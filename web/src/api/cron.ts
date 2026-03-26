import apiClient from "./client";

export interface CronJob {
  id: string;
  environment_id: string;
  organization_id: string;
  name: string;
  schedule: string;
  image: string;
  command: string | null;
  timeout: number;
  concurrency_policy: string;
  max_retries: number;
  cpu_limit: number;
  memory_limit: number;
  enabled: boolean;
  last_run_at: string | null;
  next_run_at: string | null;
  created_at: string;
}

export interface CronRun {
  id: string;
  cron_job_id: string;
  started_at: string;
  finished_at: string | null;
  status: string;
  exit_code: number | null;
  log_snippet: string | null;
  duration_ms: number | null;
  created_at: string;
}

export const cronApi = {
  list: (orgSlug: string) =>
    apiClient.get<{ data: CronJob[] }>(`/orgs/${orgSlug}/cron-jobs`),

  create: (
    orgSlug: string,
    data: {
      name: string;
      schedule: string;
      image: string;
      command?: string;
      environment_id: string;
      timeout?: number;
      concurrency_policy?: string;
      max_retries?: number;
    }
  ) => apiClient.post<{ data: CronJob }>(`/orgs/${orgSlug}/cron-jobs`, data),

  get: (orgSlug: string, cronId: string) =>
    apiClient.get<{ data: CronJob }>(`/orgs/${orgSlug}/cron-jobs/${cronId}`),

  update: (orgSlug: string, cronId: string, data: Record<string, unknown>) =>
    apiClient.put<{ data: CronJob }>(
      `/orgs/${orgSlug}/cron-jobs/${cronId}`,
      data
    ),

  delete: (orgSlug: string, cronId: string) =>
    apiClient.delete(`/orgs/${orgSlug}/cron-jobs/${cronId}`),

  toggle: (orgSlug: string, cronId: string) =>
    apiClient.post<{ data: CronJob }>(
      `/orgs/${orgSlug}/cron-jobs/${cronId}/toggle`
    ),

  trigger: (orgSlug: string, cronId: string) =>
    apiClient.post(`/orgs/${orgSlug}/cron-jobs/${cronId}/run`),

  listRuns: (orgSlug: string, cronId: string) =>
    apiClient.get<{ data: CronRun[] }>(
      `/orgs/${orgSlug}/cron-jobs/${cronId}/runs`
    ),

  getRunLogs: (orgSlug: string, cronId: string, runId: string) =>
    apiClient.get<{ data: { logs: string } }>(
      `/orgs/${orgSlug}/cron-jobs/${cronId}/runs/${runId}/logs`
    ),
};
