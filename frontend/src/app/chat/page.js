'use client';

import { useEffect, useRef, useMemo } from 'react';
import { MdAutoAwesome, MdChat } from 'react-icons/md';
import { useVirtualizer } from '@tanstack/react-virtual';
import useChatStore from '@/lib/chat/store';
import { chatApi } from '@/lib/chat/api';
import ChatMessage from '@/components/chat/ChatMessage';
import ChatInput from '@/components/chat/ChatInput';
import WelcomeScreen from '@/components/chat/WelcomeScreen';
import useChatStream from '@/components/chat/ChatStream';
import { toast } from '@/components/Toast';

export default function ChatPage() {
  const sessions = useChatStore((s) => s.sessions);
  const activeSessionId = useChatStore((s) => s.activeSessionId);
  const messages = useChatStore((s) => s.messages);
  const streaming = useChatStore((s) => s.streaming);
  const streamingContent = useChatStore((s) => s.streamingContent);
  const hasMoreMessages = useChatStore((s) => s.hasMoreMessages);
  const modelPresets = useChatStore((s) => s.modelPresets);
  const selectedPreset = useChatStore((s) => s.selectedPreset);
  const error = useChatStore((s) => s.error);
  
  const setSessions = useChatStore((s) => s.setSessions);
  const setActiveSession = useChatStore((s) => s.setActiveSession);
  const setMessages = useChatStore((s) => s.setMessages);
  const prependMessages = useChatStore((s) => s.prependMessages);
  const addMessage = useChatStore((s) => s.addMessage);
  const setStreaming = useChatStore((s) => s.setStreaming);
  const setStreamingContent = useChatStore((s) => s.setStreamingContent);
  const appendStreamContent = useChatStore((s) => s.appendStreamContent);
  const clearStreamContent = useChatStore((s) => s.clearStreamContent);
  const setHasMoreMessages = useChatStore((s) => s.setHasMoreMessages);
  const setSessionsLoading = useChatStore((s) => s.setSessionsLoading);
  const setError = useChatStore((s) => s.setError);
  const clearError = useChatStore((s) => s.clearError);
  const selectPreset = useChatStore((s) => s.selectPreset);
  const optimisticAddSession = useChatStore((s) => s.optimisticAddSession);
  const updateSession = useChatStore((s) => s.updateSession);

  const messagesEndRef = useRef(null);
  const scrollContainerRef = useRef(null);
  const isLoadingMore = useRef(false);
  const pendingSessionRef = useRef(null);
  const isNearBottom = useRef(true);

  const { startStream, stopStream } = useChatStream();

  useEffect(() => {
    if (error) {
      toast(error, 'error');
      clearError();
    }
  }, [error, clearError]);

  useEffect(() => {
    loadSessions();

    const hash = window.location.hash.replace(/^#/, '');
    if (hash) {
      pendingSessionRef.current = hash;
    }

    const onHashChange = () => {
      const h = window.location.hash.replace(/^#/, '');
      if (h && h !== activeSessionId) {
        selectSession(h);
      } else if (!h) {
        setActiveSession(null);
        setMessages([]);
      }
    };
    window.addEventListener('hashchange', onHashChange);
    return () => window.removeEventListener('hashchange', onHashChange);
  }, []);

  useEffect(() => {
    if (pendingSessionRef.current && sessions.length > 0) {
      const id = pendingSessionRef.current;
      const exists = sessions.find((s) => s.id === id);
      if (exists) {
        selectSession(id);
      }
      pendingSessionRef.current = null;
    }
  }, [sessions]);

  useEffect(() => {
    if (activeSessionId) {
      window.history.pushState(null, '', `/chat#${activeSessionId}`);
    } else {
      window.history.pushState(null, '', '/chat');
    }
  }, [activeSessionId]);

  useEffect(() => {
    if (messagesEndRef.current && !isLoadingMore.current) {
      messagesEndRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [messages.length, streamingContent]);

  async function loadSessions() {
    setSessionsLoading(true);
    try {
      const list = await chatApi.listSessions();
      setSessions(list);
    } catch (err) {
      console.error('Failed to load sessions:', err);
    } finally {
      setSessionsLoading(false);
    }
  }

  async function selectSession(id) {
    setActiveSession(id);
    setMessages([]);
    setHasMoreMessages(true);
    try {
      const msgs = await chatApi.listMessages(id);
      setMessages(msgs.reverse());
      if (msgs.length < 30) setHasMoreMessages(false);
    } catch (err) {
      setError(err.message);
    }
  }

  async function handleSelectPreset(preset) {
    selectPreset(preset);
    try {
      const session = await chatApi.createSession();
      optimisticAddSession(session);
      setActiveSession(session.id);
      setMessages([]);
      setHasMoreMessages(true);
    } catch (err) {
      setError(err.message);
    }
  }

  function fileToBase64(file) {
    return new Promise((resolve, reject) => {
      const reader = new FileReader();
      reader.onload = () => resolve(reader.result);
      reader.onerror = reject;
      reader.readAsDataURL(file);
    });
  }

  async function buildMultimodalContent(text, files) {
    if (!files || files.length === 0) return text;
    const parts = [];
    if (text) parts.push({ type: 'text', text });
    for (const file of files) {
      const isImage = file.type?.startsWith('image/') || /\.(jpg|jpeg|png|gif|webp|svg)$/i.test(file.name);
      if (isImage) {
        try {
          const b64 = await fileToBase64(file);
          parts.push({ type: 'image_url', image_url: { url: b64 } });
        } catch {}
      }
    }
    return parts.length > 1 || (parts.length === 1 && parts[0].type === 'image_url') ? parts : text;
  }

  async function handleSend(text) {
    if (!activeSessionId || !text.trim() || streaming) return;

    const model = selectedPreset?.id || 'oc/deepseek-v4-flash-free';
    const files = useChatStore.getState().uploadQueue;
    const clearUploadQueue = useChatStore.getState().clearUploadQueue;

    const hasFiles = files.length > 0;
    const multimodalContent = hasFiles ? await buildMultimodalContent(text, files) : text;
    let fileRefs = [];

    if (hasFiles) {
      for (const file of files) {
        try {
          const uploaded = await chatApi.uploadFile(activeSessionId, file);
          fileRefs.push(uploaded);
        } catch (err) {
          setError('Upload failed: ' + err.message);
        }
      }
      clearUploadQueue();
    }

    const storeContent = Array.isArray(multimodalContent) ? text : multimodalContent;
    const userMsg = { role: 'user', content: storeContent };
    if (fileRefs.length > 0) {
      userMsg.file_preview = fileRefs[0];
    }
    addMessage(userMsg);
    clearStreamContent();
    setStreaming(true);

    try {
      await chatApi.saveMessage(activeSessionId, { role: 'user', content: storeContent });
    } catch (err) {
      setError(err.message);
      setStreaming(false);
      return;
    }

    const allMessages = [...messages, { role: 'user', content: multimodalContent }];

    startStream(model, allMessages, {
      onChunk: (chunk) => {
        appendStreamContent(chunk);
      },
      onDone: async () => {
        const finalContent = useChatStore.getState().streamingContent;
        addMessage({ role: 'assistant', content: finalContent });
        clearStreamContent();
        setStreaming(false);

        if (finalContent) {
          try {
            await chatApi.saveMessage(activeSessionId, { role: 'assistant', content: finalContent });
            chatApi.generateTitle(activeSessionId).then((res) => {
              if (res?.title) updateSession(activeSessionId, { title: res.title });
            }).catch(() => {});
          } catch (err) {
            setError(err.message);
          }
        }
      },
      onError: (err) => {
        setError(err.message);
        setStreaming(false);
        clearStreamContent();
      },
    });
  }

  const handleScroll = async (e) => {
    const el = e.target;
    isNearBottom.current = el.scrollHeight - el.scrollTop - el.clientHeight < 100;

    if (el.scrollTop < 100 && hasMoreMessages && !isLoadingMore.current && activeSessionId) {
      isLoadingMore.current = true;
      const beforeHeight = el.scrollHeight;

      try {
        const older = await chatApi.listMessages(activeSessionId, { before: messages[0]?.id, limit: 30 });
        if (older.length > 0) {
          prependMessages(older.reverse());
        }
        if (older.length < 30) setHasMoreMessages(false);

        requestAnimationFrame(() => {
          const afterHeight = el.scrollHeight;
          el.scrollTop = afterHeight - beforeHeight;
        });
      } catch (err) {
        setError(err.message);
      } finally {
        isLoadingMore.current = false;
      }
    }
  };

  const virtualItems = useMemo(() => {
    const items = [...messages];
    if (streaming && streamingContent) {
      items.push({ id: 'streaming', role: 'assistant', content: streamingContent });
    }
    return items;
  }, [messages, streaming, streamingContent]);

  const virtualizer = useVirtualizer({
    count: virtualItems.length,
    getScrollElement: () => scrollContainerRef.current,
    estimateSize: () => 100,
    overscan: 5,
  });

  if (!activeSessionId) {
    return (
      <WelcomeScreen
        onSelectPreset={handleSelectPreset}
        onSelectSession={selectSession}
      />
    );
  }

  return (
    <div className="flex flex-col flex-1 min-h-0">
      <div className="shrink-0 flex items-center justify-between px-4 lg:px-6 py-2 border-b border-border-subtle bg-surface/40">
        <div className="flex items-center gap-2">
          <span className="text-sm text-text-muted">Model:</span>
          <span className="text-sm font-medium text-text-main">
            {selectedPreset?.label || selectedPreset?.id || 'Fast Chat'}
          </span>
        </div>

      </div>

      <div
        ref={scrollContainerRef}
        onScroll={handleScroll}
        className="flex-1 overflow-y-auto custom-scrollbar"
      >
        <div style={{ height: virtualizer.getTotalSize(), position: 'relative', width: '100%' }}>
          <div
            style={{
              position: 'absolute',
              top: 0,
              left: 0,
              width: '100%',
              transform: `translateY(${virtualizer.getVirtualItems()[0]?.start ?? 0}px)`,
            }}
          >
            <div className="max-w-4xl mx-auto">
              {virtualizer.getVirtualItems().map((virtualRow) => {
                const item = virtualItems[virtualRow.index];
                return (
                  <div
                    key={item.id || virtualRow.index}
                    ref={virtualRow.measureElement}
                    data-index={virtualRow.index}
                  >
                    <ChatMessage
                      message={item}
                      isStreaming={streaming && virtualRow.index === virtualItems.length - 1}
                    />
                  </div>
                );
              })}
              <div ref={messagesEndRef} />
            </div>
          </div>
        </div>
      </div>

      <ChatInput onSend={handleSend} disabled={streaming} streaming={streaming} streamingContent={streamingContent} onStop={stopStream} />
    </div>
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
