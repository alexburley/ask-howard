import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
  server: {
    port: 5173,
    host: true,
    allowedHosts: ['web'],
    proxy: {
      '/api': process.env.API_URL ?? 'http://localhost:8080',
    },
  },
})
