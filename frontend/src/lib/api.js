const API_BASE = process.env.NEXT_PUBLIC_API_URL || '';

async function request(method, path, body) {
  const token = typeof window !== 'undefined' ? localStorage.getItem('jwt_token') : null;
  const headers = { 'Content-Type': 'application/json' };
  if (token) headers['Authorization'] = 'Bearer ' + token;

  const opts = { method, headers };
  if (body) opts.body = JSON.stringify(body);

  const res = await fetch(API_BASE + path, opts);

  if (res.status === 401 && !path.startsWith('/v1/') && path !== '/api/login') {
    if (typeof window !== 'undefined') {
      localStorage.removeItem('jwt_token');
      window.location.href = '/login';
    }
    throw new Error('Unauthorized');
  }

  const text = await res.text();
  let data;
  try { data = JSON.parse(text); } catch { data = text; }

  if (!res.ok) {
    throw new Error(data?.error || data?.message || `HTTP ${res.status}`);
  }
  return data;
}

export const api = {
  _request: request,
  login:         (username, password) => request('POST', '/api/login', { username, password }),
  health:        ()         => request('GET',  '/health'),
  models:        ()         => request('GET',  '/v1/models'),
  accounts:      ()         => request('GET',  '/api/admin/accounts'),
  createAccount: (d)        => request('POST', '/api/admin/accounts', d),
  deleteAccount: (id)       => request('DELETE', `/api/admin/accounts/${id}`),
  keys:          ()         => request('GET',  '/api/admin/keys'),
  createKey:     (d)        => request('POST', '/api/admin/keys', d),
  deleteKey:     (id)       => request('DELETE', `/api/admin/keys/${id}`),
  proxies:       ()         => request('GET',  '/api/proxies'),
  proxySettings: ()         => request('GET',  '/api/proxy/settings'),
  updateProxy:   (d)        => request('POST', '/api/proxy/settings', d),
  proxyPools:    ()         => request('GET',  '/api/proxy-pools'),
  createPool:    (d)        => request('POST', '/api/proxy-pools', d),
  deletePool:    (id)       => request('DELETE', `/api/proxy-pools/${id}`),
  testPool:      (id)       => request('POST', `/api/proxy-pools/${id}/test`),
  testAllPools:  ()         => request('POST', '/api/proxy-pools/test-all'),
  addProxy:      (d)        => request('POST', '/api/proxies', d),
  deleteProxy:   (id)       => request('DELETE', `/api/proxies/${id}`),
  chat:          (model, messages) =>
    request('POST', '/v1/chat/completions', { model, messages, stream: false }),
  usage:         ()         => request('GET',  '/api/admin/usage/stats'),
  usageLogs:     (params)   => request('GET',  '/api/admin/usage?' + new URLSearchParams(params || {})),
  me:            ()         => request('GET',  '/api/me'),
  settings:      ()         => request('GET',  '/api/admin/settings'),
};
