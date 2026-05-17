'use client';

import { MdClose, MdDescription, MdCode, MdTerminal, MdFolder } from 'react-icons/md';
import useChatStore from '@/lib/chat/store';
import ArtifactPanel from './ArtifactPanel';

const tabs = [
  { id: 'artifacts', label: 'Artifacts', icon: MdCode },
  { id: 'tools', label: 'Tools', icon: MdTerminal },
  { id: 'files', label: 'Files', icon: MdFolder },
  { id: 'github', label: 'GitHub', icon: MdDescription },
];

export default function RightPanel() {
  const rightPanelTab = useChatStore((s) => s.rightPanelTab);
  const setRightPanelTab = useChatStore((s) => s.setRightPanelTab);
  const setArtifactPanelOpen = useChatStore((s) => s.setArtifactPanelOpen);
  const toolHistory = useChatStore((s) => s.toolHistory);
  const uploadQueue = useChatStore((s) => s.uploadQueue);

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-between px-3 h-12 border-b border-border-subtle shrink-0">
        <div className="flex gap-0.5">
          {tabs.map((tab) => {
            const Icon = tab.icon;
            const isActive = rightPanelTab === tab.id;
            return (
              <button
                key={tab.id}
                onClick={() => setRightPanelTab(tab.id)}
                className={`flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg text-xs font-medium transition-colors ${
                  isActive ? 'bg-primary/10 text-primary' : 'text-text-muted hover:text-text-main hover:bg-surface-2'
                }`}
              >
                <Icon className="text-sm" />
                {tab.label}
              </button>
            );
          })}
        </div>
        <button
          onClick={() => setArtifactPanelOpen(false)}
          className="text-text-muted hover:text-text-main transition-colors"
        >
          <MdClose className="text-lg" />
        </button>
      </div>

      <div className="flex-1 overflow-y-auto custom-scrollbar">
        {rightPanelTab === 'artifacts' && <ArtifactPanel />}

        {rightPanelTab === 'tools' && (
          <div className="p-3 space-y-2">
            <p className="text-xs font-semibold text-text-muted/60 uppercase tracking-wider mb-2">Tool History</p>
            {toolHistory.length === 0 ? (
              <p className="text-sm text-text-muted px-2 py-4 text-center">No tool executions yet</p>
            ) : (
              toolHistory.map((tool, i) => (
                <div key={tool.id || i} className="p-3 rounded-xl bg-surface border border-border-subtle">
                  <div className="flex items-center gap-2 text-sm">
                    <MdTerminal className="text-primary" />
                    <span className="font-medium text-text-main">{tool.type || 'Tool'}</span>
                    <span className="text-xs text-success ml-auto">Completed</span>
                  </div>
                  {tool.output && (
                    <pre className="mt-2 text-xs font-mono text-text-muted whitespace-pre-wrap max-h-32 overflow-y-auto">
                      {typeof tool.output === 'string' ? tool.output : JSON.stringify(tool.output)}
                    </pre>
                  )}
                </div>
              ))
            )}
          </div>
        )}

        {rightPanelTab === 'files' && (
          <div className="p-3">
            <p className="text-xs font-semibold text-text-muted/60 uppercase tracking-wider mb-2">Uploaded Files</p>
            {uploadQueue.length === 0 ? (
              <p className="text-sm text-text-muted px-2 py-4 text-center">No files uploaded</p>
            ) : (
              uploadQueue.map((file, i) => (
                <div key={i} className="flex items-center gap-2 p-2 rounded-lg hover:bg-surface-2 transition-colors">
                  <MdFolder className="text-info text-lg" />
                  <div className="flex-1 min-w-0">
                    <div className="text-sm text-text-main truncate">{file.name}</div>
                    <div className="text-xs text-text-muted">{(file.size / 1024).toFixed(1)} KB</div>
                  </div>
                </div>
              ))
            )}
          </div>
        )}

        {rightPanelTab === 'github' && (
          <div className="p-3">
            <p className="text-xs font-semibold text-text-muted/60 uppercase tracking-wider mb-2">GitHub</p>
            <p className="text-sm text-text-muted px-2 py-4 text-center">GitHub integration not connected</p>
          </div>
        )}
      </div>
    </div>
  );
}
