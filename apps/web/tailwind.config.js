/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{ts,tsx,js,jsx}'],
  darkMode: ['class', '[data-theme="light"]'],
  theme: {
    extend: {
      fontFamily: {
        sans: ['"Space Grotesk"', '"IBM Plex Sans"', 'Avenir Next', 'sans-serif'],
      },
      colors: {
        // Design tokens (from src/index.css :root / [data-theme="light"]).
        // Use as: bg-surface, text-muted, border-border-strong, bg-accent-soft, etc.
        'bg-1': 'var(--bg-1)',
        'bg-2': 'var(--bg-2)',
        surface: 'var(--surface)',
        'surface-2': 'var(--surface-2)',
        'surface-hover': 'var(--surface-hover)',
        border: 'var(--border)',
        'border-strong': 'var(--border-strong)',
        text: 'var(--text)',
        muted: 'var(--text-muted)',
        dim: 'var(--text-dim)',
        accent: 'var(--accent)',
        'accent-hover': 'var(--accent-hover)',
        'accent-soft': 'var(--accent-soft)',
        success: 'var(--success)',
        'success-soft': 'var(--success-soft)',
        warning: 'var(--warning)',
        'warning-soft': 'var(--warning-soft)',
        danger: 'var(--danger)',
        'danger-soft': 'var(--danger-soft)',
        info: 'var(--info)',
        'info-soft': 'var(--info-soft)',
        // Existing blue brand palette (kept for backward compat with existing
        // components and recharts axes).
        brand: {
          50: '#edf2ff',
          100: '#dce4ff',
          200: '#bdd0ff',
          300: '#9cb7ff',
          400: '#6a93ff',
          500: '#396eff',
          600: '#2655cc',
          700: '#1b3d99',
          800: '#152b73',
          900: '#0f1d52',
        },
      },
      borderRadius: {
        sm: 'var(--radius-sm)',
        md: 'var(--radius-md)',
        lg: 'var(--radius-lg)',
        xl: 'var(--radius-xl)',
      },
      boxShadow: {
        sm: 'var(--shadow-sm)',
        md: 'var(--shadow-md)',
        lg: 'var(--shadow-lg)',
      },
      transitionDuration: {
        fast: '150ms',
      },
    },
  },
  plugins: [],
}
