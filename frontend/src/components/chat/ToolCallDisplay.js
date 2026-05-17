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

const typeLabels = {
  'github.search': 'Searching GitHub',
  'file.read': 'Reading file',
  'code.generate': 'Generating code',
  'web.search': 'Searching web',
};

export default function ToolCallDisplay({ tool, defaultExpanded }) {
  const [expanded, setExpanded] = useState(defaultExpanded || tool?.status === 'running');
  if (!tool) return null;

  const StatusIcon = statusIcons[tool.status] || MdCode;
  const statusColor = statusColors[tool.status] || 'text-text-muted';
  const label = typeLabels[tool.type] || tool.type || 'Tool';

  return (
    <div className="mx-4 lg:mx-6 my-2 rounded-xl border border-border-subtle bg-surface overflow-hidden">
      <button
        onClick={() => setExpanded(!expanded)}
        className="w-full flex items-center gap-3 px-4 py-2.5 hover:bg-surface-2 transition-colors text-left"
      >
        <StatusIcon className={`text-lg ${statusColor}`} />
        <div className="flex-1 min-w-0">
          <div className="text-sm font-medium text-text-main truncate">{label}</div>
          {tool.progress > 0 && tool.progress < 100 && (
            <div className="mt-1 h-1 bg-surface-3 rounded-full overflow-hidden">
              <div className="h-full bg-primary rounded-full transition-all duration-500" style={{ width: `${tool.progress}%` }} />
            </div>
          )}
        </div>
        <span className="text-xs text-text-muted capitalize shrink-0">{tool.status}</span>
        {expanded ? <MdExpandLess className="text-text-muted shrink-0" /> : <MdExpandMore className="text-text-muted shrink-0" />}
      </button>

      {expanded && tool.steps?.length > 0 && (
        <div className="border-t border-border-subtle px-4 py-2 space-y-1">
          {tool.steps.map((step, i) => (
            <div key={i} className="flex items-center gap-2 text-xs">
              <span className={step.status === 'completed' ? 'text-success' : step.status === 'running' ? 'text-info' : 'text-text-muted'}>
                {step.status === 'completed' ? '✓' : step.status === 'running' ? '⟳' : '○'}
              </span>
              <span className="text-text-main">{step.name}</span>
            </div>
          ))}
        </div>
      )}

      {expanded && tool.output && (
        <div className="border-t border-border-subtle px-4 py-2">
          <pre className="text-xs font-mono text-text-muted whitespace-pre-wrap max-h-40 overflow-y-auto">
            {typeof tool.output === 'string' ? tool.output : JSON.stringify(tool.output, null, 2)}
          </pre>
        </div>
      )}
    </div>
  );
}
