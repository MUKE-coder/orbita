import apiClient from "./client";

export interface ManagedDatabase {
  id: string;
  environment_id: string;
  organization_id: string;
  name: string;
  engine: string;
  version: string;
  volume_name: string | null;
  status: string;
  port: number | null;
  cpu_limit: number;
  memory_limit: number;
  created_at: string;
}

export interface Backup {
  id: string;
  source_id: string;
  source_type: string;
  status: string;
  size_bytes: number;
  storage_path: string | null;
  created_at: string;
  expires_at: string | null;
}

export interface BackupSchedule {
  id: string;
  frequency: string;
  retention_count: number;
  enabled: boolean;
  last_run_at: string | null;
  next_run_at: string | null;
}

export const databasesApi = {
  list: (orgSlug: string) =>
    apiClient.get<{ data: ManagedDatabase[] }>(`/orgs/${orgSlug}/databases`),

  create: (
    orgSlug: string,
    data: {
      name: string;
      engine: string;
      version: string;
      environment_id: string;
      cpu_limit?: number;
      memory_limit?: number;
    }
  ) =>
    apiClient.post<{ data: ManagedDatabase }>(
      `/orgs/${orgSlug}/databases`,
      data
    ),

  get: (orgSlug: string, dbId: string, showConnection?: boolean) =>
    apiClient.get<{
      data: { database: ManagedDatabase; connection_string?: string };
    }>(
      `/orgs/${orgSlug}/databases/${dbId}${showConnection ? "?show_connection=true" : ""}`
    ),

  delete: (orgSlug: string, dbId: string) =>
    apiClient.delete(`/orgs/${orgSlug}/databases/${dbId}`),

  restart: (orgSlug: string, dbId: string) =>
    apiClient.post(`/orgs/${orgSlug}/databases/${dbId}/restart`),

  stop: (orgSlug: string, dbId: string) =>
    apiClient.post(`/orgs/${orgSlug}/databases/${dbId}/stop`),

  start: (orgSlug: string, dbId: string) =>
    apiClient.post(`/orgs/${orgSlug}/databases/${dbId}/start`),

  createBackup: (orgSlug: string, dbId: string) =>
    apiClient.post<{ data: Backup }>(
      `/orgs/${orgSlug}/databases/${dbId}/backups`
    ),

  listBackups: (orgSlug: string, dbId: string) =>
    apiClient.get<{ data: Backup[] }>(
      `/orgs/${orgSlug}/databases/${dbId}/backups`
    ),

  restoreBackup: (orgSlug: string, dbId: string, backupId: string) =>
    apiClient.post(
      `/orgs/${orgSlug}/databases/${dbId}/backups/${backupId}/restore`
    ),

  getBackupSchedule: (orgSlug: string, dbId: string) =>
    apiClient.get<{ data: BackupSchedule | null }>(
      `/orgs/${orgSlug}/databases/${dbId}/backup-schedule`
    ),

  setBackupSchedule: (
    orgSlug: string,
    dbId: string,
    data: { frequency: string; retention_count: number }
  ) =>
    apiClient.put<{ data: BackupSchedule }>(
      `/orgs/${orgSlug}/databases/${dbId}/backup-schedule`,
      data
    ),
};
