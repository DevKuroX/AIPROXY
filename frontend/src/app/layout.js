import './globals.css';
import AuthGuard from '@/components/AuthGuard';

export const metadata = {
  title: 'AIPROXY Dashboard',
  description: 'AI Gateway Management',
};

export default function RootLayout({ children }) {
  return (
    <html lang="en" suppressHydrationWarning>
      <head>
        <link rel="icon" href="/favicon.ico" sizes="any" />
        <script src="/theme-init.js" />
      </head>
      <body className="font-sans antialiased" style={{ background: 'var(--color-bg)', color: 'var(--color-text-main)' }}>
        <AuthGuard>{children}</AuthGuard>
      </body>
    </html>
  );
}
