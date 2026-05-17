'use client';

import { useRef, useCallback } from 'react';
import { chatApi } from '@/lib/chat/api';

export default function useChatStream() {
  const abortRef = useRef(null);
  const cleanupRef = useRef(null);
  const bufferRef = useRef('');
  const rafRef = useRef(null);
  const onContentCallback = useRef(null);
  const onDoneCallback = useRef(null);
  const onErrorCallback = useRef(null);

  const flushBuffer = useCallback(() => {
    if (bufferRef.current && onContentCallback.current) {
      onContentCallback.current(bufferRef.current);
      bufferRef.current = '';
    }
  }, []);

  const rafLoop = useCallback(() => {
    flushBuffer();
    rafRef.current = requestAnimationFrame(rafLoop);
  }, [flushBuffer]);

  const startStream = useCallback((model, messages, { onChunk, onDone, onError }) => {
    onContentCallback.current = onChunk;
    onDoneCallback.current = onDone;
    onErrorCallback.current = onError;

    const abortController = new AbortController();
    abortRef.current = abortController;

    rafRef.current = requestAnimationFrame(rafLoop);

    const cancel = chatApi.streamChat(model, messages, {
      signal: abortController.signal,
      onChunk: (text) => {
        bufferRef.current += text;
      },
      onDone: () => {
        flushBuffer();
        if (rafRef.current) cancelAnimationFrame(rafRef.current);
        rafRef.current = null;
        if (onDoneCallback.current) onDoneCallback.current();
      },
      onError: (err) => {
        if (rafRef.current) cancelAnimationFrame(rafRef.current);
        rafRef.current = null;
        if (onErrorCallback.current) onErrorCallback.current(err);
      },
    });

    cleanupRef.current = cancel;
  }, [rafLoop, flushBuffer]);

  const stopStream = useCallback(() => {
    if (abortRef.current) {
      abortRef.current.abort();
      abortRef.current = null;
    }
    if (rafRef.current) {
      cancelAnimationFrame(rafRef.current);
      rafRef.current = null;
    }
    if (cleanupRef.current) {
      cleanupRef.current();
      cleanupRef.current = null;
    }
    flushBuffer();
  }, [flushBuffer]);

  return { startStream, stopStream };
}
