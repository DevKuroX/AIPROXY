'use client';

import { useEffect, useState } from 'react';
import { MdClose, MdDescription, MdImage, MdCode, MdArticle, MdTerminal, MdAccountTree } from 'react-icons/md';
import useChatStore from '@/lib/chat/store';
import { chatApi } from '@/lib/chat/api';

const typeIcons = {
  diff: MdCode,
  file: MdDescription,
  image: MdImage,
  markdown: MdArticle,
  log: MdTerminal,
  'repo-tree': MdAccountTree,
};

const typeLabels = {
  diff: 'Diff',
  file: 'File',
  image: 'Image',
  markdown: 'Markdown',
  log: 'Log',
  'repo-tree': 'Repository Tree',
};

function CodeBlock({ content, path }) {
  const [copied, setCopied] = useState(false);
  const handleCopy = () => {
    navigator.clipboard.writeText(content).catch(() => {});
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };
  return (
    <div className="p-4">
      {path && (
        <div className="text-xs text-text-muted mb-2 font-mono">{path}</div>
      )}
      <div className="relative group">
        <pre className="text-xs font-mono text-text-main whitespace-pre-wrap leading-relaxed bg-surface-2 rounded-lg p-3 overflow-x-auto">
          {content}
        </pre>
        <button
          onClick={handleCopy}
          className="absolute top-2 right-2 text-xs px-2 py-1 bg-surface-1 hover:bg-surface-3 text-text-muted rounded transition-colors opacity-0 group-hover:opacity-100"
        >
          {copied ? 'Copied!' : 'Copy'}
        </button>
      </div>
    </div>
  );
}

function renderPayload(type, payload) {
  if (!payload) return <p className="text-sm text-text-muted px-4 py-8 text-center">No content</p>;

  switch (type) {
    case 'image':
      return (
        <div className="p-4">
          <img src={payload.url || payload} alt="artifact" className="max-w-full rounded-lg" />
        </div>
      );
    case 'markdown':
      return (
        <div className="p-4 text-sm text-text-main leading-relaxed whitespace-pre-wrap font-mono">
          {typeof payload === 'string' ? payload : JSON.stringify(payload, null, 2)}
        </div>
      );
    case 'diff':
      const diffLines = (payload.diff || payload.content || '').split('\n');
      const adds = diffLines.filter(l => l.startsWith('+')).length;
      const rems = diffLines.filter(l => l.startsWith('-')).length;
      return (
        <div className="p-2">
          <div className="text-xs text-text-muted px-2 py-1 mb-1 border-b border-border-subtle">
            <span className="text-success">+{adds}</span> <span className="text-danger">-{rems}</span> total {diffLines.length} lines
          </div>
          <pre className="text-xs leading-relaxed font-mono overflow-x-auto">
            {diffLines.map((line, i) => {
              let cls = '';
              if (line.startsWith('+')) cls = 'text-success';
              else if (line.startsWith('-')) cls = 'text-danger';
              else if (line.startsWith('@@')) cls = 'text-info';
              return (
                <div key={i} className={cls + ' px-2 py-0.5 hover:bg-surface-2'}>
                  {line}
                </div>
              );
            })}
          </pre>
        </div>
      );
    case 'log':
      const logContent = typeof payload === 'string' ? payload : JSON.stringify(payload, null, 2);
      return (
        <div className="p-4">
          <pre className="text-xs font-mono text-text-main whitespace-pre-wrap leading-relaxed">
            {logContent.split('\n').map((line, i) => (
              <span key={i} className="block">
                <span className="text-text-muted select-none mr-3 inline-block w-8 text-right">{i + 1}</span>
                {line || ' '}
              </span>
            ))}
          </pre>
        </div>
      );
    case 'file':
      return <CodeBlock content={payload.content || ''} path={payload.path} />;
    case 'repo-tree':
      const getDepth = (p) => (p && typeof p === 'string' ? p.split('/').length - 1 : 0);
      return (
        <div className="p-2">
          {Array.isArray(payload) ? (
            payload.map((item, i) => {
              const label = item.path || item.name || item;
              const depth = getDepth(item.path || item.name);
              return (
                <div
                  key={i}
                  className="flex items-center gap-2 py-1.5 text-sm hover:bg-surface-2 rounded-lg"
                  style={{ paddingLeft: `${12 + depth * 16}px` }}
                >
                  <MdAccountTree className="text-text-muted text-base shrink-0" />
                  <span className="text-text-main truncate">{label}</span>
                </div>
              );
            })
          ) : (
            <pre className="text-xs font-mono text-text-main whitespace-pre-wrap">
              {JSON.stringify(payload, null, 2)}
            </pre>
          )}
        </div>
      );
    default:
      return (
        <pre className="text-xs font-mono text-text-main whitespace-pre-wrap p-4">
          {JSON.stringify(payload, null, 2)}
        </pre>
      );
  }
}

export default function ArtifactPanel() {
  const activeArtifact = useChatStore((s) => s.activeArtifact);
  const artifactPanelOpen = useChatStore((s) => s.artifactPanelOpen);
  const setActiveArtifact = useChatStore((s) => s.setActiveArtifact);
  const setArtifactPanelOpen = useChatStore((s) => s.setArtifactPanelOpen);

  if (!activeArtifact) {
    return (
      <div className="flex flex-col h-full">
        <div className="flex items-center justify-between px-4 h-14 border-b border-border-subtle">
          <h2 className="text-xs font-semibold text-text-muted uppercase tracking-wider">Artifacts</h2>
          <button
            onClick={() => setArtifactPanelOpen(false)}
            className="text-text-muted hover:text-text-main transition-colors"
          >
            <MdClose className="text-lg" />
          </button>
        </div>
        <div className="flex-1 flex items-center justify-center">
          <div className="text-center px-6">
            <MdDescription className="text-3xl text-text-muted mx-auto mb-3 opacity-40" />
            <p className="text-sm text-text-muted">
              Select an artifact to view
            </p>
          </div>
        </div>
      </div>
    );
  }

  const Icon = typeIcons[activeArtifact.type] || MdDescription;
  const label = typeLabels[activeArtifact.type] || activeArtifact.type;

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-between px-4 h-14 border-b border-border-subtle">
        <div className="flex items-center gap-2 min-w-0">
          <Icon className="text-primary text-lg shrink-0" />
          <h2 className="text-sm font-semibold text-text-main truncate">{label}</h2>
        </div>
        <button
          onClick={() => setActiveArtifact(null)}
          className="text-text-muted hover:text-text-main transition-colors shrink-0"
        >
          <MdClose className="text-lg" />
        </button>
      </div>
      <div className="flex-1 overflow-y-auto custom-scrollbar">
        {renderPayload(activeArtifact.type, activeArtifact.payload)}
      </div>
    </div>
  );
}
