'use client';

import { useState, useEffect } from 'react';
import { api } from '@/lib/api';
import { MdPeople, MdCheckCircle, MdWarning, MdError } from 'react-icons/md';

const statusIcons = {
  active:       { icon: MdCheckCircle, color: 'text-success', bg: 'bg-success/10' },
  depleting:    { icon: MdWarning, color: 'text-warning', bg: 'bg-warning/10' },
  rate_limited: { icon: MdWarning, color: 'text-orange-500', bg: 'bg-orange-500/10' },
  exhausted:    { icon: MdError, color: 'text-danger', bg: 'bg-danger/10' },
};

export default function AccountsPage() {
  const [accounts, setAccounts] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    api.accounts()
      .then(data => setAccounts(data?.accounts || data || []))
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  // Group by provider
  const grouped = {};
  for (const a of accounts) {
    const prov = a.provider_name || a.provider_id || 'unknown';
    if (!grouped[prov]) grouped[prov] = [];
    grouped[prov].push(a);
  }

  const providerKeys = Object.keys(grouped).sort();

  return (
    <div className="fade-in">
      <div className="flex items-center justify-between mb-6">
        <div className="text-xs font-semibold text-text-muted/60 uppercase tracking-wider">Account Pool</div>
        <span className="text-sm text-text-muted">{accounts.length} accounts</span>
      </div>

      {loading ? (
        <div className="text-center py-16 text-sm text-text-subtle">Loading...</div>
      ) : providerKeys.length === 0 ? (
        <div className="text-center py-16 text-sm text-text-subtle">No accounts found</div>
      ) : (
        <div className="space-y-6">
          {providerKeys.map(prov => (
            <div key={prov}>
              <div className="flex items-center gap-2 mb-3">
                <MdPeople className="text-text-muted text-sm" />
                <h2 className="text-sm font-semibold text-text-main">{prov}</h2>
                <span className="text-xs text-text-muted">{grouped[prov].length} accounts</span>
              </div>
              <div className="overflow-x-auto rounded-lg border border-border-subtle">
                <table className="w-full">
                  <thead>
                    <tr className="bg-bg-alt">
                      <th className="text-left px-4 py-3 text-xs font-semibold uppercase tracking-wider text-text-muted">ID</th>
                      <th className="text-left px-4 py-3 text-xs font-semibold uppercase tracking-wider text-text-muted">Status</th>
                      <th className="text-left px-4 py-3 text-xs font-semibold uppercase tracking-wider text-text-muted">Credit</th>
                    </tr>
                  </thead>
                  <tbody>
                    {grouped[prov].map(a => {
                      const status = (a.status || 'unknown').toLowerCase();
                      const s = statusIcons[status] || { icon: MdError, color: 'text-text-muted', bg: 'bg-text-muted/10' };
                      const Icon = s.icon;
                      return (
                        <tr key={a.id} className="border-t border-border">
                          <td className="px-4 py-3 text-sm font-mono text-text-main">{a.name || a.id}</td>
                          <td className="px-4 py-3">
                            <span className={`inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-semibold ${s.bg} ${s.color}`}>
                              <Icon className="text-sm" />
                              {status}
                            </span>
                          </td>
                          <td className="px-4 py-3 text-sm text-text-muted">{a.credit != null ? a.credit + '%' : '—'}</td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
