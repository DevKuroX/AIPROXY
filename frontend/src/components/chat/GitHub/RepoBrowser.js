'use client';

import { useState, useEffect } from 'react';
import { MdFolder, MdInsertDriveFile, MdChevronRight, MdRefresh } from 'react-icons/md';
import { chatApi } from '@/lib/chat/api';

function sortTreeItems(items) {
  if (!items) return [];
  return [...items].sort((a, b) => {
    if (a.type !== b.type) return a.type === 'tree' ? -1 : 1;
    return (a.path || a.name || '').localeCompare(b.path || b.name || '');
  });
}

export default function RepoBrowser({ onSelectFile, repo }) {
  const [items, setItems] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [currentPath, setCurrentPath] = useState('');

  useEffect(() => {
    if (repo) loadTree(repo);
  }, [repo]);

  async function loadTree(r) {
    setLoading(true);
    setError(null);
    try {
      const data = await chatApi.github.getRepoTree(r.owner, r.repo);
      setItems(sortTreeItems(data.tree || []));
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  function navigateToDir(path) {
    setCurrentPath(path);
  }

  function isInCurrentPath(item) {
    if (!currentPath) return !item.path || !item.path.includes('/');
    return item.path && item.path.startsWith(currentPath + '/') && item.path.split('/').length === currentPath.split('/').length + 1;
  }

  const visibleItems = currentPath
    ? items.filter(isInCurrentPath)
    : items.filter((item) => !item.path || !item.path.includes('/'));

  function handleClick(item) {
    if (item.type === 'tree') {
      navigateToDir(item.path || item.name);
    } else {
      onSelectFile?.(repo, item.path);
    }
  }

  function goBack() {
    const parts = currentPath.split('/');
    parts.pop();
    setCurrentPath(parts.join('/'));
  }

  const pathParts = currentPath ? currentPath.split('/') : [];

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-between px-4 py-2 border-b border-border-subtle">
        <div className="flex items-center gap-1 text-xs text-text-muted min-w-0">
          <span className="font-medium text-text-main">{repo?.owner}/{repo?.repo}</span>
          {pathParts.map((part, i) => (
            <span key={i} className="flex items-center gap-1">
              <MdChevronRight />
              <span>{part}</span>
            </span>
          ))}
        </div>
        <button onClick={() => loadTree(repo)} className="text-text-muted hover:text-text-main transition-colors">
          <MdRefresh className="text-lg" />
        </button>
      </div>

      <div className="flex-1 overflow-y-auto custom-scrollbar p-1">
        {loading ? (
          <p className="text-sm text-text-muted px-3 py-8 text-center">Loading...</p>
        ) : error ? (
          <p className="text-sm text-danger px-3 py-8 text-center">{error}</p>
        ) : (
          <>
            {currentPath && (
              <button
                onClick={goBack}
                className="w-full flex items-center gap-2 px-3 py-2 text-sm text-text-muted hover:text-text-main hover:bg-surface-2 rounded-lg"
              >
                <MdChevronRight className="rotate-180 text-lg" />
                <span>..</span>
              </button>
            )}
            {visibleItems.map((item, i) => (
              <button
                key={item.path || item.name || i}
                onClick={() => handleClick(item)}
                className="w-full flex items-center gap-3 px-3 py-1.5 text-sm hover:bg-surface-2 rounded-lg transition-colors text-left"
              >
                {item.type === 'tree' ? (
                  <MdFolder className="text-info text-lg shrink-0" />
                ) : (
                  <MdInsertDriveFile className="text-text-muted text-lg shrink-0" />
                )}
                <span className="truncate text-text-main">{item.path?.split('/').pop() || item.name}</span>
              </button>
            ))}
            {visibleItems.length === 0 && (
              <p className="text-sm text-text-muted px-3 py-8 text-center">Empty directory</p>
            )}
          </>
        )}
      </div>
    </div>
  );
}
