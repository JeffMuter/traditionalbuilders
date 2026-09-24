/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./templates/**/*.templ",
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        // Warm leather browns for light-mode surfaces (buttons, footers).
        leather: {
          50: '#faf6f1',
          100: '#f3e9dd',
          200: '#e6d1b8',
          300: '#d6b28c',
          400: '#c49060',
          500: '#b0743f',
          600: '#9a5e2f',
          700: '#7d4a28',
          800: '#653d24',
          900: '#543320',
          950: '#2e1a10',
        },
        // Very light beige/cream for text on dark surfaces in light mode.
        cream: {
          50: '#fdfaf3',
          100: '#faf3e4',
          200: '#f3e6cb',
          300: '#e9d6ac',
        },
      },
    },
  },
  plugins: [],
}
