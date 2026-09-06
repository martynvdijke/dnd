import { defineConfig } from 'vite';
import { resolve } from 'path';

const dir = import.meta.dirname;

export default defineConfig({
  build: {
    outDir: 'static/js',
    emptyOutDir: true,
    rollupOptions: {
      input: {
        app: resolve(dir, 'ts/app.ts'),
        admin: resolve(dir, 'ts/admin.ts'),
        pwa: resolve(dir, 'ts/pwa.ts'),
        setup: resolve(dir, 'ts/setup.ts'),
        login: resolve(dir, 'ts/login.ts'),
      },
      output: {
        entryFileNames: '[name].js',
        chunkFileNames: 'chunks/[name]-[hash].js',
        format: 'es',
        manualChunks(id) {
          if (id.includes('ts/lib/') || id.includes('dompurify')) return 'shared';
        },
      },
    },
    sourcemap: false,
    minify: true,
  },
});
