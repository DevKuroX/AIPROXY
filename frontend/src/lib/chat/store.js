'use client';

import { create } from 'zustand';

const useChatStore = create((set, get) => ({
  // Existing state
  sessions: [],
  activeSessionId: null,
  messages: [],
  streaming: false,
  streamingContent: '',
  hasMoreMessages: true,
  sessionsLoading: false,
  messagesLoading: false,
  error: null,

  setSessions: (sessions) => set({ sessions }),
  setActiveSession: (id) => set({ activeSessionId: id, messages: [], hasMoreMessages: true }),
  setMessages: (messages) => set({ messages }),
  prependMessages: (olderMessages) =>
    set((s) => ({ messages: [...olderMessages, ...s.messages] })),
  addMessage: (msg) => set((s) => ({ messages: [...s.messages, msg] })),
  updateLastMessage: (updater) =>
    set((s) => {
      const msgs = [...s.messages];
      if (msgs.length > 0) msgs[msgs.length - 1] = updater(msgs[msgs.length - 1]);
      return { messages: msgs };
    }),

  setStreaming: (v) => set({ streaming: v }),
  setStreamingContent: (content) => set({ streamingContent: content }),
  appendStreamContent: (text) =>
    set((s) => ({ streamingContent: s.streamingContent + text })),
  clearStreamContent: () => set({ streamingContent: '' }),

  setHasMoreMessages: (v) => set({ hasMoreMessages: v }),
  setSessionsLoading: (v) => set({ sessionsLoading: v }),
  setMessagesLoading: (v) => set({ messagesLoading: v }),
  setError: (error) => set({ error }),
  clearError: () => set({ error: null }),

  artifactPanelOpen: false,
  activeArtifact: null,
  setArtifactPanelOpen: (v) => set({ artifactPanelOpen: v }),
  toggleArtifactPanel: () => set((s) => ({ artifactPanelOpen: !s.artifactPanelOpen })),
  setActiveArtifact: (artifact) => set({ activeArtifact: artifact, artifactPanelOpen: !!artifact }),

  // NEW: Thinking state
  thinkingState: null,          // null | 'thinking'
  thinkingStep: 0,
  thinkingMessages: ['Thinking...', 'Analyzing request...', 'Exploring repository...', 'Routing model...'],
  thinkingIntervalId: null,
  startThinking: () => {
    const s = get();
    if (s.thinkingIntervalId) clearInterval(s.thinkingIntervalId);
    set({ thinkingState: 'thinking', thinkingStep: 0 });
    const id = setInterval(() => {
      set((state) => ({ thinkingStep: (state.thinkingStep + 1) % state.thinkingMessages.length }));
    }, 2000);
    set({ thinkingIntervalId: id });
  },
  stopThinking: () => {
    const s = get();
    if (s.thinkingIntervalId) clearInterval(s.thinkingIntervalId);
    set({ thinkingState: null, thinkingStep: 0, thinkingIntervalId: null });
  },

  // NEW: Upload queue
  uploadQueue: [],
  uploadProgress: {},
  addToUploadQueue: (file) => set((s) => ({ uploadQueue: [...s.uploadQueue, file] })),
  removeFromUploadQueue: (index) => set((s) => ({ uploadQueue: s.uploadQueue.filter((_, i) => i !== index) })),
  setUploadProgress: (filename, percent) => set((s) => ({ uploadProgress: { ...s.uploadProgress, [filename]: percent } })),
  clearUploadQueue: () => set({ uploadQueue: [], uploadProgress: {} }),

  // NEW: Tool execution
  activeTools: [],
  toolHistory: [],
  addActiveTool: (tool) => set((s) => ({ activeTools: [...s.activeTools, tool] })),
  updateActiveTool: (id, updates) => set((s) => ({ activeTools: s.activeTools.map((t) => t.id === id ? { ...t, ...updates } : t) })),
  removeActiveTool: (id) => set((s) => ({ activeTools: s.activeTools.filter((t) => t.id !== id) })),
  completeActiveTool: (id) => {
    const s = get();
    const tool = s.activeTools.find((t) => t.id === id);
    if (tool) {
      set((state) => ({
        activeTools: state.activeTools.filter((t) => t.id !== id),
        toolHistory: [...state.toolHistory, { ...tool, completedAt: Date.now() }],
      }));
    }
  },
  clearToolHistory: () => set({ toolHistory: [] }),

  // NEW: Right panel tab
  rightPanelTab: 'artifacts',
  setRightPanelTab: (tab) => set({ rightPanelTab: tab }),

  // NEW: Model presets
  modelPresets: [
    { id: 'oc/deepseek-v4-flash-free', label: 'Fast Chat', icon: '\u26a1', desc: 'Quick responses' },
    { id: 'oc/claude-sonnet-4-5', label: 'Deep Reasoning', icon: '\ud83e\udde0', desc: 'Complex analysis', supportsVision: true },
    { id: 'oc/gpt-5.4', label: 'Coding Agent', icon: '\ud83d\udcbb', desc: 'Code generation', supportsVision: true },
  ],
  selectedPreset: null,
  selectPreset: (preset) => set({ selectedPreset: preset }),

  // NEW: Optimistic sidebar updates
  optimisticAddSession: (session) => set((s) => ({ sessions: [session, ...s.sessions] })),
  optimisticRemoveSession: (id) => set((s) => ({ sessions: s.sessions.filter((sess) => sess.id !== id) })),
  updateSession: (id, updates) => set((s) => ({ sessions: s.sessions.map((sess) => sess.id === id ? { ...sess, ...updates } : sess) })),
}));

export default useChatStore;
