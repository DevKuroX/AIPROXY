'use client';

import { useState, useEffect, useCallback } from 'react';
import { api } from '@/lib/api';
import { toast } from '@/components/Toast';
import { MdCloudDownload, MdScience, MdCheck, MdWarning, MdClose, MdSettings, MdDeleteSweep, MdAdd } from 'react-icons/md';

function fmtTime(t) {
  if (!t) return '-';
  try { return new Date(t).toLocaleString(); } catch { return '-'; }
}

function parseURL(url) {
  if (!url) return { host: '-', port: '-', type: '-' };
  try {
    const u = new URL(url);
    return { host: u.hostname, port: u.port || '-', type: u.protocol.replace(':', '') };
  } catch {
    const p = url.split('@').pop() || url;
    const [h, port] = p.split(':');
    return { host: h || '-', port: port || '-', type: url.startsWith('socks5') ? 'socks5' : url.startsWith('socks4') ? 'socks4' : 'http' };
  }
}

export default function ProxyPoolsPage() {
  const [pools, setPools] = useState([]);
  const [settings, setSettings] = useState({ enabled: false, webshare_api_key: '', max_latency_ms: 5000 });
  const [loading, setLoading] = useState(true);
  const [fetching, setFetching] = useState(false);
  const [testing, setTesting] = useState(null);
  const [testingAll, setTestingAll] = useState(false);
  const [selected, setSelected] = useState(new Set());
  const [search, setSearch] = useState('');
  const [showSettings, setShowSettings] = useState(false);
  const [showImport, setShowImport] = useState(false);
  const [importText, setImportText] = useState('');
  const [importing, setImporting] = useState(false);

  const load = useCallback(async () => {
    try {
      const [p, s] = await Promise.all([
        api.proxyPools().catch(() => ({ proxy_pools: [] })),
        api.proxySettings().catch(() => ({})),
      ]);
      setPools(p.proxy_pools || []);
      setSettings(prev => ({ ...prev, ...s }));
    } catch {} finally { setLoading(false); }
  }, []);

  useEffect(() => { load(); }, [load]);

  async function handleDelete(id) {
    try { await api.deletePool(id); load(); } catch (err) { toast(err.message, 'error'); }
  }

  async function handleDeleteSelected() {
    if (selected.size === 0) return;
    if (!confirm(`Delete ${selected.size} proxy(es)?`)) return;
    for (const id of selected) { try { await api.deletePool(id); } catch {} }
    setSelected(new Set()); load();
  }

  async function handleDeleteAll() {
    if (pools.length === 0) return;
    if (!confirm(`Delete ALL ${pools.length} proxies?`)) return;
    for (const p of pools) { try { await api.deletePool(p.id); } catch {} }
    load();
  }

  async function handleTest(id) {
    setTesting(id);
    try { await api.testPool(id); load(); } catch (err) { toast(err.message, 'error'); }
    setTesting(null);
  }

  async function handleTestAll() {
    if (pools.length === 0) return;
    setTestingAll(true);
    try {
      const r = await api.testAllPools();
      toast(`Alive: ${r.ok || 0}, Dead: ${r.fail || 0}`);
    } catch (err) { toast(err.message, 'error'); }
    setTestingAll(false); load();
  }

  async function handleBatchImport() {
    const lines = importText.split('\n').map(l => l.trim()).filter(Boolean);
    if (lines.length === 0) return;
    setImporting(true);
    let ok = 0, dup = 0;
    for (const line of lines) {
      try {
        let url = line;
        if (!line.includes('://')) {
          const p = line.split(':');
          if (p.length === 4) url = `http://${encodeURIComponent(p[2])}:${encodeURIComponent(p[3])}@${p[0]}:${p[1]}`;
          else url = `http://${line}`;
        }
        if (pools.some(x => x.proxy_url === url)) { dup++; continue; }
        await api.createPool({ name: url.replace(/http[s]?:\/\//, '').split('@').pop(), proxy_url: url });
        ok++;
      } catch { dup++; }
    }
    toast(`Imported: ${ok}, Skipped: ${dup}`);
    setShowImport(false); setImportText(''); setImporting(false); load();
  }

  async function handleFetchWebshare() {
    if (!settings.webshare_api_key) { toast('Set Webshare API key in Settings first', 'error'); return; }
    setFetching(true);
    try {
      const res = await fetch('/api/scraper/webshare', { method: 'POST' }).then(r => r.json());
      toast(`Imported: ${res.imported || 0}`);
      load();
    } catch (err) { toast(err.message, 'error'); }
    setFetching(false);
  }

  function togglePool(id) {
    const next = new Set(selected);
    if (next.has(id)) next.delete(id); else next.add(id);
    setSelected(next);
  }

  function toggleActive(pool) {
    api.createPool({ id: pool.id, is_active: !pool.is_active, proxy_url: pool.proxy_url, name: pool.name }).then(load).catch(() => {});
  }

  const filtered = search ? pools.filter(p => p.proxy_url?.includes(search) || p.name?.includes(search)) : pools;
  const allSelected = pools.length > 0 && selected.size === pools.length;

  return (
    <>
      <div className="flex items-center justify-between mb-4">
        <div>
          <h1 className="text-lg font-semibold text-text-main">Proxy Pools</h1>
          <p className="text-xs text-text-muted mt-0.5">
            {pools.length} proxies · {pools.filter(p => p.test_status === 'ok').length} alive
            · {pools.filter(p => !p.is_active).length} disabled
            {settings.enabled ? ' · Proxy ON' : ' · Proxy OFF'}
          </p>
        </div>
        <div className="flex items-center gap-2">
          <button onClick={() => setShowImport(true)}
            className="inline-flex items-center gap-1 px-3 py-1.5 rounded text-sm font-semibold bg-surface-2 text-text-main hover:bg-surface-3 transition-all">
            <MdAdd className="text-[16px]" /> Import
          </button>
          <button onClick={handleFetchWebshare} disabled={fetching}
            className="inline-flex items-center gap-1 px-3 py-1.5 rounded text-sm font-semibold bg-primary/10 text-primary hover:bg-primary/20 disabled:opacity-50 transition-all">
            <MdCloudDownload className="text-[16px]" /> {fetching ? '...' : 'Webshare'}
          </button>
          <button onClick={() => setShowSettings(true)}
            className="p-1.5 rounded text-text-muted hover:text-text-main hover:bg-surface-2 transition-all">
            <MdSettings className="text-lg" />
          </button>
        </div>
      </div>

      <div className="flex flex-wrap items-center gap-2 mb-3">
        <input value={search} onChange={e => setSearch(e.target.value)}
          placeholder="Search proxies..." className="h-8 px-2.5 rounded-lg bg-bg-alt border border-border-subtle text-xs text-text-main outline-none focus:border-primary/50 w-48" />
        <button onClick={handleTestAll} disabled={testingAll || pools.length === 0}
          className="inline-flex items-center gap-1 px-2.5 py-1.5 rounded text-xs font-medium bg-surface-2 text-text-main hover:bg-surface-3 disabled:opacity-50 transition-all">
          <MdScience className={`text-sm ${testingAll ? 'animate-spin' : ''}`} /> Test All
        </button>
        <button onClick={handleDeleteSelected} disabled={selected.size === 0}
          className="inline-flex items-center gap-1 px-2.5 py-1.5 rounded text-xs font-medium text-danger hover:bg-danger/10 disabled:opacity-30 transition-all">
          <MdClose className="text-sm" /> Delete Selected ({selected.size})
        </button>
        <button onClick={handleDeleteAll} disabled={pools.length === 0}
          className="inline-flex items-center gap-1 px-2.5 py-1.5 rounded text-xs font-medium text-danger hover:bg-danger/10 disabled:opacity-30 transition-all">
          <MdDeleteSweep className="text-sm" /> Delete All
        </button>
        <label className="flex items-center gap-1 text-xs text-text-muted ml-auto cursor-pointer">
          <input type="checkbox" checked={allSelected} onChange={() => setSelected(allSelected ? new Set() : new Set(pools.map(p => p.id)))} className="rounded border-border-subtle" />
          Select all
        </label>
      </div>

      <div className="overflow-x-auto rounded-lg border border-border-subtle">
        <table className="w-full text-xs">
          <thead>
            <tr className="bg-bg-alt text-text-muted font-semibold uppercase tracking-wider">
              <th className="px-2 py-2 w-8"></th>
              <th className="px-2 py-2 w-12">ON</th>
              <th className="px-2 py-2 w-16">Status</th>
              <th className="px-2 py-2 w-12">Type</th>
              <th className="px-2 py-2 w-20">Region</th>
              <th className="px-2 py-2">Host</th>
              <th className="px-2 py-2 w-12">Port</th>
              <th className="px-2 py-2 w-14 text-right">Latency</th>
              <th className="px-2 py-2">Last Checked</th>
              <th className="px-2 py-2 w-16 text-right">Actions</th>
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr><td colSpan={10} className="text-center py-8 text-text-subtle">Loading...</td></tr>
            ) : filtered.length === 0 ? (
              <tr><td colSpan={10} className="text-center py-8 text-text-subtle">No proxies.</td></tr>
            ) : filtered.map(pool => {
              const info = parseURL(pool.proxy_url);
              return (
                <tr key={pool.id} className="border-t border-border hover:bg-surface/40">
                  <td className="px-2 py-2">
                    <input type="checkbox" checked={selected.has(pool.id)} onChange={() => togglePool(pool.id)} className="rounded border-border-subtle" />
                  </td>
                  <td className="px-2 py-2">
                    <button onClick={() => toggleActive(pool)}
                      className={`relative inline-flex h-4 w-7 items-center rounded-full transition-colors ${pool.is_active ? 'bg-success' : 'bg-surface-3'}`}>
                      <span className={`inline-block h-3 w-3 rounded-full bg-white transition-transform ${pool.is_active ? 'translate-x-3.5' : 'translate-x-0.5'}`} />
                    </button>
                  </td>
                  <td className="px-2 py-2">
                    <span className={`inline-flex items-center gap-0.5 font-semibold ${pool.test_status === 'ok' ? 'text-success' : pool.test_status === 'error' ? 'text-danger' : 'text-text-muted'}`}>
                      {pool.test_status === 'ok' ? <MdCheck className="text-sm" /> : pool.test_status === 'error' ? <MdWarning className="text-sm" /> : null}
                      {pool.test_status || 'untested'}
                    </span>
                  </td>
                  <td className="px-2 py-2 font-mono text-text-muted">{info.type}</td>
                  <td className="px-2 py-2 text-text-muted">{pool.region || '-'}</td>
                  <td className="px-2 py-2 font-mono text-text-main" title={pool.proxy_url}>{info.host}</td>
                  <td className="px-2 py-2 font-mono text-text-muted">{info.port}</td>
                  <td className="px-2 py-2 text-right font-mono text-text-muted">{pool.latency_ms > 0 ? `${pool.latency_ms}ms` : '-'}</td>
                  <td className="px-2 py-2 text-text-muted">{fmtTime(pool.last_tested_at)}</td>
                  <td className="px-2 py-2 text-right">
                    <button onClick={() => handleTest(pool.id)} disabled={testing === pool.id}
                      className="p-1 rounded text-text-muted hover:text-text-main hover:bg-surface-2 transition-all" title="Test">
                      <MdScience className={`text-sm ${testing === pool.id ? 'animate-spin' : ''}`} />
                    </button>
                    <button onClick={() => handleDelete(pool.id)}
                      className="p-1 rounded text-text-muted hover:text-danger hover:bg-danger/10 transition-all" title="Delete">
                      <MdClose className="text-sm" />
                    </button>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>

      {showImport && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/30" onClick={() => setShowImport(false)}>
          <div className="bg-surface rounded-2xl shadow-elev p-6 w-full max-w-lg mx-4" onClick={e => e.stopPropagation()}>
            <h2 className="text-sm font-semibold text-text-main mb-1">Batch Import</h2>
            <p className="text-xs text-text-muted mb-3">One per line: <code className="font-mono bg-bg-alt px-1 rounded">http://user:pass@host:port</code></p>
            <textarea value={importText} onChange={e => setImportText(e.target.value)}
              placeholder="http://user:pass@1.2.3.4:7897"
              className="w-full h-28 px-3 py-2 rounded-lg bg-bg-alt border border-border-subtle text-xs text-text-main font-mono outline-none resize-none" />
            <div className="flex items-center gap-2 mt-3">
              <button onClick={handleBatchImport} disabled={importing || !importText.trim()}
                className="flex-1 py-1.5 rounded-lg text-xs font-semibold bg-primary text-white hover:bg-primary-hover disabled:opacity-50">
                {importing ? 'Importing...' : `Import ${importText.split('\n').filter(l => l.trim()).length || 0}`}
              </button>
              <button onClick={() => setShowImport(false)} className="px-3 py-1.5 rounded-lg text-xs font-medium text-text-muted hover:bg-surface-2">Cancel</button>
            </div>
          </div>
        </div>
      )}

      {showSettings && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/30" onClick={() => setShowSettings(false)}>
          <div className="bg-surface rounded-2xl shadow-elev p-6 w-full max-w-sm mx-4" onClick={e => e.stopPropagation()}>
            <h2 className="text-sm font-semibold text-text-main mb-4">Proxy Settings</h2>
            <div className="flex flex-col gap-3">
              <label className="flex items-center justify-between text-sm cursor-pointer">
                <span className="text-text-main font-medium">Proxy ON</span>
                <input type="checkbox" checked={settings.enabled} onChange={e => { const s={...settings,enabled:e.target.checked}; setSettings(s); api.updateProxy(s).catch(()=>{}); }} className="rounded" />
              </label>
              <div className="flex flex-col gap-1">
                <label className="text-xs text-text-main font-medium">Webshare API Key</label>
                <input type="password" value={settings.webshare_api_key || ''} onChange={e => setSettings(s => ({...s, webshare_api_key: e.target.value}))}
                  placeholder="ws_..." className="w-full px-2.5 py-1.5 rounded-lg bg-bg-alt border border-border-subtle text-xs text-text-main font-mono outline-none" />
              </div>
            </div>
            <div className="flex items-center gap-2 mt-4">
              <button onClick={() => { api.updateProxy(settings).then(() => { toast('Saved'); setShowSettings(false); }).catch(() => {}); }}
                className="flex-1 py-1.5 rounded-lg text-xs font-semibold bg-primary text-white hover:bg-primary-hover">Save</button>
              <button onClick={() => setShowSettings(false)} className="px-3 py-1.5 rounded-lg text-xs font-medium text-text-muted hover:bg-surface-2">Cancel</button>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
