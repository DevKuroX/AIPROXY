'use client';

import { MdChat } from 'react-icons/md';
import useChatStore from '@/lib/chat/store';

function formatTime(iso) {
  if (!iso) return '';
  const d = new Date(iso);
  const now = new Date();
  const diffMs = now - d;
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMs / 3600000);
  const diffDays = Math.floor(diffMs / 86400000);
  if (diffMins < 1) return 'Just now';
  if (diffMins < 60) return `${diffMins}m ago`;
  if (diffHours < 24) return `${diffHours}h ago`;
  if (diffDays < 7) return `${diffDays}d ago`;
  return d.toLocaleDateString();
}

export default function WelcomeScreen({ onSelectPreset, onSelectSession }) {
  const sessions = useChatStore((s) => s.sessions);
  const modelPresets = useChatStore((s) => s.modelPresets);

  return (
    <div className="flex-1 flex flex-col items-center justify-center px-6 py-12 overflow-y-auto">
      <div className="flex items-center justify-center size-16 rounded-2xl bg-gradient-to-br from-brand-500 to-brand-700 shadow-warm mb-5">
        <MdChat className="text-white text-3xl" />
      </div>
      <h1 className="text-2xl font-semibold text-text-main mb-1">AI Workspace</h1>
      <p className="text-sm text-text-muted mb-8 max-w-md text-center">
        Your connected development environment
      </p>

      <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 max-w-2xl w-full mb-8">
        {modelPresets.map((preset) => (
          <button
            key={preset.id}
            onClick={() => onSelectPreset(preset)}
            className="flex flex-col items-start gap-2 p-4 rounded-xl bg-surface border border-border-subtle hover:border-primary/50 hover:shadow-warm transition-all text-left group"
          >
            <span className="text-2xl">{preset.icon}</span>
            <div>
              <div className="text-sm font-medium text-text-main group-hover:text-primary transition-colors">{preset.label}</div>
              <div className="text-xs text-text-muted mt-0.5">{preset.desc}</div>
            </div>
          </button>
        ))}
      </div>

      {sessions.length > 0 && (
        <div className="max-w-2xl w-full">
          <p className="text-xs font-semibold text-text-muted/60 uppercase tracking-wider mb-2 px-1">Recent Conversations</p>
          <div className="space-y-1">
            {sessions.slice(0, 5).map((s) => (
              <button
                key={s.id}
                onClick={() => onSelectSession(s.id)}
                className="w-full flex items-center gap-3 px-3 py-2.5 rounded-xl hover:bg-surface-2 transition-all text-left"
              >
                <MdChat className="text-lg text-text-muted shrink-0" />
                <span className="text-sm text-text-main truncate">{s.title || 'New Chat'}</span>
                <span className="text-xs text-text-muted shrink-0 ml-auto">{formatTime(s.updated_at)}</span>
              </button>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
