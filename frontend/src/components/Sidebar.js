'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { logout } from '@/lib/auth';
import { MdDashboard, MdDns, MdPeople, MdBarChart, MdKey, MdSettings, MdLogout, MdHub, MdChat, MdShield, MdCloudDownload, MdTerminal } from 'react-icons/md';

const mainItems = [
  { href: '/dashboard',            icon: MdDashboard, label: 'Dashboard' },
  { href: '/dashboard/providers',  icon: MdDns,       label: 'Providers' },
  { href: '/dashboard/accounts',   icon: MdPeople,    label: 'Accounts' },

  { href: '/dashboard/usage',      icon: MdBarChart,  label: 'Usage' },
  { href: '/dashboard/keys',       icon: MdKey,       label: 'API Keys' },
  { href: '/dashboard/proxy-pools',  icon: MdShield,  label: 'Proxy Pools' },
  { href: '/chat',                    icon: MdChat,         label: 'AI Chat' },
  { href: '/dashboard/console-log',    icon: MdTerminal,     label: 'Console Log' },
];

function isActive(href, pathname) {
  if (href === '/dashboard') return pathname === '/dashboard' || pathname === '/';
  return pathname.startsWith(href);
}

function NavLink({ item, pathname, onClose }) {
  const active = isActive(item.href, pathname);
  const Icon = item.icon;
  return (
    <Link
      href={item.href}
      onClick={onClose}
      className={`flex items-center gap-3 px-3 py-1 rounded-lg transition-all group ${
        active ? 'bg-primary/10 text-primary' : 'text-text-muted hover:bg-surface-2 hover:text-text-main'
      }`}
    >
      <Icon className="text-[18px]" />
      <span className="text-[13px] font-medium">{item.label}</span>
    </Link>
  );
}

export default function Sidebar({ onClose }) {
  const pathname = usePathname();

  return (
    <aside className="flex w-72 flex-col border-r border-border-subtle bg-vibrancy backdrop-blur-xl transition-colors duration-300 min-h-full">
      <div className="flex items-center gap-2 px-6 pt-5 pb-2">
        <div className="w-3 h-3 rounded-full bg-[#FF5F56]" />
        <div className="w-3 h-3 rounded-full bg-[#FFBD2E]" />
        <div className="w-3 h-3 rounded-full bg-[#27C93F]" />
      </div>

      <div className="px-6 py-4 flex flex-col gap-2">
        <Link href="/dashboard" className="flex items-center gap-3">
          <div className="flex items-center justify-center size-9 rounded-[10px] bg-gradient-to-br from-brand-500 to-brand-700 shadow-warm">
            <MdHub className="text-white text-[20px]" />
          </div>
          <div className="flex flex-col">
            <h1 className="text-lg font-semibold tracking-tight text-text-main">AIPROXY</h1>
            <span className="text-xs text-text-muted">Dashboard</span>
          </div>
        </Link>
      </div>

      <nav className="flex-1 px-4 py-2 space-y-0.5 overflow-y-auto custom-scrollbar">
        {mainItems.map(item => (
          <NavLink key={item.href} item={item} pathname={pathname} onClose={onClose} />
        ))}

        <div className="pt-3 mt-2 space-y-0.5">
          <p className="px-4 text-xs font-semibold text-text-muted/60 uppercase tracking-wider mb-2">System</p>
          <NavLink
            item={{ href: '/dashboard/settings', icon: MdSettings, label: 'Settings' }}
            pathname={pathname}
            onClose={onClose}
          />
          <button
            onClick={logout}
            className="w-full flex items-center gap-3 px-3 py-1 rounded-lg transition-all group text-text-muted hover:bg-surface-2 hover:text-text-main"
          >
            <MdLogout className="text-[18px] group-hover:text-primary transition-colors" />
            <span className="text-[13px] font-medium">Logout</span>
          </button>
        </div>
      </nav>
    </aside>
  );
}
