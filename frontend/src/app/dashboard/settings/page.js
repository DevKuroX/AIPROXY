'use client';

import { useState, useEffect } from 'react';
import { api } from '@/lib/api';
import { toast } from '@/components/Toast';

function Toggle({ label, checked, onChange }) {
  return (
    <label className="flex items-center justify-between py-2 cursor-pointer">
      <span className="text-sm text-text-muted">{label}</span>
      <button onClick={() => onChange(!checked)}
        className={`relative w-10 h-5 rounded-full transition-colors ${checked ? 'bg-primary' : 'bg-surface-3'}`}>
        <span className={`absolute top-0.5 left-0.5 w-4 h-4 bg-white rounded-full transition-transform ${checked ? 'translate-x-5' : ''}`} />
      </button>
    </label>
  );
}

function Section({ title, desc, children }) {
  return (
    <div className="card-soft p-6 mb-6">
      <h2 className="text-base font-semibold text-text-main mb-1">{title}</h2>
      <p className="text-xs text-text-muted mb-4">{desc}</p>
      {children}
    </div>
  );
}

export default function SettingsPage() {
  const [proxySettings, setProxySettings] = useState({});
  const [rtkEnabled, setRtkEnabled] = useState(true);
  const [cavemanEnabled, setCavemanEnabled] = useState(false);
  const [compactEnabled, setCompactEnabled] = useState(true);
  const [profile, setProfile] = useState({});

  useEffect(() => {
    api.proxySettings().then(d => setProxySettings(d?.settings || d || {})).catch(() => {});
    api.settings().then(d => {
      const s = d?.settings || d || {};
      setRtkEnabled(s.rtkEnabled ?? s.rtk_enabled ?? true);
      setCavemanEnabled(s.cavemanEnabled ?? s.caveman_enabled ?? false);
      setCompactEnabled(s.compactEnabled ?? s.compact_enabled ?? true);
    }).catch(() => {});
    api.me().then(d => setProfile(d?.user || d || {})).catch(() => {});
  }, []);

  async function handleSave() {
    try {
      await api.updateProxy({
        proxy_enabled: proxySettings.proxy_enabled ?? proxySettings.enabled ?? false,
        auto_scrape: proxySettings.auto_scrape ?? proxySettings.autoScrape ?? false,
        max_latency: parseInt(proxySettings.max_latency || proxySettings.maxLatency) || 5000,
      });
      toast('Settings saved');
    } catch (err) { toast(err.message, 'error'); }
  }

  return (
    <div className="max-w-2xl">
      <Section title="Proxy Settings" desc="Configure proxy pool behavior">
        <Toggle label="Enable Proxy" checked={proxySettings.proxy_enabled ?? proxySettings.enabled ?? false}
          onChange={v => setProxySettings(p => ({ ...p, proxy_enabled: v }))} />
        <Toggle label="Auto-scrape Proxies" checked={proxySettings.auto_scrape ?? proxySettings.autoScrape ?? false}
          onChange={v => setProxySettings(p => ({ ...p, auto_scrape: v }))} />
        <div className="mt-3">
          <label className="block text-xs text-text-muted mb-1">Max Latency (ms)</label>
          <input type="number" value={proxySettings.max_latency || proxySettings.maxLatency || 5000}
            onChange={e => setProxySettings(p => ({ ...p, max_latency: parseInt(e.target.value) || 5000 }))}
            className="w-32 px-3 py-2 rounded text-sm outline-none bg-bg-alt border border-border-subtle text-text-main focus:border-primary focus:shadow-focus transition-colors" />
        </div>
      </Section>

      <Section title="RTK, Caveman & Compact" desc="Token optimization settings">
        <Toggle label="RTK Compression" checked={rtkEnabled} onChange={setRtkEnabled} />
        <Toggle label="Caveman Mode" checked={cavemanEnabled} onChange={setCavemanEnabled} />
        <Toggle label="Compact Response" checked={compactEnabled} onChange={setCompactEnabled} />
      </Section>

      <Section title="Profile" desc="Account information">
        <div className="text-sm text-text-muted">
          <div><span className="text-text-subtle">User:</span> {profile.username || profile.name || 'admin'}</div>
          <div className="mt-1"><span className="text-text-subtle">Role:</span> {profile.role || 'admin'}</div>
        </div>
      </Section>

      <button onClick={handleSave}
        className="px-5 py-2.5 rounded text-sm font-semibold transition-all active:scale-[0.97] bg-primary text-white hover:bg-primary-hover">
        Save Settings
      </button>
    </div>
  );
}
