import { defineConfig } from 'vite';
import { resolve } from 'path';

const entry = process.env.ENTRY || 'app';

export default defineConfig({
  build: {
    outDir: 'static/js',
    emptyOutDir: false,
    lib: {
      entry: resolve(__dirname, `ts/${entry}.ts`),
      name: entry === 'app' ? 'Villum' : 'PWA',
      formats: ['iife'],
      fileName: () => `${entry}.js`,
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
