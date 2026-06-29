/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{ts,tsx,js,jsx}'],
  theme: {
    extend: {
      fontFamily: {
        sans: ['"Space Grotesk"', '"IBM Plex Sans"', 'Avenir Next', 'sans-serif'],
      },
      colors: {
        surface: {
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
    },
  },
  plugins: [],
}
