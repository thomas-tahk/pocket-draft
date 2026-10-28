// Playwright config for the M4 blended-look screenshots (#131). Same Go server + Vite dev server as
// the e2e suite (playwright.config.ts), but it runs only render.spec.ts in this folder.
// Run from the repo root:  npx playwright test -c docs/chief/look/m4-blend/render.config.ts
import { defineConfig } from '@playwright/test';
import { fileURLToPath } from 'node:url';
import path from 'node:path';
import base from '../../../../playwright.config';

const here = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(here, '../../../..');
const servers = Array.isArray(base.webServer) ? base.webServer : base.webServer ? [base.webServer] : [];

export default defineConfig({
  ...base,
  testDir: here,
  testMatch: ['render.spec.ts'],
  outputDir: path.join(root, 'test-results/m4-blend'),
  // render.spec.ts turns motion on and freezes the holo sheen mid-sweep, so shots are deterministic.
  use: { ...base.use, viewport: { width: 390, height: 844 }, deviceScaleFactor: 2 },
  webServer: servers.map((s) => ({ ...s, cwd: root })),
});
