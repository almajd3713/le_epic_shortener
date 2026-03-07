import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    react({
      babel: {
        plugins: [['babel-plugin-react-compiler']],
      },
    }),
    tailwindcss(),
  ],
  server: {
    host: true,
    // Proxy /api/* to the Go backend during development.
    // This lets the browser make same-origin requests, sidestepping CORS for dev.
    // In production, the backend must send proper CORS headers.
    proxy: {
      '/api': {
        target: process.env.VITE_API_URL ?? 'http://localhost:8080',
        changeOrigin: true,
      },
      // Forward short-code redirects (/:code) to the backend.
      // Regex excludes known frontend assets so the SPA still loads normally.
      '^/[^/]+$': {
        target: process.env.VITE_API_URL ?? 'http://localhost:8080',
        changeOrigin: true,
        bypass(req) {
          // Let Vite serve the SPA entry point and static assets normally.
          if (req.url === '/' || req.url?.match(/\.(html|js|ts|css|ico|png|svg|map)(\?.*)?$/)) {
            return req.url;
          }
        },
      },
    },
  },
})
