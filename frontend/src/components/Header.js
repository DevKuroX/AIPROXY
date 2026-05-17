'use client';

import { useState, useEffect } from 'react';
import { MdHub, MdMenu, MdDarkMode, MdLightMode } from 'react-icons/md';

export default function Header({ onMenuClick }) {
  const [dark, setDark] = useState(false);
  const [time, setTime] = useState('');

  useEffect(() => {
    setDark(document.documentElement.classList.contains('dark'));
    setTime(new Date().toLocaleString());
    const interval = setInterval(() => setTime(new Date().toLocaleString()), 30000);
    return () => clearInterval(interval);
  }, []);

  function toggleTheme() {
    const next = !dark;
    setDark(next);
    document.documentElement.classList.toggle('dark', next);
    localStorage.setItem('theme', next ? 'dark' : 'light');
  }

  return (
    <header className="shrink-0 flex items-center justify-between gap-3 px-4 lg:px-8 pt-3 pb-2 border-b border-border-subtle bg-surface/60 backdrop-blur-xl lg:bg-transparent lg:backdrop-blur-none z-20 transition-colors duration-300">
      <div className="flex items-center gap-3 lg:hidden shrink-0">
        <button onClick={onMenuClick} className="text-text-main hover:text-primary transition-colors">
          <MdMenu className="text-xl" />
        </button>
      </div>

      <div className="flex flex-col min-w-0 flex-1">
        <div className="flex items-center gap-2">
          <MdHub className="text-primary text-xl lg:text-2xl" />
          <h1 className="text-base lg:text-2xl font-semibold tracking-tight truncate text-text-main">AIPROXY</h1>
        </div>
        <p className="hidden lg:block text-sm text-text-muted truncate">AI Gateway Dashboard</p>
      </div>

      <div className="flex items-center gap-1 shrink-0">
        <button
          onClick={toggleTheme}
          className="flex items-center justify-center size-10 rounded-full text-text-muted hover:text-text-main hover:bg-surface-2 transition-colors"
          aria-label={dark ? 'Switch to light mode' : 'Switch to dark mode'}
        >
          {dark ? <MdLightMode className="text-[22px]" /> : <MdDarkMode className="text-[22px]" />}
        </button>
        <span className="hidden lg:block text-sm text-text-muted ml-2">{time}</span>
      </div>
    </header>
  );
}
