const API_BASE = process.env.NEXT_PUBLIC_API_URL || '';

function getToken() {
  if (typeof window === 'undefined') return null;
  return localStorage.getItem('jwt_token');
}

async function authFetch(path, options = {}) {
  const token = getToken();
  const headers = { 'Content-Type': 'application/json', ...options.headers };
  if (token) headers['Authorization'] = 'Bearer ' + token;

  const res = await fetch(API_BASE + path, { ...options, headers });

  if (res.status === 401) {
    if (typeof window !== 'undefined') {
      localStorage.removeItem('jwt_token');
      window.location.href = '/login';
    }
    throw new Error('Unauthorized');
  }

  if (!res.ok) {
    const text = await res.text();
    let msg;
    try {
      msg = JSON.parse(text).error || text;
    } catch {
      msg = text;
    }
    throw new Error(msg || `HTTP ${res.status}`);
  }
  return res;
}

export const chatApi = {
  async listSessions() {
    const res = await authFetch('/api/chat/sessions');
    const data = await res.json();
    return data.sessions || [];
  },

  async createSession(title) {
    const res = await authFetch('/api/chat/sessions', {
      method: 'POST',
      body: JSON.stringify({ title: title || 'New Chat' }),
    });
    return res.json();
  },

  async deleteSession(id) {
    await authFetch(`/api/chat/sessions/${id}`, { method: 'DELETE' });
  },

  async listMessages(sessionId, { limit = 30, before } = {}) {
    let path = `/api/chat/sessions/${sessionId}/messages?limit=${limit}`;
    if (before) path += `&before=${before}`;
    const res = await authFetch(path);
    const data = await res.json();
    return data.messages || [];
  },

  async saveMessage(sessionId, { role, content, artifact_id }) {
    const res = await authFetch(`/api/chat/sessions/${sessionId}/messages`, {
      method: 'POST',
      body: JSON.stringify({ role, content, artifact_id }),
    });
    return res.json();
  },

  streamChat(model, messages, { signal, onChunk, onDone, onError }) {
    const token = getToken();
    const headers = { 'Content-Type': 'application/json' };
    if (token) headers['Authorization'] = 'Bearer ' + token;

    let cancelled = false;

    (async () => {
      try {
        const res = await fetch(API_BASE + '/api/chat/completions', {
          method: 'POST',
          headers,
          body: JSON.stringify({ model, messages, stream: true }),
          signal,
        });

        if (res.status === 401) {
          if (typeof window !== 'undefined') {
            localStorage.removeItem('jwt_token');
            window.location.href = '/login';
          }
          onError?.(new Error('Unauthorized'));
          return;
        }

        if (!res.ok) {
          const text = await res.text();
          let msg;
          try {
            msg = JSON.parse(text).error || text;
          } catch {
            msg = text;
          }
          throw new Error(msg || `HTTP ${res.status}`);
        }

        const reader = res.body.getReader();
        const decoder = new TextDecoder();
        let buffer = '';

        while (!cancelled) {
          const { done, value } = await reader.read();
          if (done) break;

          buffer += decoder.decode(value, { stream: true });
          const lines = buffer.split('\n');
          buffer = lines.pop() || '';

          for (const line of lines) {
            const trimmed = line.trim();
            if (!trimmed || !trimmed.startsWith('data: ')) continue;

            const data = trimmed.slice(6);
            if (data === '[DONE]') {
              if (!cancelled) onDone?.();
              return;
            }

            try {
              const chunk = JSON.parse(data);
              if (chunk.error) {
                if (!cancelled) onError?.(new Error(chunk.error));
                return;
              }
              const content = chunk.choices?.[0]?.delta?.content || '';
              if (content && !cancelled) onChunk?.(content);
              if (chunk.choices?.[0]?.finish_reason === 'stop') {
                if (!cancelled) onDone?.();
              }
            } catch {
            }
          }
        }
        if (!cancelled) onDone?.();
      } catch (err) {
        if (!cancelled) onError?.(err);
      }
    })();

    return () => { cancelled = true; };
  },

  async getArtifact(id) {
    const res = await authFetch(`/api/chat/artifacts/${id}`);
    return res.json();
  },

  async createArtifact(data) {
    const res = await authFetch('/api/chat/artifacts', {
      method: 'POST',
      body: JSON.stringify(data),
    });
    return res.json();
  },

  async listProviderModels(provider) {
    const res = await authFetch(`/api/chat/models?provider=${encodeURIComponent(provider)}`);
    return res.json();
  },

  async uploadFile(sessionId, file) {
    const token = getToken();
    const formData = new FormData();
    formData.append('file', file);
    formData.append('session_id', sessionId);
    const res = await fetch(API_BASE + '/api/chat/files', {
      method: 'POST',
      headers: token ? { 'Authorization': 'Bearer ' + token } : {},
      body: formData,
    });
    if (!res.ok) {
      const text = await res.text();
      throw new Error(text || 'Upload failed');
    }
    return res.json();
  },

  async generateTitle(sessionId) {
    const res = await authFetch(`/api/chat/sessions/${sessionId}/generate-title`, {
      method: 'POST',
    });
    return res.json();
  },

  github: {
    async startAuth() {
      const res = await authFetch('/api/chat/github/auth/start', { method: 'POST' });
      return res.json();
    },

    async pollAuth(deviceCode) {
      const res = await authFetch('/api/chat/github/auth/poll', {
        method: 'POST',
        body: JSON.stringify({ device_code: deviceCode }),
      });
      return res.json();
    },

    async api(method, path, body) {
      const res = await authFetch('/api/chat/github/api', {
        method: 'POST',
        body: JSON.stringify({ method, path, body }),
      });
      return res.json();
    },

    async listRepos() {
      return this.api('GET', '/user/repos?per_page=100&sort=updated');
    },

    async getRepoTree(owner, repo, sha) {
      const path = sha ? `/repos/${owner}/${repo}/git/trees/${sha}?recursive=1` : `/repos/${owner}/${repo}/git/trees/HEAD?recursive=1`;
      return this.api('GET', path);
    },

    async getFileContent(owner, repo, path) {
      return this.api('GET', `/repos/${owner}/${repo}/contents/${path}`);
    },

    async createCommit(owner, repo, data) {
      return this.api('POST', `/repos/${owner}/${repo}/git/commits`, data);
    },

    async createPR(owner, repo, data) {
      return this.api('POST', `/repos/${owner}/${repo}/pulls`, data);
    },

    async getDiff(owner, repo, base, head) {
      const res = await authFetch('/api/chat/github/api', {
        method: 'POST',
        body: JSON.stringify({
          method: 'GET',
          path: `/repos/${owner}/${repo}/compare/${base}...${head}`,
        }),
      });
      return res.json();
    },
  },
};
