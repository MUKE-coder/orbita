import apiClient from "./client";

export interface Notification {
  id: string;
  user_id: string;
  organization_id: string | null;
  type: string;
  title: string;
  body: string | null;
  read: boolean;
  created_at: string;
}

export interface AuditLog {
  id: string;
  organization_id: string | null;
  user_id: string | null;
  action: string;
  resource_type: string | null;
  resource_id: string | null;
  metadata: Record<string, unknown>;
  ip: string | null;
  created_at: string;
}

export const notificationsApi = {
  list: (orgSlug: string) =>
    apiClient.get<{
      data: { notifications: Notification[]; unread_count: number };
    }>(`/orgs/${orgSlug}/notifications`),

  markRead: (orgSlug: string, id: string) =>
    apiClient.put(`/orgs/${orgSlug}/notifications/${id}/read`),

  markAllRead: (orgSlug: string) =>
    apiClient.put(`/orgs/${orgSlug}/notifications/read-all`),

  getSettings: (orgSlug: string) =>
    apiClient.get(`/orgs/${orgSlug}/notification-settings`),

  updateSettings: (
    orgSlug: string,
    data: {
      event_type: string;
      email_enabled: boolean;
      webhook_enabled: boolean;
      webhook_url?: string;
    }
  ) => apiClient.put(`/orgs/${orgSlug}/notification-settings`, data),

  listAuditLogs: (orgSlug: string, page?: number, pageSize?: number) =>
    apiClient.get<{
      data: {
        audit_logs: AuditLog[];
        total: number;
        page: number;
        page_size: number;
      };
    }>(
      `/orgs/${orgSlug}/audit-logs?page=${page || 1}&page_size=${pageSize || 50}`
    ),
};
