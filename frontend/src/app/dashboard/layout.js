'use client';

import { useState } from 'react';
import Sidebar from '@/components/Sidebar';
import Header from '@/components/Header';
import ToastProvider from '@/components/Toast';

export default function DashboardLayout({ children }) {
  const [sidebarOpen, setSidebarOpen] = useState(false);

  return (
    <ToastProvider>
      <div className="flex h-screen w-full overflow-hidden bg-bg transition-colors duration-300">
        {sidebarOpen && (
          <div className="fixed inset-0 z-40 bg-black/20 lg:hidden" onClick={() => setSidebarOpen(false)} />
        )}

        <div className="hidden lg:flex">
          <Sidebar />
        </div>

        <div className={`fixed inset-y-0 left-0 z-50 transform lg:hidden transition-transform duration-300 ease-in-out ${sidebarOpen ? 'translate-x-0' : '-translate-x-full'}`}>
          <Sidebar onClose={() => setSidebarOpen(false)} />
        </div>

        <main className="flex flex-col flex-1 h-full min-w-0 relative transition-colors duration-300 isolate">
          <div className="landing-grid absolute inset-0 pointer-events-none -z-10" aria-hidden="true" />
          <Header onMenuClick={() => setSidebarOpen(true)} />
          <div className="flex-1 overflow-y-auto custom-scrollbar p-6 lg:p-10">
            <div className="max-w-7xl mx-auto">{children}</div>
          </div>
        </main>
      </div>
    </ToastProvider>
  );
}
