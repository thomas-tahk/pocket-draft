import { test, expect, type Page } from '@playwright/test';
import { readFileSync } from 'node:fs';
import { idOf } from './helpers/draft';

// M4 QA: a double-tap on a draft offer makes exactly one pick; a disabled Finalize says why on screen.
// Card images are blocked in this sandbox, so the CDN is answered with a placeholder (as pocket-look.spec.ts).

const IMAGE_HOST = '**://limitlesstcg.nyc3.cdn.digitaloceanspaces.com/**';
const PLACEHOLDER = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 250 350"><rect width="250" height="350" rx="14" fill="#f4f1e6"/></svg>`;
const KEY = 'pocket-draft-v0.4';

type Raw = { id: string; stage: string };
const stageById = new Map((JSON.parse(readFileSync('public/data/cards.json', 'utf8')) as Raw[]).map((c) => [c.id, c.stage]));
const basicIds = [...stageById].filter(([, s]) => s === 'Basic').map(([id]) => id);

async function start(page: Page) {
  await page.route(IMAGE_HOST, (route) => route.fulfill({ status: 200, contentType: 'image/svg+xml', body: PLACEHOLDER }));
  await page.goto('/');
  await page.getByRole('button', { name: 'Start draft' }).click();
  await expect(page.getByText('Pack 1 / 20')).toBeVisible();
  await expect(page.locator('main button[title]')).toHaveCount(5);
}

const picks = (page: Page) => page.evaluate((k) => (JSON.parse(localStorage.getItem(k)!).state.pickIds as string[]).length, KEY);

test.use({ viewport: { width: 390, height: 844 }, hasTouch: true, isMobile: true });

for (const gap of [30, 80, 150, 300]) {
  test(`a double-tap ${gap} ms apart picks exactly one card`, async ({ page }) => {
    await start(page);
    const box = (await page.locator('main button[title]').first().boundingBox())!;
    const x = box.x + box.width / 2;
    const y = box.y + box.height / 2;
    await page.touchscreen.tap(x, y);
    await page.waitForTimeout(gap);
    await page.touchscreen.tap(x, y);
    await expect(page.getByText('Pack 2 / 20')).toBeVisible();
    await page.waitForTimeout(600); // let any late second pick land
    await expect(page.getByText('Pack 2 / 20')).toBeVisible();
    expect(await picks(page)).toBe(1);
  });
}

test('a deliberate tap after the settle time still picks', async ({ page }) => {
  await start(page);
  await page.locator('main button[title]').first().tap();
  await expect(page.getByText('Pack 2 / 20')).toBeVisible();
  await page.waitForTimeout(500);
  await page.locator('main button[title]').first().tap();
  await expect(page.getByText('Pack 3 / 20')).toBeVisible();
});

test('two tiles of one pack clicked in the same tick pick only one', async ({ page }) => {
  await start(page);
  await page.locator('main button[title]').evaluateAll((els) => {
    (els[0] as HTMLElement).click();
    (els[1] as HTMLElement).click();
  });
  await expect(page.getByText('Pack 2 / 20')).toBeVisible();
  expect(await picks(page)).toBe(1);
});

test('five scripted clicks on one pack pick only one', async ({ page }) => {
  await start(page);
  await page.locator('main button[title]').evaluateAll((els) => els.forEach((e) => (e as HTMLElement).click()));
  await expect(page.getByText('Pack 2 / 20')).toBeVisible();
  expect(await picks(page)).toBe(1);
});

// ---------- B: the Finalize reason is on screen ----------

for (const [w, h] of [[390, 844], [1440, 900]]) {
  test(`a 20-card deck with no Basic says why Finalize is off, at ${w}`, async ({ page }) => {
    await page.setViewportSize({ width: w, height: h });
    await start(page);
    for (let pack = 1; pack <= 20; pack++) {
      await expect(page.getByText(`Pack ${pack} / 20`)).toBeVisible();
      const tiles = page.locator('main button[title]');
      await expect(tiles).toHaveCount(5);
      const ids = (await tiles.evaluateAll((els) => els.map((e) => e.getAttribute('title') ?? ''))).map(idOf);
      const i = ids.findIndex((id) => stageById.get(id) !== 'Basic');
      await tiles.nth(i < 0 ? 0 : i).click();
    }
    await page.getByRole('button', { name: 'Skip shop' }).click();
    await expect(page.getByRole('heading', { name: 'Deckbuilding' })).toBeVisible();
    // Deck = the drafted cards that are not Basic, one each (a picked name never comes back).
    const ok = await page.evaluate(
      ([k, basics]) => {
        const all = JSON.parse(localStorage.getItem(k)!);
        const deck: Record<string, number> = {};
        for (const id of all.state.pickIds as string[]) if (!basics.includes(id)) deck[id] = 1;
        if (Object.keys(deck).length !== 20) return false;
        all.state.deck = deck;
        localStorage.setItem(k, JSON.stringify(all));
        return true;
      },
      [KEY, basicIds] as [string, string[]],
    );
    test.skip(!ok, 'this draft had a pack with only Basic cards');
    await page.reload();
    const status = page.locator('[role=status]', { hasText: 'Deck:' });
    await expect(status).toContainText('Deck: 20 / 20');
    await expect(page.getByText('Add a Basic Pokémon to finalize')).toBeVisible();
    await expect(page.getByRole('button', { name: 'Finalize deck' })).toBeDisabled();
    expect(await status.textContent()).not.toContain(' . ');
    expect(await status.textContent()).not.toMatch(/\.\s*Add/);
  });
}
