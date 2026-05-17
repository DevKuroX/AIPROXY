'use client';

import { useState, useCallback } from 'react';

let toastId = 0;
let setToastsFn = null;

export function toast(message, type = 'success') {
  if (setToastsFn) {
    const id = ++toastId;
    setToastsFn(prev => [...prev, { id, message, type }]);
    setTimeout(() => {
      setToastsFn(prev => prev.filter(t => t.id !== id));
    }, 3000);
  }
}

export default function ToastProvider({ children }) {
  const [toasts, setToasts] = useState([]);
  setToastsFn = setToasts;

  const style = useCallback((type) => {
    if (type === 'success') return 'bg-green-900/90 text-green-200 border border-green-700/50';
    if (type === 'error') return 'bg-red-900/90 text-red-200 border border-red-700/50';
    return 'bg-surface-2 text-text-main border border-border-subtle';
  }, []);

  return (
    <>
      {children}
      <div className="fixed bottom-4 right-4 z-[9999] flex flex-col gap-2 pointer-events-none">
        {toasts.map(t => (
          <div key={t.id} className={`px-4 py-2.5 rounded-lg text-sm shadow-lg animate-fade-in backdrop-blur-sm ${style(t.type)}`}>
            {t.message}
          </div>
        ))}
      </div>
    </>
  );
}
