import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8079',
        changeOrigin: true,
        rewrite: (path) => path,
      },
      '/ws': {
        target: 'ws://localhost:8079',
        ws: true,
        changeOrigin: true,
      },
    },
  },
})