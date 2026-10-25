import { defineConfig, devices } from '@playwright/test';

// Battle screen checks run on a phone-sized screen against the real Go server
// behind the Vite dev server, the same path the app uses in dev. Ports come from
// this checkout's path, so parallel worktrees never test each other's servers.
const hash = [...process.cwd()].reduce((h, c) => (h * 31 + c.charCodeAt(0)) >>> 0, 7) % 900;
const apiPort = 18100 + hash;
const webPort = 19100 + hash;

export default defineConfig({
  // One Go server holds one global game, so tests must not run in parallel.
  workers: 1,
  testDir: 'e2e',
  // Reduced motion by default: cards sliding around would race the gesture tests. motion.spec.ts turns motion back on.
  use: { baseURL: `http://localhost:${webPort}`, ...devices['Pixel 7'], viewport: { width: 390, height: 844 }, contextOptions: { reducedMotion: 'reduce' } },
  projects: [{ name: 'phone', use: { browserName: 'chromium' } }],
  webServer: [
    { command: `cd server && go build -o ../.pd-e2e-server . && PD_ADDR=:${apiPort} ../.pd-e2e-server`, url: `http://localhost:${apiPort}/api/state`, reuseExistingServer: false, timeout: 120_000 },
    { command: `PD_API=http://localhost:${apiPort} npx vite --port ${webPort} --strictPort`, url: `http://localhost:${webPort}`, reuseExistingServer: false },
  ],
});
