import { defineConfig } from 'vite';
import { resolve } from 'path';

export default defineConfig({
  build: {
    outDir: 'static/js',
    emptyOutDir: false,
    lib: {
      entry: resolve(__dirname, 'ts/app.ts'),
      name: 'Villum',
      formats: ['iife'],
      fileName: () => 'app.js',
    },
    rollupOptions: {
      output: {
        extend: true,
      },
    },
    sourcemap: false,
    minify: false,
  },
});
