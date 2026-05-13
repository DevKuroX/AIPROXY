import { getToken } from './api';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:20128';

// OAuth Account types
export interface OAuthAccount {
  id: string;
  provider: string;
  email: string;
  status: 'active' | 'expired' | 'error';
  created_at: string;
  last_used_at: string | null;
}

// Device code flow types
export interface DeviceCodeInfo {
  device_code: string;
  user_code: string;
  verification_url: string;
  expires_in: number;
  interval: number;
}

export interface PollResult {
  status: 'pending' | 'completed' | 'expired' | 'error';
  account?: OAuthAccount;
  error?: string;
}

// Helper for authenticated requests
async function authFetch<T>(path: string, options: RequestInit = {}): Promise<T> {
  const token = getToken();
  if (!token) {
    throw new Error('Not authenticated');
  }

  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...options,
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
      ...options.headers,
    },
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ message: 'Request failed' }));
    throw new Error(error.message || error.error || 'Request failed');
  }

  // Handle 204 No Content
  if (response.status === 204) {
    return {} as T;
  }

  return response.json();
}

// List connected OAuth accounts
export async function listOAuthAccounts(): Promise<OAuthAccount[]> {
  return authFetch<OAuthAccount[]>('/api/admin/oauth/accounts');
}

// Start OAuth device code flow
export async function startOAuthFlow(provider: string): Promise<DeviceCodeInfo> {
  return authFetch<DeviceCodeInfo>('/api/admin/oauth/start', {
    method: 'POST',
    body: JSON.stringify({ provider }),
  });
}

// Poll for OAuth completion
export async function pollOAuthStatus(deviceCode: string): Promise<PollResult> {
  return authFetch<PollResult>(`/api/admin/oauth/poll?deviceCode=${encodeURIComponent(deviceCode)}`);
}

// Disconnect an OAuth account
export async function disconnectOAuth(providerId: string, accountId: string): Promise<void> {
  await authFetch<void>(`/api/admin/oauth/${providerId}/${accountId}`, {
    method: 'DELETE',
  });
}
