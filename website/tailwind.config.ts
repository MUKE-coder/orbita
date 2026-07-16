import type { Config } from 'tailwindcss'

// Theme + layout mirror the animateicons project (MIT, © Avijit Dey) — see
// NOTICE. Coral-on-black, glass surfaces, hairline borders. Content is Orbita's.
const config: Config = {
  darkMode: 'class',
  content: [
    './app/**/*.{ts,tsx,md,mdx}',
    './components/**/*.{ts,tsx}',
    './mdx-components.tsx',
  ],
  theme: {
    extend: {
      colors: {
        bgDark: '#000000',
        primary: {
          DEFAULT: '#f45b48',
          hover: '#e04e3d',
          glow: 'rgba(244,91,72,0.25)',
        },
        surface: '#0b0b0b',
        surfaceElevated: '#161616',
        surfaceHover: '#111111',
        textPrimary: '#e5e7eb',
        textSecondary: '#b0b3b8',
        textMuted: '#7c7c7c',
        textDisabled: '#575757',
        hair: '#1f2933',
        divider: '#2a2c2f',
        success: '#22c55e',
        warning: '#f59e0b',
        error: '#ef4444',
        info: '#38bdf8',
      },
      fontFamily: {
        sans: ['var(--font-geist-sans)', 'ui-sans-serif', 'system-ui', 'sans-serif'],
        mono: ['var(--font-geist-mono)', 'ui-monospace', 'SFMono-Regular', 'Menlo', 'monospace'],
      },
      maxWidth: { content: '46rem' },
      keyframes: {
        'fade-up': {
          '0%': { opacity: '0', transform: 'translateY(10px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' },
        },
        pulseDot: {
          '0%,100%': { opacity: '1' },
          '50%': { opacity: '0.4' },
        },
      },
      animation: {
        'fade-up': 'fade-up 0.5s ease-out both',
        'pulse-dot': 'pulseDot 1.6s ease-in-out infinite',
      },
    },
  },
  plugins: [],
}

export default config
