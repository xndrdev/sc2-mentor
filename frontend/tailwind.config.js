/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        // SC2 inspirierte Farben
        'sc2-blue': '#0057B8',
        'sc2-gold': '#FFD700',
        'terran': '#2563eb',
        'zerg': '#7c3aed',
        'protoss': '#f59e0b',
      },
    },
  },
  plugins: [],
}
