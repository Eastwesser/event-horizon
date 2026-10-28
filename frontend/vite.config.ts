import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

const apiProxy = {
  '/api': {
    target: 'http://localhost:8079',
    changeOrigin: true,
    rewrite: (path: string) => path,
  },
  '/uploads': {
    target: 'http://localhost:8079',
    changeOrigin: true,
  },
  '/ws': {
    target: 'ws://localhost:8079',
    ws: true,
    changeOrigin: true,
  },
}

export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    port: 5173,
    proxy: apiProxy,
    // DANGER (TEST ONLY):
    // host: '0.0.0.0',      // ← добавил: слушать все интерфейсы
    // ws: false,            // ← добавил: отключить HMR WebSocket
    // allowedHosts: true,   // ← добавил: пускать любой хост (Quick Tunnel)
  },
  preview: {
    port: 4173,
    proxy: apiProxy,
  },
})
