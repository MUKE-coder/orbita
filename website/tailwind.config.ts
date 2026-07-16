import type { Config } from 'tailwindcss'

// Orbita brand tokens (design guide): dark-first, indigo/cyan, hairline borders.
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
        bg: '#0B0D12',
        elev1: '#12151C',
        elev2: '#1A1E27',
        hair: '#252A34',
        ink: '#E6E8EC',
        muted: '#9AA0AA',
        faint: '#5C626D',
        indigo: { DEFAULT: '#6D5CE7', hover: '#5B4BD1', soft: 'rgba(109,92,231,0.14)' },
        cyan: '#22D3EE',
        lime: '#C9F31D',
      },
      fontFamily: {
        sans: ['var(--font-inter)', 'ui-sans-serif', 'system-ui', 'sans-serif'],
        mono: ['var(--font-mono)', 'ui-monospace', 'SFMono-Regular', 'Menlo', 'monospace'],
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
