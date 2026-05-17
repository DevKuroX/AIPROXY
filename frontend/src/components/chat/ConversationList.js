'use client';

import { MdChat, MdDelete, MdCheck, MdSearch, MdPushPin } from 'react-icons/md';
import { useState, useMemo } from 'react';
import useChatStore from '@/lib/chat/store';
import { chatApi } from '@/lib/chat/api';

export default function ConversationList({ onSelectSidebar }) {
  const [confirmDelete, setConfirmDelete] = useState(null);
  const [searchQuery, setSearchQuery] = useState('');
  const sessions = useChatStore((s) => s.sessions);
  const activeSessionId = useChatStore((s) => s.activeSessionId);
  const sessionsLoading = useChatStore((s) => s.sessionsLoading);
  const setSessions = useChatStore((s) => s.setSessions);
  const setActiveSession = useChatStore((s) => s.setActiveSession);
  const setMessages = useChatStore((s) => s.setMessages);
  const setHasMoreMessages = useChatStore((s) => s.setHasMoreMessages);
  const setError = useChatStore((s) => s.setError);

  async function handleSelect(id) {
    setActiveSession(id);
    setMessages([]);
    setHasMoreMessages(true);
    onSelectSidebar?.();

    try {
      const msgs = await chatApi.listMessages(id);
      setMessages(msgs.reverse());
      if (msgs.length < 30) setHasMoreMessages(false);
    } catch (err) {
      setError(err.message);
    }
  }

  async function handleDelete(e, sessionId) {
    e.stopPropagation();
    if (confirmDelete === sessionId) {
      try {
        await chatApi.deleteSession(sessionId);
        const updated = await chatApi.listSessions();
        setSessions(updated);
        if (activeSessionId === sessionId) {
          setActiveSession(null);
          setMessages([]);
        }
      } catch (err) {
        setError(err.message);
      }
      setConfirmDelete(null);
    } else {
      setConfirmDelete(sessionId);
      setTimeout(() => setConfirmDelete(null), 3000);
    }
  }

  const filteredSessions = useMemo(() => {
    if (!searchQuery.trim()) return sessions;
    const q = searchQuery.toLowerCase();
    return sessions.filter((s) => (s.title || '').toLowerCase().includes(q));
  }, [sessions, searchQuery]);

  const groupedSessions = useMemo(() => {
    const groups = { today: [], yesterday: [], earlier: [] };
    const now = new Date();
    const today = new Date(now.getFullYear(), now.getMonth(), now.getDate());
    const yesterday = new Date(today);
    yesterday.setDate(yesterday.getDate() - 1);

    filteredSessions.forEach((session) => {
      const d = new Date(session.updated_at || session.created_at || Date.now());
      if (d >= today) {
        groups.today.push(session);
      } else if (d >= yesterday) {
        groups.yesterday.push(session);
      } else {
        groups.earlier.push(session);
      }
    });

    return groups;
  }, [filteredSessions]);

  const hasGroups =
    groupedSessions.today.length > 0 ||
    groupedSessions.yesterday.length > 0 ||
    groupedSessions.earlier.length > 0;

  function renderSession(session) {
    const isActive = session.id === activeSessionId;
    return (
      <button
        key={session.id}
        onClick={() => handleSelect(session.id)}
        className={`w-full flex items-center gap-3 px-3 py-2 rounded-xl text-left transition-all duration-150 group ${
          isActive
            ? 'bg-primary/10 text-primary'
            : 'text-text-muted hover:bg-surface-2 hover:text-text-main'
        }`}
      >
        <MdChat className="text-lg shrink-0" />
        <div className="flex-1 min-w-0">
          <div className="text-sm font-medium truncate">
            {session.title || 'New Chat'}
          </div>
          <div className="text-xs text-text-subtle mt-0.5">
            {formatTime(session.updated_at)}
          </div>
        </div>
        <MdPushPin className="text-lg shrink-0 text-text-subtle/30" />
        <button
          onClick={(e) => handleDelete(e, session.id)}
          className={`shrink-0 transition-all duration-150 ${
            confirmDelete === session.id
              ? 'text-danger'
              : 'opacity-0 group-hover:opacity-100 text-text-subtle hover:text-danger'
          }`}
          title={confirmDelete === session.id ? 'Click again to confirm' : 'Delete'}
        >
          {confirmDelete === session.id ? (
            <MdCheck className="text-lg" />
          ) : (
            <MdDelete className="text-lg" />
          )}
        </button>
      </button>
    );
  }

  function renderGroup(label, sessions) {
    if (sessions.length === 0) return null;
    return (
      <div key={label}>
        <div className="px-3 py-1.5 text-xs font-semibold text-text-subtle uppercase tracking-wider">
          {label}
        </div>
        {sessions.map(renderSession)}
      </div>
    );
  }

  return (
    <>
      <div className="px-3 pt-2 pb-1">
        <div className="relative">
          <MdSearch className="absolute left-2.5 top-1/2 -translate-y-1/2 text-text-subtle text-lg" />
          <input
            type="text"
            placeholder="Search conversations..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full bg-surface-2 text-text-main text-sm rounded-lg pl-8 pr-3 py-1.5 placeholder:text-text-subtle focus:outline-none focus:ring-1 focus:ring-primary/40 transition-all duration-150"
          />
        </div>
      </div>

      <div className="flex-1 overflow-y-auto custom-scrollbar p-2 space-y-0.5">
        {sessionsLoading ? (
          <p className="text-sm text-text-muted px-3 py-8 text-center">Loading...</p>
        ) : !hasGroups ? (
          <p className="text-sm text-text-muted px-3 py-8 text-center">
            {searchQuery ? 'No matching conversations' : 'No conversations yet'}
          </p>
        ) : (
          <>
            {renderGroup('Today', groupedSessions.today)}
            {renderGroup('Yesterday', groupedSessions.yesterday)}
            {renderGroup('Earlier', groupedSessions.earlier)}
          </>
        )}
      </div>
    </>
  );
}

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
