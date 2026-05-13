import { getToken } from './api';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:20128';

// Types for analytics data
export interface UsageStats {
  totalTokens: number;
  totalCost: number;
  inputTokens: number;
  outputTokens: number;
  requestCount: number;
  byModel: ModelUsage[];
  byProvider: ProviderUsage[];
  overTime: TimeSeriesPoint[];
}

export interface ModelUsage {
  model: string;
  tokens: number;
  cost: number;
  requests: number;
}

export interface ProviderUsage {
  provider: string;
  tokens: number;
  cost: number;
  requests: number;
}

export interface TimeSeriesPoint {
  date: string;
  tokens: number;
  cost: number;
  requests: number;
}

export interface UsageLog {
  id: string;
  timestamp: string;
  model: string;
  provider: string;
  inputTokens: number;
  outputTokens: number;
  totalTokens: number;
  cost: number;
  status: string;
}

export interface UsageLogFilters {
  start?: string;
  end?: string;
  model?: string;
  provider?: string;
  limit?: number;
  offset?: number;
}

export interface PricingRule {
  id: string;
  model: string;
  inputPricePer1k: number;
  outputPricePer1k: number;
  createdAt: string;
  updatedAt: string;
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

/**
 * Get usage statistics for a date range
 */
export async function getUsageStats(start: string, end: string): Promise<UsageStats> {
  const params = new URLSearchParams({ start, end });
  return authFetch<UsageStats>(`/api/admin/analytics/stats?${params}`);
}

/**
 * Get usage logs with optional filters
 */
export async function getUsageLogs(filters: UsageLogFilters = {}): Promise<UsageLogsResponse> {
  const params = new URLSearchParams();
  if (filters.start) params.append('start', filters.start);
  if (filters.end) params.append('end', filters.end);
  if (filters.model) params.append('model', filters.model);
  if (filters.provider) params.append('provider', filters.provider);
  if (filters.limit) params.append('limit', filters.limit.toString());
  if (filters.offset) params.append('offset', filters.offset.toString());
  
  return authFetch<UsageLogsResponse>(`/api/admin/analytics/logs?${params}`);
}

export interface UsageLogsResponse {
  logs: UsageLog[];
  total: number;
}

/**
 * Get pricing rules
 */
export async function getPricingRules(): Promise<PricingRule[]> {
  return authFetch<PricingRule[]>('/api/admin/analytics/pricing');
}

/**
 * Update a pricing rule
 */
export async function updatePricingRule(id: string, data: Partial<Omit<PricingRule, 'id' | 'createdAt' | 'updatedAt'>>): Promise<PricingRule> {
  return authFetch<PricingRule>(`/api/admin/analytics/pricing/${id}`, {
    method: 'PATCH',
    body: JSON.stringify(data),
  });
}
