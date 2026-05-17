'use client';

import { useMemo } from 'react';
import { MdAutoAwesome } from 'react-icons/md';
import ReactMarkdown from 'react-markdown';
import rehypeHighlight from 'rehype-highlight';
import remarkGfm from 'remark-gfm';

function CodeBlock({ className, children, ...props }) {
  return (
    <pre className={className}>
      <code {...props}>{children}</code>
    </pre>
  );
}

function FilePreview({ file }) {
  if (!file) return null;
  if (file.type?.startsWith('image/') || /\.(jpg|jpeg|png|gif|webp|svg)$/i.test(file.name || file.url)) {
    return (
      <div className="mb-2 rounded-lg overflow-hidden border border-border-subtle bg-surface-2">
        <img src={file.url} alt={file.name || 'Uploaded image'} className="max-w-full max-h-48 object-contain" />
      </div>
    );
  }
  return (
    <div className="mb-2 inline-flex items-center gap-2 px-3 py-1.5 rounded-lg bg-surface-2 border border-border-subtle text-xs">
      <span className="text-text-muted">📎</span>
      <span className="text-text-main truncate max-w-32">{file.name || 'File'}</span>
    </div>
  );
}

export default function ChatMessage({ message, isStreaming }) {
  const isUser = message.role === 'user';
  const content = message.content || '';

  const shouldRenderMarkdown = useMemo(() => {
    if (isUser) return false;
    const hasMarkdown = /[#*`\[\]()_]/.test(content);
    return hasMarkdown || content.includes('```') || content.length > 200;
  }, [content, isUser]);

  return (
    <div className={`flex px-4 sm:px-8 py-4 ${isUser ? 'justify-end' : 'bg-surface/40'}`}>
      {!isUser && (
        <div className="flex items-center justify-center size-8 rounded-full shrink-0 mt-0.5 mr-3 bg-gradient-to-br from-brand-500 to-brand-700 text-white shadow-warm relative">
          <MdAutoAwesome className="text-lg" />
          {isStreaming && (
            <span className="absolute -bottom-0.5 -right-0.5 flex gap-0.5">
              <span className="size-1.5 rounded-full bg-primary animate-bounce" style={{ animationDelay: '0ms' }} />
              <span className="size-1.5 rounded-full bg-primary animate-bounce" style={{ animationDelay: '150ms' }} />
              <span className="size-1.5 rounded-full bg-primary animate-bounce" style={{ animationDelay: '300ms' }} />
            </span>
          )}
        </div>
      )}
      <div className={`max-w-[75%] min-w-0 ${isUser ? 'order-first' : ''}`}>
        <div className={`rounded-2xl px-4 py-2.5 ${isUser ? 'bg-primary/10 text-text-main' : 'bg-surface border border-border-subtle'}`}>
          <div className="text-sm sm:text-base leading-relaxed break-words space-y-2">
            {message.file_preview && <FilePreview file={message.file_preview} />}
            {content ? (
              shouldRenderMarkdown ? (
                <div className="prose prose-sm dark:prose-invert max-w-none prose-pre:bg-surface-2 prose-pre:border prose-pre:border-border-subtle prose-code:text-sm prose-code:bg-surface-2 prose-code:px-1 prose-code:rounded">
                  <ReactMarkdown rehypePlugins={[rehypeHighlight]} remarkPlugins={[remarkGfm]}>
                    {content}
                  </ReactMarkdown>
                </div>
              ) : (
                <span className="whitespace-pre-wrap">{content}</span>
              )
            ) : null}
            {isStreaming && !content && (
              <span className="inline-block w-2 h-4 bg-primary ml-0.5 animate-pulse rounded-sm align-text-bottom typing-cursor" />
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
