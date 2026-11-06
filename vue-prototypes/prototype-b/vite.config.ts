import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'

// /api проксируется на mock API, чтобы фронт и API были на одном origin.
// Адрес API настраивается через VITE_API_PROXY (.env.local): порт 8000 может
// быть занят другой копией mock API, тогда нужен свободный порт.
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const apiTarget = env.VITE_API_PROXY || 'http://127.0.0.1:8000'

  return {
    plugins: [vue()],
    server: {
      port: 5174,
      proxy: {
        '/api': { target: apiTarget, changeOrigin: true },
        // Swagger и OpenAPI живут на корне API, а не под /api.
        '/docs': { target: apiTarget, changeOrigin: true },
        '/openapi.json': { target: apiTarget, changeOrigin: true },
      },
    },
    preview: {
      port: 4174,
      proxy: {
        '/api': { target: apiTarget, changeOrigin: true },
        '/docs': { target: apiTarget, changeOrigin: true },
        '/openapi.json': { target: apiTarget, changeOrigin: true },
      },
    },
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
      },
    },
  }
})
