/** @type {import('tailwindcss').Config} */

export default {
  content: [
    "./index.html",
    // This glob pattern ensures all your Vue components and scripts are included
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  theme: {},
  plugins: [],
}
