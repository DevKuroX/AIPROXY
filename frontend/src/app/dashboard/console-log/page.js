'use client';

import { useState, useEffect, useRef } from 'react';
import { MdRefresh } from 'react-icons/md';

const LOG_COLORS = {
  LOG: 'text-green-400',
  INFO: 'text-blue-400',
  WARN: 'text-yellow-400',
  ERROR: 'text-red-400',
  DEBUG: 'text-purple-400',
};

function colorLine(line) {
  const m = line.match(/\[(\w+)\]/);
  const tag = m ? m[1] : null;
  const c = LOG_COLORS[tag] || 'text-green-400';
  return <span className={c}>{line}</span>;
}

export default function ConsoleLogPage() {
  const [logs, setLogs] = useState([]);
  const [connected, setConnected] = useState(false);
  const [error, setError] = useState(null);
  const [retryCount, setRetryCount] = useState(0);
  const logRef = useRef(null);
  const esRef = useRef(null);

  useEffect(() => {
    const connect = () => {
      if (esRef.current) {
        esRef.current.close();
      }

      const es = new EventSource('/api/translator/console-logs/stream');
      esRef.current = es;

      es.onopen = () => {
        setConnected(true);
        setError(null);
      };
      es.onmessage = (e) => {
        try {
          const msg = JSON.parse(e.data);
          if (msg.type === 'init') {
            setLogs(msg.logs ? msg.logs.slice(-200) : []);
          } else if (msg.type === 'line') {
            setLogs((prev) => {
              const next = [...prev, msg.line];
              return next.length > 200 ? next.slice(-200) : next;
            });
          } else if (msg.type === 'clear') {
            setLogs([]);
          }
        } catch {}
      };
      es.onerror = () => {
        setConnected(false);
        setError('SSE connection failed');
      };
    };

    connect();
    return () => {
      if (esRef.current) esRef.current.close();
    };
  }, [retryCount]);

  function handleRetry() {
    setError(null);
    setRetryCount(c => c + 1);
  }

  useEffect(() => {
    if (logRef.current) logRef.current.scrollTop = logRef.current.scrollHeight;
  }, [logs]);

  async function handleClear() {
    try {
      await fetch('/api/translator/console-logs', { method: 'DELETE' });
    } catch {}
  }

  return (
    <div className="flex flex-col gap-3">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-lg font-semibold text-text-main">Console Log</h1>
          <p className="text-xs text-text-muted mt-0.5">Live server logs via SSE</p>
        </div>
        <div className="flex items-center gap-2">
          {error && (
            <button onClick={handleRetry} className="inline-flex items-center gap-1 text-xs text-danger hover:text-danger/80 transition-colors">
              <MdRefresh /> Retry
            </button>
          )}
          <span className={`inline-flex items-center gap-1 text-xs ${connected ? 'text-success' : 'text-danger'}`}>
            <span className={`w-1.5 h-1.5 rounded-full ${connected ? 'bg-success' : 'bg-danger'}`} />
            {connected ? 'Connected' : 'Disconnected'}
          </span>
          <button onClick={handleClear}
            className="px-3 py-1.5 rounded text-xs font-semibold bg-surface-2 text-text-main hover:bg-surface-3 transition-all">
            Clear
          </button>
        </div>
      </div>

      <div ref={logRef} className="bg-[#0d0d0d] rounded-xl p-4 text-xs font-mono h-[calc(100vh-220px)] overflow-y-auto border border-border-subtle">
        {logs.length === 0 ? (
          <span className="text-text-muted">No console logs yet. Waiting for SSE connection...</span>
        ) : (
          <div className="space-y-0.5">
            {logs.map((line, i) => (
              <div key={i}>{colorLine(line)}</div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
