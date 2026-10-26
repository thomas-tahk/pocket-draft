import { test, expect } from '@playwright/test';
import { readFileSync } from 'node:fs';
import { keepWorkingText, toDraftablePool } from '../src/data/loader';
import { buildSynergy, tallyTypes, weightOf } from '../src/draft/weights';
import type { Card, RawCard } from '../src/types/card';

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

// Eevee ex becomes draftable (owner 2026-10-25, issue #128): it is a Basic, but
// nothing evolves from "Eevee ex", so the final-evolutions pool keeps it; the
// working-text filter keeps it once Veevee ’volve is implemented. Picking an
// Eeveelution then raises its odds (src/draft/eeveeBoost.ts).
test('Eevee ex is in the draft pool and an Eeveelution pick raises its odds', () => {
  const rawCards = JSON.parse(readFileSync('public/data/cards.json', 'utf8')) as RawCard[];
  const pool = keepWorkingText(toDraftablePool(rawCards), { onlyWorkingText: true, ids: filter.ids, working: [] });
  const eeveeEx = pool.find((c) => c.name === 'Eevee ex');
  expect(eeveeEx, 'Eevee ex missing from the draft pool').toBeTruthy();

  const sylveon = pool.find((c) => c.name === 'Sylveon ex')!;
  const other = pool.find((c) => c.name === 'Cresselia ex')!;
  expect(sylveon.evolvesFrom).toBe('Eevee');
  const w = (picks: Card[]) => weightOf(eeveeEx!, tallyTypes(picks), buildSynergy(picks));
  expect(w([sylveon])).toBe(3 * w([other]));
  expect(w([])).toBe(w([other]));
});
