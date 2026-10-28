// Screenshots of every non-battle screen in the M4 blended look, light and dark, at 390×844 (#131).
// Run from the repo root:  npx playwright test -c docs/chief/look/m4-blend/render.config.ts
// Writes docs/chief/look/m4-blend/<screen>-<light|dark>-390.png.
//
// The light pass drives the real app (draft 20, skip shop, build, concede one battle, quit) and keeps a
// copy of localStorage at each screen. The dark pass seeds those copies, so both themes show the same
// cards and the same run. The card-image CDN is blocked in the cloud, so card images are a local
// placeholder card (as in e2e/pocket-look.spec.ts).
import { test, expect, type Page } from '@playwright/test';
import { fileURLToPath } from 'node:url';
import path from 'node:path';
import { buildFromDraft, draftTwenty } from '../../../../e2e/helpers/draft';

const OUT = path.dirname(fileURLToPath(import.meta.url));
const KEY = 'pocket-draft-v0.4';
const IMAGE_HOST = '**://limitlesstcg.nyc3.cdn.digitaloceanspaces.com/**';
const PLACEHOLDER = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 250 350">
<rect width="250" height="350" rx="14" fill="#f4f1e6"/>
<rect x="10" y="10" width="230" height="330" rx="10" fill="#fff8d8" stroke="#c9b458" stroke-width="4"/>
<rect x="24" y="48" width="202" height="150" rx="6" fill="#bfe3f2"/>
<circle cx="125" cy="123" r="42" fill="#7fc4e0"/>
<rect x="24" y="22" width="120" height="14" rx="7" fill="#8a8466"/>
<rect x="24" y="214" width="150" height="10" rx="5" fill="#8a8466"/>
<rect x="24" y="236" width="202" height="8" rx="4" fill="#bdb58f"/>
<rect x="24" y="254" width="180" height="8" rx="4" fill="#bdb58f"/>
<rect x="24" y="272" width="190" height="8" rx="4" fill="#bdb58f"/>
</svg>`;

type Theme = 'light' | 'dark';

const settled = (page: Page) =>
  page.waitForFunction(() =>
    [...document.images].every((i) => {
      const r = i.getBoundingClientRect();
      const onScreen = r.bottom > 0 && r.top < innerHeight && r.right > 0 && r.left < innerWidth;
      return !onScreen || (i.complete && i.naturalWidth > 0);
    }),
  );

async function shot(page: Page, name: string, theme: Theme) {
  await page.evaluate(() => window.scrollTo(0, 0));
  await settled(page);
  // The sheen is off under reduced motion, so shots run with motion on and every sheen frozen mid-sweep.
  await page.addStyleTag({ content: '.look *::after { animation-play-state: paused !important; animation-delay: -1.1s !important; }' });
  await page.waitForTimeout(300);
  await page.screenshot({ path: path.join(OUT, `${name}-${theme}-390.png`) });
}

const saved = (page: Page) => page.evaluate((k) => localStorage.getItem(k)!, KEY);
async function seed(page: Page, state: string) {
  await page.evaluate(([k, v]) => localStorage.setItem(k, v), [KEY, state]);
  await page.goto('/');
}

test('every non-battle screen, light then dark', async ({ page }) => {
  test.setTimeout(300_000);
  await page.route(IMAGE_HOST, (route) => route.fulfill({ status: 200, contentType: 'image/svg+xml', body: PLACEHOLDER }));
  const offers = page.locator('.offer-grid button');
  const snap: Record<string, string> = {};

  // ---- light: drive the app ----
  await page.emulateMedia({ colorScheme: 'light', reducedMotion: 'no-preference' });
  await page.goto('/');
  await expect(page.getByRole('button', { name: 'Start draft' })).toBeVisible();
  await shot(page, 'home', 'light');

  await page.getByRole('button', { name: 'Start draft' }).click();
  await expect(offers).toHaveCount(5);
  snap.draft = await saved(page);
  await shot(page, 'draft', 'light');

  await draftTwenty(page);
  await expect(page.getByRole('heading', { name: 'Shop' })).toBeVisible();
  snap.shop = await saved(page);
  await shot(page, 'shop', 'light');

  await page.getByRole('button', { name: 'Skip shop' }).click();
  await expect(page.getByRole('heading', { name: 'Deckbuilding' })).toBeVisible();
  snap.deckbuild = await saved(page);
  await shot(page, 'deckbuild', 'light');

  await buildFromDraft(page);
  await page.getByRole('button', { name: 'Next battle' }).click();
  await expect(page.getByTestId('battle-board')).toBeVisible();
  await page.getByTestId('menu').click();
  await page.getByTestId('menu-concede').click();
  await page.getByTestId('confirm-ok').click();
  await expect(page.getByTestId('end-result')).toHaveText('Loss');
  await page.getByRole('button', { name: 'Back' }).click();
  await expect(page.getByTestId('run-record')).toHaveText('Run: 0–1 (active)');
  snap.run = await saved(page);
  await shot(page, 'run', 'light');

  await page.getByRole('button', { name: 'Quit run' }).click();
  await page.getByTestId('quit-confirm').click();
  await expect(page.getByRole('main', { name: 'Run summary' })).toBeVisible();
  snap.summary = await saved(page);
  await shot(page, 'summary', 'light');

  // A won run (five wins) shows the gold title plate: the same run with its battle copied as wins.
  const won = JSON.parse(snap.summary);
  const b = won.state.run.battles[0];
  won.state.run.battles = [1, 2, 3, 4, 5].map((seed) => ({ ...b, seed: b.seed + seed, result: 'win' }));
  won.state.run.status = 'won';
  snap.won = JSON.stringify(won);

  await page.getByRole('button', { name: 'Home' }).click();
  await page.getByRole('button', { name: /Past runs/ }).click();
  await expect(page.getByRole('heading', { name: 'Draft history' })).toBeVisible();
  await page.getByRole('button', { name: 'View' }).first().click();
  await shot(page, 'history', 'light');

  const summaryWon = async (theme: Theme) => {
    await seed(page, snap.won);
    await page.getByRole('button', { name: 'View last run' }).click();
    await expect(page.getByTestId('run-status')).toHaveText('Run won');
    await shot(page, 'summary-won', theme);
  };
  await summaryWon('light');

  // ---- dark: the same states, seeded ----
  await page.emulateMedia({ colorScheme: 'dark', reducedMotion: 'no-preference' });
  await page.evaluate(() => localStorage.clear());
  await page.goto('/');
  await expect(page.getByRole('button', { name: 'Start draft' })).toBeVisible();
  await shot(page, 'home', 'dark');

  await seed(page, snap.draft);
  await expect(offers).toHaveCount(5);
  await shot(page, 'draft', 'dark');

  await seed(page, snap.shop);
  await expect(page.getByRole('heading', { name: 'Shop' })).toBeVisible();
  await shot(page, 'shop', 'dark');

  await seed(page, snap.deckbuild);
  await expect(page.getByRole('heading', { name: 'Deckbuilding' })).toBeVisible();
  await shot(page, 'deckbuild', 'dark');

  await seed(page, snap.run);
  await page.getByRole('button', { name: 'Continue run' }).click();
  await expect(page.getByTestId('run-record')).toHaveText('Run: 0–1 (active)');
  await shot(page, 'run', 'dark');

  await seed(page, snap.summary);
  await page.getByRole('button', { name: 'View last run' }).click();
  await expect(page.getByRole('main', { name: 'Run summary' })).toBeVisible();
  await shot(page, 'summary', 'dark');

  await page.getByRole('button', { name: 'Home' }).click();
  await page.getByRole('button', { name: /Past runs/ }).click();
  await expect(page.getByRole('heading', { name: 'Draft history' })).toBeVisible();
  await page.getByRole('button', { name: 'View' }).first().click();
  await shot(page, 'history', 'dark');

  await summaryWon('dark');
});
