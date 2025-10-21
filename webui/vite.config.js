import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig(() => {
  const API_URL = "http://localhost:3000"; // allowed here (config-only)

  return {
    plugins: [vue()],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
      },
    },
    define: {
      // Do not modify this constant, it is used in the evaluation.
      "__API_URL__": JSON.stringify(API_URL),
    },
    server: {
      proxy: {
        // Dev API calls go through Vite on :5173 and are proxied to the Go backend
        '/api': {
          target: API_URL,
          changeOrigin: true,
          rewrite: p => p.replace(/^\/api/, ''), // strip /api → backend expects /
        },
        // Let uploads go through the same origin in dev (no hardcoded host in app code)
        '/uploads': {
          target: API_URL,
          changeOrigin: true,
        },
      },
      host: true,          // bind 0.0.0.0 so Docker port mapping works
      port: 5173,
      strictPort: true,
    },
  }
})
