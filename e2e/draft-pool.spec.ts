import { test, expect } from '@playwright/test';
import { readFileSync } from 'node:fs';

// TEMPORARY (removed by 4.15): the SPA's draft offers exactly the pool the Go
// server decided has working text, read from public/data/draft-pool.json.

type Raw = { id: string; name: string };
const filter = JSON.parse(readFileSync('public/data/draft-pool.json', 'utf8')) as { ids: string[] };
const cards = JSON.parse(readFileSync('public/data/cards.json', 'utf8')) as Raw[];
const allowed = new Set(filter.ids);
const allowedNames = new Set(cards.filter((c) => allowed.has(c.id)).map((c) => c.name));

test('draft offers only cards whose text the engine plays', async ({ page }) => {
  await page.goto('/');

  await expect(page.getByText(`${filter.ids.length} draftable cards`)).toBeVisible();
  await page.getByRole('button', { name: 'Start draft' }).click();

  const offered = page.locator('img[alt]');
  await expect(offered).toHaveCount(5);
  for (const name of await offered.evaluateAll((imgs) => imgs.map((i) => i.getAttribute('alt') ?? ''))) {
    expect(allowedNames, `offered ${name}`).toContain(name);
  }
});
