import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

// The Go simulator serves the game API on :8080. In dev the Vite server (:5173)
// proxies /api there so client code can use same-origin relative URLs.
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/api': 'http://localhost:8080',
    },
  },
});
