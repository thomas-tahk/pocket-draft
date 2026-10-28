import { defineConfig, devices } from '@playwright/test';

// The leave / resume specs against the production build (vite build + vite preview), in front of
// the same Go server the dev config uses. The dev server runs <StrictMode> double effects that
// hide timing races between a resume and the closed-tab expiry; the built app does not.
// Ports come from this checkout's path (offset from the dev config's, so both can run at once).
const hash = [...process.cwd()].reduce((h, c) => (h * 31 + c.charCodeAt(0)) >>> 0, 7) % 900;
const apiPort = 20100 + hash;
const webPort = 21100 + hash;

export default defineConfig({
  // One Go server holds one global game, so tests must not run in parallel.
  workers: 1,
  testDir: 'e2e',
  testMatch: ['leave-battle.spec.ts'],
  use: { baseURL: `http://localhost:${webPort}`, ...devices['Pixel 7'], viewport: { width: 390, height: 844 }, contextOptions: { reducedMotion: 'reduce' } },
  projects: [{ name: 'phone-prod', use: { browserName: 'chromium' } }],
  webServer: [
    { command: `cd server && go build -o ../.pd-e2e-prod-server . && PD_ADDR=:${apiPort} ../.pd-e2e-prod-server`, url: `http://localhost:${apiPort}/api/state`, reuseExistingServer: false, timeout: 120_000 },
    // vite preview proxies /api with the same server.proxy setting as dev.
    { command: `npm run build && PD_API=http://localhost:${apiPort} npx vite preview --port ${webPort} --strictPort`, url: `http://localhost:${webPort}`, reuseExistingServer: false, timeout: 180_000 },
  ],
});
