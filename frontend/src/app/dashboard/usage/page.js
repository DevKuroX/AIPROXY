'use client';

import { useState, useEffect, useCallback } from 'react';
import { api } from '@/lib/api';
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';

import { MdRefresh } from 'react-icons/md';

const fmt = (n) => (n != null ? new Intl.NumberFormat().format(n) : '—');
const fmtTokens = (n) => {
  if (n >= 1000000) return `${(n / 1000000).toFixed(1)}M`;
  if (n >= 1000) return `${(n / 1000).toFixed(1)}K`;
  return String(n || 0);
};
const fmtTime = (iso) => {
  if (!iso) return 'Never';
  const d = new Date(iso);
  const now = new Date();
  const diff = Math.floor((now - d) / 1000);
  if (diff < 60) return `${diff}s ago`;
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`;
  return d.toLocaleDateString();
};

const PERIODS = [{ value: '24h', label: '24H' }, { value: '7d', label: '7D' }, { value: '30d', label: '30D' }];
const TABLE_OPTIONS = [
  { value: 'model', label: 'Usage by Model' },
  { value: 'account', label: 'Usage by Account' },
  { value: 'apiKey', label: 'Usage by API Key' },
  { value: 'endpoint', label: 'Usage by Endpoint' },
];

const fmtCost = (n) => `$${(n || 0).toFixed(4)}`;

function RecentRequests({ logs = [] }) {
  return (
    <div className="card-soft flex min-w-0 flex-col overflow-hidden" style={{ height: 340 }}>
      <div className="px-4 py-3 border-b border-border shrink-0">
        <span className="text-xs font-semibold text-text-muted uppercase tracking-wide">Recent Requests</span>
      </div>
      <div className="flex-1 overflow-y-auto">
        {logs.length === 0 ? (
          <div className="flex items-center justify-center h-full text-sm text-text-muted">No requests yet.</div>
        ) : (
          <table className="w-full min-w-[300px] border-collapse text-xs">
            <thead className="sticky top-0 bg-bg z-10">
              <tr className="border-b border-border">
                <th className="py-2 pl-4 text-left font-semibold text-text-muted w-2"></th>
                <th className="py-2 text-left font-semibold text-text-muted">Model</th>
                <th className="py-2 text-right font-semibold text-text-muted whitespace-nowrap pr-4">In / Out</th>
                <th className="py-2 text-right font-semibold text-text-muted pr-4">When</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border/50">
              {logs.map((r, i) => {
                const ok = !r.status || r.status === 'ok' || r.status === 'success';
                return (
                  <tr key={r.id || i} className="hover:bg-bg-alt transition-colors">
                    <td className="py-2 pl-4"><span className={`block w-1.5 h-1.5 rounded-full ${ok ? 'bg-success' : 'bg-danger'}`} /></td>
                    <td className="py-2 font-mono truncate max-w-[120px] text-text-main" title={r.model}>{r.model || '-'}</td>
                    <td className="py-2 text-right whitespace-nowrap pr-4">
                      <span className="text-primary">{fmtTokens(r.prompt_tokens || r.input_tokens || 0)}↑</span>{' '}
                      <span className="text-success">{fmtTokens(r.completion_tokens || r.output_tokens || 0)}↓</span>
                    </td>
                    <td className="py-2 text-right text-text-muted whitespace-nowrap pr-4">{fmtTime(r.timestamp || r.created_at)}</td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}

function ProviderTopology({ models = [], providers = [] }) {
  const all = providers.length > 0 ? providers : models.map(m => ({ provider: m.model?.split('/')[0] || '?', tokens: m.tokens }));
  const top = all.slice(0, 5);
  const colors = ['bg-accent','bg-primary','bg-info','bg-success','bg-warning'];
  return (
    <div className="card-soft flex items-center justify-center p-6" style={{ height: 340 }}>
      {top.length === 0 ? (
        <span className="text-sm text-text-muted">No providers active</span>
      ) : (
        <div className="flex flex-wrap gap-6 items-center justify-center">
          {top.map((p, i) => (
            <div key={p.provider || i} className="flex flex-col items-center gap-1.5">
              <div className={`flex items-center justify-center w-10 h-10 rounded-xl text-white text-sm font-bold ${colors[i % colors.length]}`}>
                {(p.provider || '?')[0].toUpperCase()}
              </div>
              <span className="text-[10px] font-mono text-text-muted">{p.provider}</span>
              <span className="text-[10px] font-medium text-success">Active</span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

export default function UsagePage() {
  const [period, setPeriod] = useState('7d');
  const [stats, setStats] = useState(null);
  const [logs, setLogs] = useState([]);
  const [loading, setLoading] = useState(true);
  const [chartView, setChartView] = useState('tokens');
  const [tableView, setTableView] = useState('model');
  const [tableMode, setTableMode] = useState('costs');
  const [chartData, setChartData] = useState([]);
  const [error, setError] = useState(null);

  const load = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [statsData, logsData] = await Promise.all([
        api.usage().catch((err) => { setError(err.message || 'Failed to load stats'); return null; }),
        api.usageLogs({ limit: 50 }).catch((err) => { setError(err.message || 'Failed to load logs'); return { data: [], total: 0 }; }),
      ]);
      setStats(statsData);
      const l = logsData.data || [];
      setLogs(l);

      // Build chart data from logs
      const days = {};
      l.forEach(r => {
        const ts = r.timestamp || r.created_at;
        if (!ts) return;
        const day = new Date(ts).toLocaleDateString('en-US', { weekday: 'short', month: 'short', day: 'numeric' });
        if (!days[day]) days[day] = { tokens: 0, cost: 0 };
        days[day].tokens += (r.total_tokens || r.prompt_tokens || 0) + (r.completion_tokens || r.output_tokens || 0);
        days[day].cost += r.cost_usd || r.cost || 0;
      });
      setChartData(Object.entries(days).map(([label, v]) => ({ label, tokens: v.tokens, cost: v.cost })));
    } catch {} finally { setLoading(false); }
  }, [period]);

  useEffect(() => { load(); }, [load]);

  const models = stats?.by_model ? Object.entries(stats.by_model).map(([k, v]) => ({ model: k, tokens: v.tokens, cost: v.cost })) : [];
  const providers = stats?.by_provider ? Object.entries(stats.by_provider).map(([k, v]) => ({ provider: k, tokens: v.tokens, cost: v.cost })) : [];
  const totalReq = logs.length || models.reduce((s, m) => s + (m.tokens > 0 ? 1 : 0), 0);

  const detailData = models.map(m => ({
    rawModel: m.model, provider: m.model?.split('/')[0] || '?',
    requests: m.tokens > 0 ? Math.round(m.tokens / 1500) || 1 : 0,
    cost: m.cost,
    lastUsed: logs.find(l => l.model === m.model)?.timestamp || null,
  }));

  return (
    <div className="flex min-w-0 flex-col gap-4 sm:gap-6">
      {loading ? (
        <div className="h-48 flex items-center justify-center text-sm text-text-subtle">Loading...</div>
      ) : error ? (
        <div className="h-48 flex flex-col items-center justify-center gap-4">
          <p className="text-sm text-danger">{error}</p>
          <button onClick={() => load()} className="inline-flex items-center gap-1.5 px-4 py-2 rounded-lg bg-danger/10 text-danger text-sm font-medium hover:bg-danger/20 transition-colors">
            <MdRefresh /> Retry
          </button>
        </div>
      ) : (
        <>
          <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
            <h2 className="text-sm font-semibold text-text-main">Usage & Analytics</h2>
            <div className="flex bg-bg-alt rounded-lg border border-border-subtle p-0.5 self-start">
              {PERIODS.map(p => (
                <button key={p.value} onClick={() => setPeriod(p.value)}
                  className={`shrink-0 px-4 rounded-[8px] font-medium transition-all h-7 text-xs ${period === p.value ? 'bg-surface text-text-main shadow-sm' : 'text-text-muted hover:text-text-main'}`}>
                  {p.label}
                </button>
              ))}
            </div>
          </div>

          <div className="grid min-w-0 grid-cols-1 gap-3 sm:grid-cols-2 md:grid-cols-4 sm:gap-4">
            <div className="card-soft flex min-w-0 flex-col gap-1 px-4 py-3">
              <span className="text-text-muted text-sm uppercase font-semibold">Total Requests</span>
              <span className="truncate text-2xl font-bold text-text-main">{fmt(totalReq)}</span>
            </div>
            <div className="card-soft flex min-w-0 flex-col gap-1 px-4 py-3">
              <span className="text-text-muted text-sm uppercase font-semibold">Total Tokens</span>
              <span className="truncate text-2xl font-bold text-primary">{fmtTokens(stats?.total_tokens || 0)}</span>
            </div>
            <div className="card-soft flex min-w-0 flex-col gap-1 px-4 py-3">
              <span className="text-text-muted text-sm uppercase font-semibold">Models Used</span>
              <span className="truncate text-2xl font-bold text-success">{models.length}</span>
            </div>
            <div className="card-soft flex min-w-0 flex-col gap-1 px-4 py-3">
              <span className="text-text-muted text-sm uppercase font-semibold">Est. Cost</span>
              <span className="truncate text-2xl font-bold text-warning">${(stats?.total_cost || 0).toFixed(6)}</span>
              <span className="text-[10px] text-text-muted">Estimated, not actual billing</span>
            </div>
          </div>

          <div className="card-soft p-3 sm:p-4">
            <div className="grid w-full grid-cols-2 items-center gap-1 rounded-lg border border-border bg-bg-alt p-1 sm:w-auto sm:self-start mb-3">
              <button onClick={() => setChartView('tokens')}
                className={`px-3 py-1 rounded-md text-sm font-medium transition-colors ${chartView === 'tokens' ? 'bg-primary text-white shadow-sm' : 'text-text-muted hover:text-text-main'}`}>Tokens</button>
              <button onClick={() => setChartView('cost')}
                className={`px-3 py-1 rounded-md text-sm font-medium transition-colors ${chartView === 'cost' ? 'bg-primary text-white shadow-sm' : 'text-text-muted hover:text-text-main'}`}>Cost</button>
            </div>
            <ResponsiveContainer width="100%" height={220}>
              <AreaChart data={chartData.length > 0 ? chartData : [{label:'No data', tokens:0, cost:0}]} margin={{ top: 4, right: 8, left: 0, bottom: 0 }}>
                <defs>
                  <linearGradient id="gt" x1="0" y1="0" x2="0" y2="1"><stop offset="5%" stopColor="#6366f1" stopOpacity={0.25} /><stop offset="95%" stopColor="#6366f1" stopOpacity={0} /></linearGradient>
                  <linearGradient id="gc" x1="0" y1="0" x2="0" y2="1"><stop offset="5%" stopColor="#f59e0b" stopOpacity={0.25} /><stop offset="95%" stopColor="#f59e0b" stopOpacity={0} /></linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" strokeOpacity={0.1} />
                <XAxis dataKey="label" tick={{fontSize:10,fill:'currentColor',fillOpacity:0.5}} tickLine={false} axisLine={false} />
                <YAxis tick={{fontSize:10,fill:'currentColor',fillOpacity:0.5}} tickLine={false} axisLine={false} tickFormatter={chartView==='tokens'?fmtTokens:fmtCost} width={50} />
                <Tooltip contentStyle={{background:'var(--color-surface)',border:'1px solid var(--color-border)',borderRadius:'var(--radius-md)',fontSize:12}}
                  formatter={(value,name) => [name==='tokens'?fmtTokens(value):`$${value.toFixed(6)}`, name==='tokens'?'Tokens':'Cost']} />
                <Area type="monotone" dataKey={chartView==='tokens'?'tokens':'cost'} stroke={chartView==='tokens'?'#6366f1':'#f59e0b'} fill={chartView==='tokens'?'url(#gt)':'url(#gc)'} strokeWidth={2} dot={false} activeDot={{r:4}} />
              </AreaChart>
            </ResponsiveContainer>
          </div>

          <div className="grid min-w-0 grid-cols-1 items-stretch gap-2 lg:grid-cols-[minmax(0,2fr)_minmax(280px,1fr)]">
            <RecentRequests logs={logs} />
            <ProviderTopology models={models} providers={providers} />
          </div>

          <div className="flex flex-col gap-3">
            <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
              <select value={tableView} onChange={e => setTableView(e.target.value)}
                className="w-full rounded-lg border border-border bg-surface px-3 py-1.5 text-sm font-medium text-text-main focus:outline-none focus:ring-2 focus:ring-primary/50 sm:w-auto"
                style={{ colorScheme: 'auto' }}>
                {TABLE_OPTIONS.map(o => <option key={o.value} value={o.value}>{o.label}</option>)}
              </select>
              <div className="grid grid-cols-2 items-center gap-1 rounded-lg border border-border bg-bg-alt p-1 sm:flex">
                <button onClick={() => setTableMode('costs')}
                  className={`px-3 py-1 rounded-md text-sm font-medium transition-colors ${tableMode === 'costs' ? 'bg-primary text-white shadow-sm' : 'text-text-muted hover:text-text-main'}`}>Costs</button>
                <button onClick={() => setTableMode('tokens')}
                  className={`px-3 py-1 rounded-md text-sm font-medium transition-colors ${tableMode === 'tokens' ? 'bg-primary text-white shadow-sm' : 'text-text-muted hover:text-text-main'}`}>Tokens</button>
              </div>
            </div>
            <div className="overflow-x-auto rounded-lg border border-border-subtle">
              <table className="w-full">
                <thead>
                  <tr className="bg-bg-alt">
                    <th className="px-4 py-3 text-xs font-semibold uppercase text-text-muted text-left">Model</th>
                    <th className="px-4 py-3 text-xs font-semibold uppercase text-text-muted text-left">Provider</th>
                    <th className="px-4 py-3 text-xs font-semibold uppercase text-text-muted text-right">Requests</th>
                    <th className="px-4 py-3 text-xs font-semibold uppercase text-text-muted text-right">{tableMode === 'costs' ? 'Cost' : 'Tokens'}</th>
                    <th className="px-4 py-3 text-xs font-semibold uppercase text-text-muted text-right">Last Used</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border/50">
                  {detailData.length === 0 ? (
                    <tr><td colSpan={5} className="text-center py-8 text-sm text-text-subtle">No usage recorded yet.</td></tr>
                  ) : detailData.map((item, i) => (
                    <tr key={i} className="hover:bg-bg-alt transition-colors">
                      <td className="px-4 py-3 text-sm font-mono text-text-main">{item.rawModel}</td>
                      <td className="px-4 py-3"><span className="inline-flex px-2 py-0.5 rounded-full text-[10px] font-semibold bg-primary/10 text-primary">{item.provider}</span></td>
                      <td className="px-4 py-3 text-sm text-right text-text-muted">{fmt(item.requests)}</td>
                      <td className="px-4 py-3 text-sm text-right text-warning">{tableMode === 'costs' ? `$${(item.cost || 0).toFixed(6)}` : fmtTokens(item.cost * 100000 || 0)}</td>
                      <td className="px-4 py-3 text-sm text-right text-text-muted whitespace-nowrap">{fmtTime(item.lastUsed)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </>
      )}
    </div>
  );
}
