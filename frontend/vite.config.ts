import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  test: {
    environment: "jsdom",
  },
  server: {
    // Mirrors what nginx does in the image, so the relative /v1 the app uses
    // resolves the same way in dev as it does in a container.
    proxy: {
      '/v1': {
        target: process.env.VITE_DEV_API_TARGET ?? 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
