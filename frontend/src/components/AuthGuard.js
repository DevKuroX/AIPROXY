'use client';

import { useEffect } from 'react';
import { useRouter, usePathname } from 'next/navigation';
import { isAuthenticated } from '@/lib/auth';

export default function AuthGuard({ children }) {
  const router = useRouter();
  const pathname = usePathname();

  useEffect(() => {
    if (pathname !== '/login' && !isAuthenticated()) {
      router.push('/login');
    }
    if (pathname === '/login' && isAuthenticated()) {
      router.push('/dashboard');
    }
  }, [pathname, router]);

  return children;
}
