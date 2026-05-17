/** @type {import('tailwindcss').Config} */
export default {
  content: ['./src/**/*.{js,jsx}'],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        brand: {
          50: 'var(--color-brand-50)', 100: 'var(--color-brand-100)',
          200: 'var(--color-brand-200)', 300: 'var(--color-brand-300)',
          400: 'var(--color-brand-400)', 500: 'var(--color-brand-500)',
          600: 'var(--color-brand-600)', 700: 'var(--color-brand-700)',
          800: 'var(--color-brand-800)', 900: 'var(--color-brand-900)',
        },
        primary: 'var(--color-primary)',
        'primary-hover': 'var(--color-primary-hover)',
        accent: 'var(--color-accent)',
        bg: 'var(--color-bg)',
        'bg-alt': 'var(--color-bg-alt)',
        surface: 'var(--color-surface)',
        'surface-2': 'var(--color-surface-2)',
        'surface-3': 'var(--color-surface-3)',
        sidebar: 'var(--color-sidebar)',
        border: 'var(--color-border)',
        'border-subtle': 'var(--color-border-subtle)',
        text: 'var(--color-text)',
        'text-main': 'var(--color-text-main)',
        'text-muted': 'var(--color-text-muted)',
        'text-subtle': 'var(--color-text-subtle)',
        danger: 'var(--color-danger)',
        success: 'var(--color-success)',
        warning: 'var(--color-warning)',
        info: 'var(--color-info)',
      },
      borderRadius: {
        sm: 'var(--radius-sm)',
        md: 'var(--radius-md)',
        lg: 'var(--radius-lg)',
      },
      boxShadow: {
        soft: 'var(--shadow-soft)',
        warm: 'var(--shadow-warm)',
        elev: 'var(--shadow-elev)',
        focus: 'var(--shadow-focus)',
      },
      fontFamily: {
        sans: ['Inter', '-apple-system', 'BlinkMacSystemFont', 'SF Pro Text', 'SF Pro Display', 'system-ui', 'sans-serif'],
      },
    },
  },
  plugins: [],
};
