export function isAuthenticated() {
  if (typeof window === 'undefined') return false;
  const token = localStorage.getItem('jwt_token');
  if (!token) return false;
  try {
    const payload = JSON.parse(atob(token.split('.')[1]));
    return payload.exp * 1000 > Date.now();
  } catch {
    return false;
  }
}

export function login(token) {
  localStorage.setItem('jwt_token', token);
}

export function logout() {
  localStorage.removeItem('jwt_token');
  window.location.href = '/login';
}
