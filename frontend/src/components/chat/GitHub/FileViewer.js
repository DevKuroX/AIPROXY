'use client';

import { useState, useEffect, useRef } from 'react';
import { MdDownload, MdContentCopy } from 'react-icons/md';
import { chatApi } from '@/lib/chat/api';

function getLanguage(path) {
  const ext = path?.split('.').pop().toLowerCase();
  const map = {
    js: 'javascript', jsx: 'jsx', ts: 'typescript', tsx: 'tsx',
    py: 'python', rb: 'ruby', go: 'go', rs: 'rust',
    java: 'java', kt: 'kotlin', swift: 'swift',
    c: 'c', cpp: 'cpp', h: 'c', hpp: 'cpp',
    html: 'html', css: 'css', scss: 'scss', less: 'less',
    json: 'json', xml: 'xml', yaml: 'yaml', yml: 'yaml',
    md: 'markdown', sql: 'sql', sh: 'bash', bash: 'bash',
    toml: 'toml', dockerfile: 'dockerfile', conf: 'conf',
    vue: 'vue', svelte: 'svelte',
  };
  return map[ext] || '';
}

export default function FileViewer({ owner, repo, path, onClose }) {
  const [content, setContent] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    if (owner && repo && path) loadFile(owner, repo, path);
  }, [owner, repo, path]);

  async function loadFile(o, r, p) {
    setLoading(true);
    setError(null);
    try {
      const data = await chatApi.github.getFileContent(o, r, p);
      if (data.content) {
        const decoded = atob(data.content.replace(/\n/g, ''));
        setContent(decoded);
      } else if (data.html_url && !data.content) {
        setContent('// Binary file or too large to display');
      } else {
        setContent(JSON.stringify(data, null, 2));
      }
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  async function handleCopy() {
    if (content) {
      await navigator.clipboard.writeText(content);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    }
  }

  const lang = getLanguage(path);
  const filename = path?.split('/').pop();

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-between px-4 py-2 border-b border-border-subtle">
        <div className="flex items-center gap-2 min-w-0">
          <span className="text-xs font-mono text-text-muted truncate">{filename}</span>
          {lang && (
            <span className="text-xs bg-surface-2 text-text-muted px-1.5 py-0.5 rounded">{lang}</span>
          )}
        </div>
        <div className="flex items-center gap-1">
          <button onClick={handleCopy} className="text-text-muted hover:text-text-main p-1 transition-colors" title="Copy">
            {copied ? <span className="text-xs text-success">Copied!</span> : <MdContentCopy className="text-lg" />}
          </button>
        </div>
      </div>

      <div className="flex-1 overflow-y-auto custom-scrollbar">
        {loading ? (
          <p className="text-sm text-text-muted px-4 py-8 text-center">Loading...</p>
        ) : error ? (
          <p className="text-sm text-danger px-4 py-8 text-center">{error}</p>
        ) : (
          <pre className="text-xs font-mono text-text-main leading-relaxed p-4 overflow-x-auto">
            <code>{content}</code>
          </pre>
        )}
      </div>
    </div>
  );
}
