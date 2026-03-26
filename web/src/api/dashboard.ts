import apiClient from "./client";

export interface DashboardData {
  total_apps: number;
  running_apps: number;
  stopped_apps: number;
  total_databases: number;
  running_databases: number;
  total_cron_jobs: number;
  active_cron_jobs: number;
  recent_deploys: Array<{
    id: string;
    app_name: string;
    version: number;
    status: string;
    started_at: string | null;
    image_ref: string;
  }>;
}

export interface MetricsOverview {
  cpu_percent: number;
  memory_used: number;
  memory_total: number;
  disk_used: number;
  disk_total: number;
  network_rx: number;
  network_tx: number;
}

export const dashboardApi = {
  getDashboard: (orgSlug: string) =>
    apiClient.get<{ data: DashboardData }>(`/orgs/${orgSlug}/dashboard`),

  getMetricsOverview: (orgSlug: string) =>
    apiClient.get<{ data: MetricsOverview }>(`/orgs/${orgSlug}/metrics/overview`),
};
