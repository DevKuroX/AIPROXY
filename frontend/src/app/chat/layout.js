'use client';

import { useState, useCallback } from 'react';
import Link from 'next/link';
import { 
  MdDarkMode, MdLightMode, MdMenu, MdCode, 
  MdAdd, MdChat, MdSettings, MdLogout 
} from 'react-icons/md';
import { useTheme } from '@/components/ThemeProvider';
import ConversationList from '@/components/chat/ConversationList';
import RightPanel from '@/components/chat/RightPanel';
import useChatStore from '@/lib/chat/store';
import { chatApi } from '@/lib/chat/api';
import { logout } from '@/lib/auth';

export default function ChatLayout({ children }) {
  const { theme, toggle } = useTheme();
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const artifactPanelOpen = useChatStore((s) => s.artifactPanelOpen);
  const toggleArtifactPanel = useChatStore((s) => s.toggleArtifactPanel);
  const error = useChatStore((s) => s.error);
  const clearError = useChatStore((s) => s.clearError);
  const setSessions = useChatStore((s) => s.setSessions);
  const setActiveSession = useChatStore((s) => s.setActiveSession);
  const setMessages = useChatStore((s) => s.setMessages);
  const setHasMoreMessages = useChatStore((s) => s.setHasMoreMessages);
  const setError = useChatStore((s) => s.setError);

  const handleCreateSession = useCallback(async () => {
    try {
      const session = await chatApi.createSession();
      const updated = await chatApi.listSessions();
      setSessions(updated);
      setActiveSession(session.id);
      setMessages([]);
      setHasMoreMessages(true);
      setSidebarOpen(false);
    } catch (err) {
      setError(err.message);
    }
  }, [setSessions, setActiveSession, setMessages, setHasMoreMessages, setError]);

  return (
    <div className="flex h-screen w-full overflow-hidden bg-bg transition-colors duration-300">
      {sidebarOpen && (
        <div className="fixed inset-0 z-40 bg-black/20 lg:hidden" onClick={() => setSidebarOpen(false)} />
      )}

      <aside className={[
        'fixed inset-y-0 left-0 z-50 w-64 flex flex-col border-r border-border-subtle',
        'bg-vibrancy backdrop-blur-xl',
        'lg:static lg:translate-x-0 lg:z-auto',
        'transition-transform duration-300 ease-in-out',
        sidebarOpen ? 'translate-x-0' : '-translate-x-full',
      ].join(' ')}>
        <div className="px-5 pt-5 pb-4 shrink-0">
          <Link href="/chat" className="flex items-center gap-3">
            <div className="flex items-center justify-center size-10 rounded-xl bg-gradient-to-br from-brand-500 to-brand-700 shadow-warm">
              <MdChat className="text-white text-xl" />
            </div>
            <div className="flex flex-col">
              <h1 className="text-lg font-semibold tracking-tight text-text-main">AI Workspace</h1>
              <span className="text-xs text-text-muted">Connected development</span>
            </div>
          </Link>
        </div>

        <div className="px-4 pb-4 shrink-0">
          <button
            onClick={handleCreateSession}
            className="w-full flex items-center justify-center gap-2 px-4 py-2.5 rounded-xl bg-primary text-white font-medium text-sm hover:bg-primary-hover transition-all active:scale-[0.98] shadow-warm"
          >
            <MdAdd className="text-lg" />
            New Chat
          </button>
        </div>

        <div className="flex-1 overflow-y-auto custom-scrollbar">
          <div className="px-2">
            <p className="px-3 text-xs font-semibold text-text-muted/60 uppercase tracking-wider mb-1">Conversations</p>
            <ConversationList onSelectSidebar={() => setSidebarOpen(false)} />
          </div>

          <div className="px-4 pt-4">
            <p className="px-1 text-xs font-semibold text-text-muted/60 uppercase tracking-wider mb-2">Agents</p>
            <div className="px-3 py-3 rounded-xl text-text-muted text-sm bg-surface-2/50">
              Coming soon
            </div>
          </div>

          <div className="px-4 pt-4">
            <p className="px-1 text-xs font-semibold text-text-muted/60 uppercase tracking-wider mb-2">Files</p>
            <div className="px-3 py-3 rounded-xl text-text-muted text-sm bg-surface-2/50">
              Coming soon
            </div>
          </div>

          <div className="px-4 pt-4 pb-4">
            <p className="px-1 text-xs font-semibold text-text-muted/60 uppercase tracking-wider mb-2">GitHub</p>
            <div className="px-3 py-3 rounded-xl text-text-muted text-sm bg-surface-2/50">
              Coming soon
            </div>
          </div>
        </div>

        <div className="shrink-0 px-4 py-3 border-t border-border-subtle space-y-0.5">
          <Link
            href="/dashboard/settings"
            onClick={() => setSidebarOpen(false)}
            className="flex items-center gap-3 px-3 py-2 rounded-lg transition-all text-text-muted hover:bg-surface-2 hover:text-text-main"
          >
            <MdSettings className="text-lg" />
            <span className="text-sm font-medium">Settings</span>
          </Link>
          <button
            onClick={logout}
            className="w-full flex items-center gap-3 px-3 py-2 rounded-lg transition-all text-text-muted hover:bg-surface-2 hover:text-text-main"
          >
            <MdLogout className="text-lg" />
            <span className="text-sm font-medium">Logout</span>
          </button>
        </div>
      </aside>

      <div className="flex flex-col flex-1 min-w-0">
        {error && (
          <div className="px-4 py-2 bg-danger/10 text-danger text-sm text-center border-b border-danger/20 shrink-0 flex items-center justify-center gap-2">
            <span>{error}</span>
            <button onClick={clearError} className="underline whitespace-nowrap">Dismiss</button>
          </div>
        )}

        <header className="shrink-0 flex items-center justify-between px-4 lg:px-6 h-14 border-b border-border-subtle bg-surface/60 backdrop-blur-xl z-20 transition-colors duration-300">
          <div className="flex items-center gap-3">
            <button
              onClick={() => setSidebarOpen(true)}
              className="lg:hidden text-text-muted hover:text-text-main transition-colors"
              aria-label="Open sidebar"
            >
              <MdMenu className="text-xl" />
            </button>
            <h1 className="text-base font-semibold text-text-main">AI Workspace</h1>
          </div>

          <div className="flex items-center gap-1">
            <button
              onClick={toggleArtifactPanel}
              className={`flex items-center justify-center size-9 rounded-full transition-colors ${
                artifactPanelOpen
                  ? 'text-primary bg-primary/10'
                  : 'text-text-muted hover:text-text-main hover:bg-surface-2'
              }`}
              aria-label="Toggle artifacts"
            >
              <MdCode className="text-[20px]" />
            </button>
            <button
              onClick={toggle}
              className="flex items-center justify-center size-9 rounded-full text-text-muted hover:text-text-main hover:bg-surface-2 transition-colors"
              aria-label={theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
            >
              {theme === 'dark' ? <MdLightMode className="text-[20px]" /> : <MdDarkMode className="text-[20px]" />}
            </button>
          </div>
        </header>

        <div className="flex flex-1 min-h-0">
          <main className="flex-1 flex flex-col min-w-0">
            {children}
          </main>

          {artifactPanelOpen && (
            <>
              <div 
                className="fixed inset-0 z-30 bg-black/20 lg:hidden" 
                onClick={toggleArtifactPanel} 
              />
              <aside className="fixed inset-y-0 right-0 z-40 w-80 border-l border-border-subtle bg-surface/95 backdrop-blur-xl lg:static lg:z-auto">
                <RightPanel />
              </aside>
            </>
          )}
        </div>
      </div>
    </div>
  );
}
