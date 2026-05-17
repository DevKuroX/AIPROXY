'use client';

import { useState } from 'react';

export default function ProviderIcon({ providerId, name, size = 32, className = '' }) {
  const [errored, setErrored] = useState(false);
  const src = `/providers/${providerId}.png`;
  const fallback = (name || providerId || '?')[0].toUpperCase();

  if (!providerId || errored) {
    return (
      <span className={`inline-flex items-center justify-center font-bold rounded-lg text-primary bg-primary/10 ${className}`}
        style={{ width: size, height: size, fontSize: Math.max(10, Math.floor(size * 0.38)) }}>
        {fallback}
      </span>
    );
  }

  return (
    <img
      src={src}
      alt={name || providerId}
      width={size}
      height={size}
      className={`rounded-lg ${className}`}
      onError={() => setErrored(true)}
      style={{ objectFit: 'contain' }}
    />
  );
}
