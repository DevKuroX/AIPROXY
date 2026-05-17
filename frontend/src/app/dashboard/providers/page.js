'use client';

import { useState, useEffect } from 'react';
import { api } from '@/lib/api';
import { MdSearch, MdClose, MdAdd, MdSmartToy, MdContentCopy, MdCheckCircle, MdLock, MdKey, MdRefresh } from 'react-icons/md';
import ProviderIcon from '@/components/ProviderIcon';

const AUTH_LABELS = {
  oauth:  'OAuth Providers',
  bearer: 'API Key Providers',
  apikey: 'API Key Providers',
  cookie: 'Cookie-Based Providers',
  none:   'Free / No-Auth Providers',
};

function groupKey(auth) { return { oauth:0, bearer:1, apikey:1, cookie:2, none:3 }[auth] ?? 99; }

export default function ProvidersPage() {
  const [providers, setProviders] = useState([]);
  const [accounts, setAccounts] = useState([]);
  const [models, setModels] = useState([]);
  const [search, setSearch] = useState('');
  const [loading, setLoading] = useState(true);
  const [selected, setSelected] = useState(null);
  const [showForm, setShowForm] = useState(false);
  const [newName, setNewName] = useState('');
  const [newKey, setNewKey] = useState('');
  const [saving, setSaving] = useState(false);
  const [copiedId, setCopiedId] = useState(null);
  const [error, setError] = useState(null);

  useEffect(() => {
    Promise.all([
      api._request('GET', '/api/providers').then(d => setProviders(d?.providers || [])).catch((err) => { setError(err.message || 'Failed to load providers'); return []; }),
      api.accounts().then(d => setAccounts(d?.accounts || d || [])).catch((err) => { setError(err.message || 'Failed to load accounts'); return []; }),
      api.models().then(d => setModels(d?.data || d || [])).catch(() => []),
    ]).finally(() => setLoading(false));
  }, []);

  function retry() {
    setError(null);
    setLoading(true);
    Promise.all([
      api._request('GET', '/api/providers').then(d => setProviders(d?.providers || [])).catch((err) => { setError(err.message || 'Failed to load providers'); return []; }),
      api.accounts().then(d => setAccounts(d?.accounts || d || [])).catch((err) => { setError(err.message || 'Failed to load accounts'); return []; }),
      api.models().then(d => setModels(d?.data || d || [])).catch(() => []),
    ]).finally(() => setLoading(false));
  }

  const grouped = {};
  for (const p of providers) {
    const k = p.auth_type || 'none';
    if (!grouped[k]) grouped[k] = [];
    if (!search || p.id.toLowerCase().includes(search.toLowerCase()) || (p.name||'').toLowerCase().includes(search.toLowerCase()))
      grouped[k].push(p);
  }
  const sorted = Object.entries(grouped).filter(([,v])=>v.length).sort(([a],[b])=>groupKey(a)-groupKey(b));

  const sel = selected ? providers.find(p=>p.id===selected) : null;
  const selAccs = selected ? accounts.filter(a=>(a.provider_name||a.provider_id)===selected) : [];
  const selModels = selected ? models.filter(m=>(m.id||'').startsWith(selected+'/')) : [];
  const isBearer = sel?.auth_type==='bearer'||sel?.auth_type==='apikey';

  function copyId(id) { navigator.clipboard.writeText(id); setCopiedId(id); setTimeout(()=>setCopiedId(null),1500); }

  async function handleSave() {
    if (!selected||(isBearer&&!newKey.trim())) return;
    setSaving(true);
    try {
      const p = { provider_id: selected, name: newName.trim()||selected+'-key', is_active: true };
      if (isBearer) p.api_key = newKey.trim();
      await api.createAccount(p);
      setShowForm(false); setNewName(''); setNewKey('');
      const d = await api.accounts(); setAccounts(d?.accounts||d||[]);
    } catch(e) { console.error(e); } finally { setSaving(false); }
  }

  return (
    <div className="fade-in">
      <div className="flex items-center justify-between mb-6">
        <div className="relative">
          <MdSearch className="absolute left-3 top-1/2 -translate-y-1/2 text-[18px] text-text-subtle pointer-events-none" />
          <input type="text" value={search} onChange={e=>setSearch(e.target.value)}
            className="w-64 pl-9 pr-3 py-2 rounded text-sm outline-none bg-bg-alt border border-border-subtle text-text-main focus:border-primary focus:shadow-focus transition-colors"
            placeholder="Search providers..." />
        </div>
        <span className="text-sm text-text-muted">{providers.length} providers</span>
      </div>

      {loading ? (
        <div className="text-center py-16 text-sm text-text-subtle">Loading...</div>
      ) : error ? (
        <div className="text-center py-16">
          <p className="text-sm text-danger mb-4">{error}</p>
          <button onClick={retry} className="inline-flex items-center gap-1.5 px-4 py-2 rounded-lg bg-danger/10 text-danger text-sm font-medium hover:bg-danger/20 transition-colors">
            <MdRefresh /> Retry
          </button>
        </div>
      ) : sorted.length===0 ? (
        <div className="text-center py-16 text-sm text-text-subtle">No providers found</div>
      ) : (
        <div className="space-y-8">{sorted.map(([auth,ps])=>(
          <div key={auth}>
            <div className="flex items-center gap-2 mb-4">
              <h2 className="text-xs font-semibold text-text-muted/60 uppercase tracking-wider">{AUTH_LABELS[auth]||auth}</h2>
              <span className="text-xs text-text-muted">{ps.length}</span>
            </div>
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-3">
              {ps.map(p=>{const aa=accounts.filter(a=>(a.provider_name||a.provider_id)===p.id);const act=aa.filter(a=>a.is_active!==false).length;
              return (
                <button key={p.id} onClick={()=>{setSelected(p.id);setShowForm(false)}}
                  className="card-soft p-4 text-left transition-all hover:shadow-warm hover:border-primary/30 cursor-pointer">
                  <div className="flex items-center gap-3 mb-1.5">
                    <ProviderIcon providerId={p.id} name={p.name||p.id} size={36} />
                    <div className="min-w-0">
                      <div className="text-sm font-semibold text-text-main truncate">{p.name||p.id}</div>
                      <div className="text-[11px] text-text-muted truncate">{p.id}</div>
                    </div>
                  </div>
                  <div className="flex items-center gap-2 mt-2">
                    {act>0&&<span className="text-[11px] px-1.5 py-0.5 rounded-full bg-success/10 text-success font-medium">{act} active</span>}
                    {aa.length===0&&<span className="text-[11px] text-text-subtle">No connections</span>}
                  </div>
                </button>
              );})}
            </div>
          </div>
        ))}</div>
      )}

      {sel&&(
        <div className="fixed inset-0 z-50 flex items-start justify-center p-4 pt-12 overflow-y-auto"
          style={{background:'rgba(0,0,0,0.6)',backdropFilter:'blur(4px)'}}
          onClick={()=>{setSelected(null);setShowForm(false)}}>
          <div className="card-elev w-full max-w-2xl p-6" onClick={e=>e.stopPropagation()}>

            <div className="flex items-center justify-between mb-6">
              <div className="flex items-center gap-3">
                <ProviderIcon providerId={sel.id} name={sel.name||sel.id} size={44} />
                <div>
                  <h2 className="text-lg font-semibold text-text-main">{sel.name||sel.id}</h2>
                  <div className="flex items-center gap-2 mt-0.5">
                    <span className="text-xs font-mono text-text-muted">{sel.id}</span>
                    <span className={`text-[10px] px-1.5 py-0.5 rounded-full font-medium ${
                      sel.auth_type==='oauth'?'bg-blue-500/10 text-blue-400':
                      isBearer?'bg-green-500/10 text-green-400':
                      sel.auth_type==='none'?'bg-purple-500/10 text-purple-400':'bg-gray-500/10 text-gray-400'
                    }`}>{sel.auth_type}</span>
                  </div>
                </div>
              </div>
              <button onClick={()=>{setSelected(null);setShowForm(false)}} className="text-text-muted hover:text-text-main transition-colors">
                <MdClose className="text-xl" />
              </button>
            </div>

            <div className="mb-6">
              <div className="flex items-center justify-between mb-3">
                <h3 className="text-sm font-semibold text-text-main">Connections ({selAccs.length})</h3>
                {isBearer&&!showForm&&(
                  <button onClick={()=>setShowForm(true)}
                    className="inline-flex items-center gap-1 text-xs font-semibold text-primary hover:text-primary-hover transition-colors">
                    <MdAdd /> Add Connection
                  </button>
                )}
              </div>

              {selAccs.length>0&&(
                <div className="space-y-1.5 mb-3">
                  {selAccs.map(a=>(
                    <div key={a.id} className="flex items-center justify-between p-3 rounded-lg bg-bg-alt border border-border-subtle">
                      <div className="flex items-center gap-2 min-w-0">
                        <span className={`w-2 h-2 rounded-full shrink-0 ${a.is_active!==false?'bg-success':'bg-text-subtle'}`} />
                        <span className="text-sm text-text-main truncate">{a.name||a.id}</span>
                      </div>
                      {a.credit!=null&&<span className="text-xs text-text-muted shrink-0">{a.credit}%</span>}
                    </div>
                  ))}
                </div>
              )}

              {showForm&&isBearer&&(
                <div className="p-4 rounded-lg bg-bg-alt border border-border-subtle">
                  <h4 className="text-sm font-medium text-text-main mb-3">Add API Key</h4>
                  <div className="flex flex-col gap-3">
                    <input type="text" value={newName} onChange={e=>setNewName(e.target.value)}
                      className="w-full px-3 py-2 rounded text-sm outline-none bg-surface border border-border-subtle text-text-main focus:border-primary focus:shadow-focus transition-colors"
                      placeholder="Connection name" />
                    <input type="text" value={newKey} onChange={e=>setNewKey(e.target.value)}
                      className="w-full px-3 py-2 rounded text-sm outline-none bg-surface border border-border-subtle text-text-main focus:border-primary focus:shadow-focus transition-colors font-mono"
                      placeholder="API Key / Access Token" />
                    <div className="flex justify-end gap-2">
                      <button onClick={()=>setShowForm(false)} className="px-3 py-1.5 rounded text-sm font-semibold bg-surface-2 text-text-main hover:bg-surface-3 transition-all active:scale-[0.97]">Cancel</button>
                      <button onClick={handleSave} disabled={saving||!newKey.trim()}
                        className="px-3 py-1.5 rounded text-sm font-semibold bg-primary text-white hover:bg-primary-hover transition-all active:scale-[0.97] disabled:opacity-50">
                        {saving?'Saving...':'Save'}
                      </button>
                    </div>
                  </div>
                </div>
              )}

              {selAccs.length===0&&!showForm&&(
                <div className="flex items-center gap-3 p-4 rounded-lg bg-bg-alt border border-border-subtle">
                  <div className="inline-flex items-center justify-center w-9 h-9 rounded-full bg-primary/10 text-primary shrink-0">
                    {sel.auth_type==='oauth'?<MdLock className="text-lg" />:<MdKey className="text-lg" />}
                  </div>
                  <p className="text-sm text-text-muted">
                    {sel.auth_type==='oauth'?'No connections yet.':
                     isBearer?'No connections yet.':
                     sel.auth_type==='none'?'Auto-configured.':'No connections yet.'}
                  </p>
                  {isBearer&&(
                    <button onClick={()=>setShowForm(true)}
                      className="ml-auto inline-flex items-center gap-1 px-3 py-1.5 rounded text-xs font-semibold bg-primary text-white hover:bg-primary-hover shrink-0">
                      <MdAdd /> Add
                    </button>
                  )}
                </div>
              )}
            </div>

            <div>
              <div className="flex items-center justify-between mb-3">
                <h3 className="text-sm font-semibold text-text-main">Available Models ({selModels.length})</h3>
              </div>
              {selModels.length===0?(
                <div className="text-sm text-text-muted py-4 text-center bg-bg-alt rounded-lg border border-border-subtle">No models listed.</div>
              ):(
                <div className="flex flex-wrap gap-2">
                  {selModels.map(m=>{const id=m.id||'';
                  return (
                    <div key={id} className="group flex items-center gap-1.5 rounded-lg border border-border px-3 py-1.5 hover:bg-bg-alt transition-colors text-sm">
                      <MdSmartToy className="text-text-muted text-sm shrink-0" />
                      <code className="text-xs font-mono text-text-main">{id}</code>
                      <button onClick={()=>copyId(id)} className="text-text-muted hover:text-primary transition-colors shrink-0">
                        {copiedId===id?<MdCheckCircle className="text-success text-sm" />:<MdContentCopy className="text-sm opacity-0 group-hover:opacity-100 transition-opacity" />}
                      </button>
                    </div>
                  );})}
                </div>
              )}
            </div>

            <div className="text-xs text-text-muted space-y-1 mt-4 pt-4 border-t border-border-subtle">
              <div><span className="text-text-subtle">Type:</span> {sel.type}</div>
              <div><span className="text-text-subtle">Auth:</span> {sel.auth_type}</div>
              {sel.base_url&&<div className="truncate"><span className="text-text-subtle">Base URL:</span> {sel.base_url}</div>}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
