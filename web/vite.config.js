import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  build: { outDir: 'dist' },
  server: {
    proxy: {
      '/emby': 'http://localhost:8096',
      '/api': 'http://localhost:8096'
    }
  }
})
