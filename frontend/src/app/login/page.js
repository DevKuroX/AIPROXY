'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { api } from '@/lib/api';
import { login as authLogin } from '@/lib/auth';
import { MdHub } from 'react-icons/md';

export default function LoginPage() {
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const router = useRouter();

  async function handleSubmit(e) {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      const data = await api.login('admin', password);
      if (data.token) {
        authLogin(data.token);
        router.push('/dashboard');
      } else {
        throw new Error('No token received');
      }
    } catch (err) {
      setError(err.message || 'Login failed');
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="min-h-screen flex flex-col relative bg-bg transition-colors duration-500 overflow-x-hidden">
      <div className="fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[800px] h-[800px] bg-primary/5 rounded-full blur-[100px] pointer-events-none z-0" />
      <div className="fixed bottom-0 right-0 w-[600px] h-[600px] bg-orange-200/20 dark:bg-orange-900/10 rounded-full blur-[120px] pointer-events-none z-0 translate-y-1/3 translate-x-1/3" />

      <main className="flex-1 flex flex-col items-center justify-center p-4 sm:p-6 z-10 w-full">
        <div className="w-full max-w-sm card-elev p-8">
          <div className="flex flex-col items-center mb-8">
            <div className="flex items-center justify-center size-12 rounded-xl shadow-warm bg-gradient-to-br from-brand-500 to-brand-700">
              <MdHub className="text-white text-[24px]" />
            </div>
            <h1 className="text-xl font-bold mt-4 text-text-main">AIPROXY</h1>
            <p className="text-sm mt-1 text-text-muted">Dashboard Login</p>
          </div>

          <form onSubmit={handleSubmit} className="flex flex-col gap-4">
            <div>
              <label className="block text-sm font-medium mb-1.5 text-text-muted">Password</label>
              <input type="password" value={password} onChange={e => setPassword(e.target.value)}
                className="w-full px-3 py-2.5 rounded text-sm outline-none bg-bg-alt border border-border-subtle text-text-main focus:border-primary focus:shadow-focus transition-colors"
                placeholder="Enter admin password" autoFocus />
            </div>
            {error && <p className="text-sm text-danger">{error}</p>}
            <button type="submit" disabled={loading}
              className="w-full py-2.5 rounded text-sm font-semibold transition-all active:scale-[0.97] disabled:opacity-50 bg-primary text-white hover:bg-primary-hover">
              {loading ? 'Signing in...' : 'Sign In'}
            </button>
          </form>
        </div>
        <p className="text-xs mt-6 text-text-subtle">AIPROXY AI Gateway</p>
      </main>
    </div>
  );
}
