import { defineConfig, devices } from '@playwright/test';

// Battle screen checks run on a phone-sized screen against the real Go server
// (:8080) behind the Vite dev server (:5173), the same path the app uses in dev.
export default defineConfig({
  testDir: 'e2e',
  use: { baseURL: 'http://localhost:5173', ...devices['Pixel 7'], viewport: { width: 390, height: 844 } },
  projects: [{ name: 'phone', use: { browserName: 'chromium' } }],
  webServer: [
    { command: 'cd server && go build -o /tmp/pd-e2e-server . && /tmp/pd-e2e-server', url: 'http://localhost:8080/api/state', reuseExistingServer: true, timeout: 120_000 },
    { command: 'npx vite --port 5173 --strictPort', url: 'http://localhost:5173', reuseExistingServer: true },
  ],
});
