import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

// No @types/node here; PD_API lets parallel test runs point at their own server.
declare const process: { env: Record<string, string | undefined> };

// The Go simulator serves the game API on :8080. In dev the Vite server (:5173)
// proxies /api there so client code can use same-origin relative URLs.
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: {
      '/api': process.env.PD_API ?? 'http://localhost:8080',
    },
  },
});
