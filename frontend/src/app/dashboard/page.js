'use client';

import { useState, useEffect } from 'react';
import { api } from '@/lib/api';
import Link from 'next/link';
import { MdDns, MdPeople, MdBarChart } from 'react-icons/md';

function StatCard({ label, value, color }) {
  return (
    <div className="card-soft p-5">
      <div className="text-sm mb-1 text-text-muted">{label}</div>
      <div className={`text-2xl font-bold ${color || 'text-text-main'}`}>{value ?? '—'}</div>
    </div>
  );
}

const quickLinks = [
  { href: '/dashboard/providers', icon: MdDns, label: 'Providers', desc: 'Browse all providers' },
  { href: '/dashboard/accounts', icon: MdPeople, label: 'Accounts', desc: 'Manage account pool' },
  { href: '/dashboard/usage', icon: MdBarChart, label: 'Usage', desc: 'View analytics' },
];

export default function DashboardPage() {
  const [stats, setStats] = useState({});

  useEffect(() => {
    (async () => {
      try {
        const [health, models, accounts] = await Promise.all([
          api.health().catch(() => null),
          api.models().catch(() => null),
          api.accounts().catch(() => null),
        ]);
        const modelList = models?.data || models || [];
        const uniqueProviders = new Set(modelList.map(m => (m.id || '').split('/')[0]).filter(Boolean));
        const accList = accounts?.accounts || accounts || [];
        setStats({
          health: health?.status || 'OK',
          providers: uniqueProviders.size,
          accounts: accList.length,
          models: modelList.length,
        });
      } catch {}
    })();
  }, []);

  return (
    <>
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
        <StatCard label="Health" value={stats.health} color="text-success" />
        <StatCard label="Providers" value={stats.providers} />
        <StatCard label="Accounts" value={stats.accounts} />
        <StatCard label="Models" value={stats.models} />
      </div>

      <div className="text-xs font-semibold text-text-muted/60 uppercase tracking-wider mb-3">Quick Access</div>
      <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
        {quickLinks.map(item => {
          const Icon = item.icon;
          return (
            <Link key={item.href} href={item.href}
              className="card-soft p-4 transition-all hover:shadow-warm hover:border-primary/30 cursor-pointer block">
              <Icon className="text-accent text-xl" />
              <div className="text-sm font-medium mt-2 text-text-main">{item.label}</div>
              <div className="text-xs mt-0.5 text-text-subtle">{item.desc}</div>
            </Link>
          );
        })}
      </div>
    </>
  );
}
