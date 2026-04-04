import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    vueDevTools(),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    },
  },
  server: {
    host: '127.0.0.1',
    port: 3000,
    proxy: {
      // 代理 API 请求到后端
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        // SSE 需要忽略这些头部，避免代理干扰
        bypass: (req, res, options) => {
          // 对于 SSE 请求，直接返回 false 让代理处理
          if (req.headers.accept && req.headers.accept.includes('text/event-stream')) {
            return false
          }
        }
      }
    }
  },
})
