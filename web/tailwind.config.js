/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        primary: {
          50: 'var(--primary-50)',
          100: 'var(--primary-100)',
          200: 'var(--primary-200)',
          300: 'var(--primary-300)',
          400: 'var(--primary-400)',
          500: 'var(--primary-500)',
          600: 'var(--primary-600)',
          700: 'var(--primary-700)',
          800: 'var(--primary-800)',
          900: 'var(--primary-900)',
        },
        theme: {
          bg: 'var(--theme-bg)',
          text: 'var(--theme-text)',
          textLight: 'var(--theme-textLight)',
          border: 'var(--theme-border)',
          hover: 'var(--theme-hover)',
          surface: 'var(--surface)',
        },
      },
      boxShadow: {
        card: '0 1px 3px rgba(55, 53, 47, 0.06), 0 4px 12px rgba(55, 53, 47, 0.04)',
        code: '0 8px 30px rgba(55, 53, 47, 0.12)',
      },
      backgroundImage: {
        'hero-glow':
          'radial-gradient(ellipse 80% 50% at 50% -20%, rgba(55, 53, 47, 0.06), transparent)',
        'card-glow':
          'radial-gradient(ellipse 60% 40% at 50% 0%, rgba(55, 53, 47, 0.04), transparent)',
      },
      animation: {
        'fade-up': 'fadeUp 0.6s ease-out forwards',
      },
      keyframes: {
        fadeUp: {
          '0%': { opacity: '0', transform: 'translateY(16px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' },
        },
      },
    },
  },
  plugins: [],
}
