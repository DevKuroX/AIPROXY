'use client';

import { MdStop } from 'react-icons/md';

export default function ThinkingBar({ onStop }) {
  return (
    <div className="shrink-0 flex items-center justify-between px-4 lg:px-6 py-3 border-b border-border-subtle bg-surface/30">
      <div className="flex items-center gap-2.5">
        <div className="flex gap-1">
          <span className="size-2 rounded-full bg-primary animate-bounce" style={{ animationDelay: '0ms' }} />
          <span className="size-2 rounded-full bg-primary animate-bounce" style={{ animationDelay: '150ms' }} />
          <span className="size-2 rounded-full bg-primary animate-bounce" style={{ animationDelay: '300ms' }} />
        </div>
        <span className="text-sm text-text-muted">Thinking...</span>
      </div>
      {onStop && (
        <button
          onClick={onStop}
          className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-danger/10 text-danger text-xs font-medium hover:bg-danger/20 transition-colors"
        >
          <MdStop className="text-base" />
          Stop
        </button>
      )}
    </div>
  );
}
