const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:20128';

export interface LoginResponse {
  token: string;
  user?: {
    username: string;
  };
}

export interface MeResponse {
  username: string;
}

export async function login(username: string, password: string): Promise<LoginResponse> {
  const response = await fetch(`${API_BASE_URL}/api/login`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ username, password }),
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ message: 'Login failed' }));
    throw new Error(error.message || 'Login failed');
  }

  return response.json();
}

export async function getMe(token: string): Promise<MeResponse> {
  const response = await fetch(`${API_BASE_URL}/api/me`, {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
  });

  if (!response.ok) {
    throw new Error('Failed to get user info');
  }

  return response.json();
}

export function logout(): void {
  if (typeof window !== 'undefined') {
    localStorage.removeItem('token');
    window.location.href = '/login';
  }
}

export function getToken(): string | null {
  if (typeof window !== 'undefined') {
    return localStorage.getItem('token');
  }
  return null;
}

export function setToken(token: string): void {
  if (typeof window !== 'undefined') {
    localStorage.setItem('token', token);
  }
}
