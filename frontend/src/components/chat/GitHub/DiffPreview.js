'use client';

import { useState } from 'react';
import { MdCheck, MdClose, MdWarning } from 'react-icons/md';

export default function DiffPreview({ diff, onApprove, onReject }) {
  const [decision, setDecision] = useState(null);

  if (!diff) {
    return (
      <div className="flex items-center justify-center p-8">
        <p className="text-sm text-text-muted">No diff to preview</p>
      </div>
    );
  }

  const files = diff.files || [];
  const totalChanges = files.reduce((sum, f) => sum + (f.changes || 0), 0);
  const totalAdditions = files.reduce((sum, f) => sum + (f.additions || 0), 0);
  const totalDeletions = files.reduce((sum, f) => sum + (f.deletions || 0), 0);

  function handleApprove() {
    setDecision('approved');
    onApprove?.();
  }

  function handleReject() {
    setDecision('rejected');
    onReject?.();
  }

  return (
    <div className="flex flex-col h-full">
      <div className="px-4 py-3 border-b border-border-subtle">
        <div className="flex items-center gap-3 text-sm">
          <span className="text-text-muted">
            {files.length} file{files.length !== 1 ? 's' : ''} changed
          </span>
          <span className="text-success font-medium">+{totalAdditions}</span>
          <span className="text-danger font-medium">-{totalDeletions}</span>
          <span className="text-text-subtle">({totalChanges} changes)</span>
        </div>

        {files.length > 0 && (
          <div className="mt-2 flex flex-wrap gap-1">
            {files.map((f, i) => (
              <span
                key={i}
                className="text-xs bg-surface-2 text-text-muted px-2 py-0.5 rounded truncate max-w-48"
                title={f.filename}
              >
                {f.filename}
              </span>
            ))}
          </div>
        )}
      </div>

      <div className="flex-1 overflow-y-auto custom-scrollbar p-2">
        {files.map((file, fi) => (
          <div key={fi} className="mb-3 rounded-lg border border-border-subtle overflow-hidden">
            <div className="px-3 py-1.5 bg-surface-2 text-xs font-mono text-text-muted border-b border-border-subtle">
              {file.filename}
            </div>
            <pre className="text-xs font-mono leading-relaxed overflow-x-auto">
              {(file.patch || '').split('\n').map((line, li) => {
                let cls = '';
                if (line.startsWith('+')) cls = 'bg-success/10 text-success';
                else if (line.startsWith('-')) cls = 'bg-danger/10 text-danger';
                else if (line.startsWith('@@')) cls = 'bg-info/10 text-info';
                return (
                  <div key={li} className={'px-3 py-0.5 ' + cls}>
                    {line}
                  </div>
                );
              })}
            </pre>
          </div>
        ))}
      </div>

      {!decision && (
        <div className="px-4 py-3 border-t border-border-subtle flex items-center gap-3">
          <div className="flex items-center gap-1.5 text-xs text-text-muted">
            <MdWarning className="text-warning" />
            Review changes before applying
          </div>
          <div className="flex-1" />
          <button
            onClick={handleReject}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm font-medium text-danger hover:bg-danger/10 transition-colors"
          >
            <MdClose className="text-lg" />
            Reject
          </button>
          <button
            onClick={handleApprove}
            className="flex items-center gap-1.5 px-4 py-1.5 rounded-lg text-sm font-medium text-white bg-success hover:bg-success/90 transition-colors"
          >
            <MdCheck className="text-lg" />
            Approve
          </button>
        </div>
      )}

      {decision === 'approved' && (
        <div className="px-4 py-3 border-t border-border-subtle bg-success/10 text-success text-sm font-medium text-center">
          Changes approved
        </div>
      )}
      {decision === 'rejected' && (
        <div className="px-4 py-3 border-t border-border-subtle bg-danger/10 text-danger text-sm font-medium text-center">
          Changes rejected
        </div>
      )}
    </div>
  );
}
