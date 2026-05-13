import { getToken } from './api';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:20128';

// Provider types
export interface Provider {
  id: string;
  name: string;
  type: string;
  base_url: string;
  enabled: boolean;
  created_at: string;
  updated_at: string;
}

export interface CreateProviderRequest {
  name: string;
  type: string;
  base_url: string;
  api_key: string;
}

// API Key types
export interface ApiKey {
  id: string;
  name: string;
  key?: string; // Only present on creation
  created_at: string;
  last_used_at: string | null;
}

export interface CreateKeyRequest {
  name: string;
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

  return response.json();
}

// Provider API
export async function listProviders(): Promise<Provider[]> {
  return authFetch<Provider[]>('/api/admin/providers');
}

export async function createProvider(data: CreateProviderRequest): Promise<Provider> {
  return authFetch<Provider>('/api/admin/providers', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export async function updateProvider(id: string, data: Partial<CreateProviderRequest & { enabled: boolean }>): Promise<Provider> {
  return authFetch<Provider>(`/api/admin/providers/${id}`, {
    method: 'PATCH',
    body: JSON.stringify(data),
  });
}

export async function deleteProvider(id: string): Promise<void> {
  await authFetch<void>(`/api/admin/providers/${id}`, {
    method: 'DELETE',
  });
}

// API Key API
export async function listKeys(): Promise<ApiKey[]> {
  return authFetch<ApiKey[]>('/api/admin/keys');
}

export async function createKey(name: string): Promise<ApiKey> {
  return authFetch<ApiKey>('/api/admin/keys', {
    method: 'POST',
    body: JSON.stringify({ name }),
  });
}

export async function deleteKey(id: string): Promise<void> {
  await authFetch<void>(`/api/admin/keys/${id}`, {
    method: 'DELETE',
  });
}

// Combo types
export interface Combo {
  name: string;
  models: string[];
  strategy: string;
  sticky_limit: number;
  created_at: string;
  updated_at: string;
}

export interface CreateComboRequest {
  name: string;
  models: string[];
  strategy: string;
  sticky_limit?: number;
}

export interface UpdateComboRequest {
  models?: string[];
  strategy?: string;
  sticky_limit?: number;
}

// Combo API
export async function listCombos(): Promise<Combo[]> {
  return authFetch<Combo[]>('/api/admin/combos');
}

export async function createCombo(data: CreateComboRequest): Promise<Combo> {
  return authFetch<Combo>('/api/admin/combos', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export async function updateCombo(name: string, data: UpdateComboRequest): Promise<Combo> {
  return authFetch<Combo>(`/api/admin/combos/${encodeURIComponent(name)}`, {
    method: 'PATCH',
    body: JSON.stringify(data),
  });
}

export async function deleteCombo(name: string): Promise<void> {
  await authFetch<void>(`/api/admin/combos/${encodeURIComponent(name)}`, {
    method: 'DELETE',
  });
}

// Node types
export interface Node {
  id: string;
  name: string;
  base_url: string;
  format: string;
  created_at: string;
  updated_at: string;
}

export interface CreateNodeRequest {
  name: string;
  base_url: string;
  api_key: string;
  format: string;
}

export interface TestNodeResult {
  success: boolean;
  message: string;
}

// Node API
export async function listNodes(): Promise<Node[]> {
  return authFetch<Node[]>('/api/admin/nodes');
}

export async function createNode(data: CreateNodeRequest): Promise<Node> {
  return authFetch<Node>('/api/admin/nodes', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export async function testNode(id: string): Promise<TestNodeResult> {
  return authFetch<TestNodeResult>(`/api/admin/nodes/${id}/test`, {
    method: 'POST',
  });
}

export async function deleteNode(id: string): Promise<void> {
  await authFetch<void>(`/api/admin/nodes/${id}`, {
    method: 'DELETE',
  });
}

// Alias types
export interface Alias {
  alias: string;
  target: string;
  created_at: string;
}

export interface CreateAliasRequest {
  alias: string;
  target: string;
}

// Alias API
export async function listAliases(): Promise<Alias[]> {
  return authFetch<Alias[]>('/api/admin/aliases');
}

export async function createAlias(data: CreateAliasRequest): Promise<Alias> {
  return authFetch<Alias>('/api/admin/aliases', {
    method: 'POST',
    body: JSON.stringify(data),
  });
}

export async function deleteAlias(alias: string): Promise<void> {
  await authFetch<void>(`/api/admin/aliases/${encodeURIComponent(alias)}`, {
    method: 'DELETE',
  });
}
