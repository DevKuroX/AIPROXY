'use client';

import { useState, useRef, useEffect, useCallback } from 'react';
import { MdSend, MdAttachFile, MdAutoAwesome, MdClose, MdStop, MdWarning } from 'react-icons/md';
import useChatStore from '@/lib/chat/store';

export default function ChatInput({ onSend, disabled, streaming, streamingContent, onStop }) {
  const [text, setText] = useState('');
  const [dragOver, setDragOver] = useState(false);
  const textareaRef = useRef(null);
  const fileInputRef = useRef(null);
  const uploadQueue = useChatStore((s) => s.uploadQueue);
  const addToUploadQueue = useChatStore((s) => s.addToUploadQueue);
  const removeFromUploadQueue = useChatStore((s) => s.removeFromUploadQueue);
  const selectedPreset = useChatStore((s) => s.selectedPreset);

  const isThinking = streaming && !streamingContent;
  const hasImages = uploadQueue.some(f => f.type?.startsWith('image/') || /\.(jpg|jpeg|png|gif|webp|svg)$/i.test(f.name));
  const showVisionWarning = hasImages && selectedPreset && !selectedPreset.supportsVision;

  useEffect(() => {
    if (!disabled && textareaRef.current) {
      textareaRef.current.focus();
    }
  }, [disabled]);

  function handleKeyDown(e) {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      submit();
    }
  }

  function submit() {
    const trimmed = text.trim();
    if (!trimmed || disabled) return;
    onSend(trimmed);
    setText('');
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto';
    }
  }

  function handleInput(e) {
    setText(e.target.value);
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto';
      textareaRef.current.style.height = Math.min(textareaRef.current.scrollHeight, 160) + 'px';
    }
  }

  function handleFileSelect(e) {
    const files = Array.from(e.target.files || []);
    files.forEach((f) => addToUploadQueue(f));
    e.target.value = '';
  }

  const handleDragOver = useCallback((e) => {
    e.preventDefault();
    setDragOver(true);
  }, []);

  const handleDragLeave = useCallback(() => {
    setDragOver(false);
  }, []);

  const handleDrop = useCallback((e) => {
    e.preventDefault();
    setDragOver(false);
    const files = Array.from(e.dataTransfer.files || []);
    files.forEach((f) => addToUploadQueue(f));
  }, [addToUploadQueue]);

  return (
    <div
      className="relative border-t border-border-subtle bg-surface/80 backdrop-blur-xl"
      onDragOver={handleDragOver}
      onDragLeave={handleDragLeave}
      onDrop={handleDrop}
    >
      {dragOver && (
        <div className="absolute inset-0 z-10 flex items-center justify-center bg-primary/5 border-2 border-dashed border-primary/50 rounded-t-2xl">
          <p className="text-sm font-medium text-primary">Drop files here</p>
        </div>
      )}

      <div className="max-w-3xl mx-auto px-4 pt-2">
        {uploadQueue.length > 0 && (
          <div className="flex flex-wrap gap-2 mb-2">
            {showVisionWarning && (
              <div className="w-full flex items-center gap-1.5 px-2 py-1 rounded-lg bg-warning/10 text-warning text-xs mb-1">
                <MdWarning className="text-sm" />
                <span>This model may not support images. Consider using a vision-capable model.</span>
              </div>
            )}
            {uploadQueue.map((file, i) => {
              const isImage = file.type?.startsWith('image/') || /\.(jpg|jpeg|png|gif|webp|svg)$/i.test(file.name);
              const previewUrl = isImage ? URL.createObjectURL(file) : null;
              return (
                <div key={i} className="flex items-center gap-2 px-2 py-1 rounded-xl bg-surface-2 border border-border-subtle text-xs group">
                  {previewUrl ? (
                    <img src={previewUrl} alt="" className="size-8 rounded-lg object-cover" />
                  ) : (
                    <MdAttachFile className="text-text-muted text-lg" />
                  )}
                  <span className="text-text-main max-w-20 truncate">{file.name}</span>
                  <button onClick={() => { removeFromUploadQueue(i); if (previewUrl) URL.revokeObjectURL(previewUrl); }} className="text-text-muted hover:text-danger opacity-0 group-hover:opacity-100 transition-opacity">
                    <MdClose className="text-sm" />
                  </button>
                </div>
              );
            })}
          </div>
        )}
      </div>

      <div className="max-w-3xl mx-auto px-4 pb-3">
        {isThinking && (
          <div className="flex items-center justify-between px-4 py-2 mb-2 bg-bg-alt/80 rounded-2xl border border-border-subtle">
            <div className="flex items-center gap-2">
              <div className="flex gap-1">
                <span className="size-1.5 rounded-full bg-primary animate-bounce" style={{ animationDelay: '0ms' }} />
                <span className="size-1.5 rounded-full bg-primary animate-bounce" style={{ animationDelay: '150ms' }} />
                <span className="size-1.5 rounded-full bg-primary animate-bounce" style={{ animationDelay: '300ms' }} />
              </div>
              <span className="text-xs text-text-muted font-medium">Thinking...</span>
            </div>
            {onStop && (
              <button onClick={onStop} className="flex items-center gap-1 px-2.5 py-1 rounded-md bg-danger/10 text-danger text-xs font-medium hover:bg-danger/20 transition-colors">
                <MdStop className="text-sm" />
                Stop
              </button>
            )}
          </div>
        )}
        <div className="flex items-end gap-2 bg-bg-alt rounded-2xl border border-border-subtle focus-within:border-primary/50 focus-within:shadow-focus transition-all px-4 py-2.5">
          <button
            onClick={() => fileInputRef.current?.click()}
            className="text-text-muted hover:text-text-main transition-colors shrink-0 self-end mb-1"
            title="Attach files"
          >
            <MdAttachFile className="text-lg" />
          </button>
          <input ref={fileInputRef} type="file" multiple className="hidden" onChange={handleFileSelect} />

          <textarea
            ref={textareaRef}
            value={text}
            onChange={handleInput}
            onKeyDown={handleKeyDown}
            placeholder="Type a message... (Shift+Enter for new line)"
            rows={1}
            disabled={disabled}
            className="flex-1 resize-none bg-transparent text-sm sm:text-base text-text-main outline-none placeholder:text-text-subtle max-h-32 disabled:opacity-50 py-0.5 leading-relaxed"
          />

          <button
              onClick={submit}
              disabled={disabled || !text.trim()}
              className="flex items-center justify-center size-8 rounded-lg bg-primary text-white hover:bg-primary-hover disabled:opacity-30 disabled:cursor-not-allowed transition-all active:scale-95 shrink-0 self-end mb-1"
            >
              <MdSend className="text-base" />
            </button>
        </div>
        <div className="flex items-center justify-between mt-1.5 px-1">
          {selectedPreset && (
            <div className="flex items-center gap-1 text-xs text-text-muted">
              <MdAutoAwesome className="text-primary text-sm" />
              <span>{selectedPreset.label}</span>
            </div>
          )}
          <div className="text-xs text-text-subtle ml-auto">Shift+Enter for new line</div>
        </div>
      </div>
    </div>
  );
}
