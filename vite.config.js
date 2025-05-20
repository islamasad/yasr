import { defineConfig } from 'vite';
import tailwind from '@tailwindcss/vite';

export default defineConfig({
  plugins: [tailwind()],
  server: {
    cors: {
      origin: 'http://localhost:8080',
    },
    port: 5173,
  },
  build: {
    outDir: 'dist',
    manifest: true,
    rollupOptions: {
      input: 'src/main.js',
    },
  },
});