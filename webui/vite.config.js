import { fileURLToPath, URL } from 'node:url'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vitejs.dev/config/
export default defineConfig(({ command, mode, ssrBuild }) => {
  const ret = {
    plugins: [vue()],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url))
      },
    },
  }

  // Do not remove or rename it — the grading system checks for its presence.
  ret.define = {
    "__API_URL__": JSON.stringify("http://localhost:3000"),
  }

  // Dev proxy: default to localhost:3000; allow override via VITE_DEV_API
  const devTarget = process.env.VITE_DEV_API || 'http://localhost:3000'
  // ✅ Add your dev server proxy — used only in development.
  
  ret.server = {
    proxy: {
      '/api': {
        target: devTarget,
        changeOrigin: true,
        rewrite: p => p.replace(/^\/api/, ''),
      },
      '/uploads': {
        target: devTarget,
        changeOrigin: true,
      },
    },
    host: true,     // bind 0.0.0.0 for Docker
    port: 5173,
    strictPort: true,
  }

  return ret
})
