import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// The SPA is served at the server root, so base must be absolute so asset URLs
// resolve correctly on deep-link refresh (e.g. /profiles/x/edit).
export default defineConfig({
  base: '/',
  plugins: [react()],
  build: {
    outDir: 'dist',
    emptyOutDir: true,
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (!id.includes('node_modules')) return
          // Keep React runtime as a single singleton chunk.
          if (/[\\/]node_modules[\\/](react|react-dom|scheduler)[\\/]/.test(id)) {
            return 'react-vendor'
          }
          if (id.includes('react-router')) return 'router-vendor'
          if (id.includes('@tanstack')) return 'query-vendor'
          if (id.includes('@radix-ui')) return 'radix-vendor'
          if (id.includes('codemirror') || id.includes('@uiw/react-codemirror') || id.includes('@lezer')) {
            return 'codemirror-vendor'
          }
          if (id.includes('lucide-react')) return 'icons-vendor'
          return 'vendor'
        },
      },
    },
  },
  server: {
    proxy: {
      '/api': 'http://127.0.0.1:4747',
    },
  },
})
