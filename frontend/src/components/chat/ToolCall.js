'use client';

import { useState } from 'react';
import { MdCode, MdCheckCircle, MdError, MdHourglassEmpty, MdExpandMore, MdExpandLess } from 'react-icons/md';

const statusIcons = {
  running: MdHourglassEmpty,
  completed: MdCheckCircle,
  error: MdError,
};

const statusColors = {
  running: 'text-info',
  completed: 'text-success',
  error: 'text-danger',
};

export default function ToolCall({ tool, defaultExpanded }) {
  const [expanded, setExpanded] = useState(defaultExpanded || false);
  const StatusIcon = statusIcons[tool.status] || MdCode;
  const statusColor = statusColors[tool.status] || 'text-text-muted';

  return (
    <div className="mx-4 lg:mx-8 my-2 rounded-xl border border-border-subtle bg-surface overflow-hidden">
      <button
        onClick={() => setExpanded(!expanded)}
        className="w-full flex items-center gap-3 px-4 py-2.5 hover:bg-surface-2 transition-colors text-left"
      >
        <StatusIcon className={`text-lg ${statusColor}`} />
        <span className="text-sm font-medium text-text-main flex-1">{tool.name}</span>
        <span className="text-xs text-text-muted capitalize">{tool.status}</span>
        {expanded ? <MdExpandLess className="text-text-muted" /> : <MdExpandMore className="text-text-muted" />}
      </button>

      {expanded && (
        <div className="border-t border-border-subtle">
          {tool.input && (
            <div className="px-4 py-2.5 border-b border-border-subtle">
              <div className="text-xs font-medium text-text-muted mb-1">Input</div>
              <pre className="text-xs font-mono text-text-main whitespace-pre-wrap leading-relaxed max-h-40 overflow-y-auto">
                {typeof tool.input === 'string' ? tool.input : JSON.stringify(tool.input, null, 2)}
              </pre>
            </div>
          )}
          {tool.output && (
            <div className="px-4 py-2.5">
              <div className="text-xs font-medium text-text-muted mb-1">Output</div>
              <pre className="text-xs font-mono text-text-main whitespace-pre-wrap leading-relaxed max-h-60 overflow-y-auto">
                {typeof tool.output === 'string' ? tool.output : JSON.stringify(tool.output, null, 2)}
              </pre>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
