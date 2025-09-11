import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig(() => ({
  plugins: [vue()],
  resolve: { alias: { '@': fileURLToPath(new URL('./src', import.meta.url)) } },
  server: {
    port: 5173,
    proxy: {
      // For axios calls in dev
      '/api': {
        target: 'http://localhost:3000',
        changeOrigin: true,
        rewrite: (p) => p.replace(/^\/api/, ''),
      },
      // For login via fetch('/session') in dev
      '/session': {
        target: 'http://localhost:3000',
        changeOrigin: true,
      },
    },
  },
  define: {
    "__API_URL__": JSON.stringify("http://localhost:3000"),
  },
}))
