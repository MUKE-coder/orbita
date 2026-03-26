import apiClient from "./client";

export interface User {
  id: string;
  email: string;
  name: string;
  avatar_url: string | null;
  is_super_admin: boolean;
  is_email_verified: boolean;
  totp_enabled: boolean;
  created_at: string;
  updated_at: string;
}

export interface LoginResponse {
  user: User;
  access_token: string;
}

export const authApi = {
  register: (data: { email: string; password: string; name: string }) =>
    apiClient.post<{ data: LoginResponse }>("/auth/register", data),

  login: (data: { email: string; password: string }) =>
    apiClient.post<{ data: LoginResponse }>("/auth/login", data),

  logout: () => apiClient.post("/auth/logout"),

  refresh: () => apiClient.post<{ data: LoginResponse }>("/auth/refresh"),

  forgotPassword: (email: string) =>
    apiClient.post("/auth/forgot-password", { email }),

  resetPassword: (data: {
    email: string;
    otp: string;
    new_password: string;
  }) => apiClient.post("/auth/reset-password", data),

  verifyEmail: (token: string) =>
    apiClient.post("/auth/verify-email", { token }),

  getProfile: () => apiClient.get<{ data: User }>("/me"),

  updateProfile: (data: { name?: string; avatar_url?: string | null }) =>
    apiClient.put<{ data: User }>("/me", data),

  changePassword: (data: {
    current_password: string;
    new_password: string;
  }) => apiClient.post("/me/change-password", data),

  getSessions: () =>
    apiClient.get<{
      data: Array<{
        id: string;
        device_info: string | null;
        ip_address: string | null;
        expires_at: string;
        created_at: string;
      }>;
    }>("/me/sessions"),

  revokeSession: (id: string) => apiClient.delete(`/me/sessions/${id}`),

  getApiKeys: () =>
    apiClient.get<{
      data: Array<{
        id: string;
        name: string;
        key_prefix: string;
        scopes: string[];
        last_used_at: string | null;
        created_at: string;
      }>;
    }>("/me/api-keys"),

  createApiKey: (data: { name: string; scopes: string[] }) =>
    apiClient.post<{
      data: { api_key: { id: string; name: string }; key: string };
    }>("/me/api-keys", data),

  deleteApiKey: (id: string) => apiClient.delete(`/me/api-keys/${id}`),
};
