import { test, expect } from '@playwright/test';
import { readFileSync } from 'node:fs';

// #107: no free universal Trainers. Deckbuild offers a Trainer only if the
// player drafted it (or bought it), so a finished deck holds no other Trainer.

type Raw = { id: string; cardType: string };
const typeById = new Map(
  (JSON.parse(readFileSync('public/data/cards.json', 'utf8')) as Raw[]).map((c) => [c.id, c.cardType]),
);
const idOf = (title: string) => title.split(' · ').pop() ?? '';

test('deckbuild offers no Trainer that was not drafted', async ({ page }) => {
  test.setTimeout(60_000);
  await page.goto('/');
  await page.getByRole('button', { name: 'Start draft' }).click();

  const picked = new Set<string>();
  // The run length comes from the "Pack n / N" header, not a hardcoded count.
  const header = page.getByText(/^Pack \d+ \/ \d+$/);
  await expect(header).toBeVisible();
  const total = Number((await header.textContent())!.match(/\/ (\d+)/)![1]);
  for (let pack = 1; pack <= total; pack++) {
    await expect(page.getByText(`Pack ${pack} / ${total}`)).toBeVisible();
    const offer = page.locator('main button[title]');
    await expect(offer).toHaveCount(5);
    const ids = (await offer.evaluateAll((els) => els.map((e) => e.getAttribute('title') ?? ''))).map(idOf);
    // Take a Pokémon so the picks hold no Trainer: any Trainer offered at
    // deckbuild would then be a free one.
    const pick = Math.max(0, ids.findIndex((id) => typeById.get(id) !== 'Trainer'));
    picked.add(ids[pick]);
    await offer.nth(pick).click();
  }
  await page.getByRole('button', { name: 'Skip shop' }).click();
  await expect(page.getByRole('heading', { name: 'Deckbuilding' })).toBeVisible();

  await expect(page.getByRole('heading', { name: /^Free universals/ })).toHaveCount(0);
  const offered = (await page.locator('main section [title]').evaluateAll((els) => els.map((e) => e.getAttribute('title') ?? ''))).map(idOf);
  expect(offered.length).toBeGreaterThan(0);
  for (const id of offered.filter((id) => typeById.get(id) === 'Trainer')) {
    expect(picked, `deckbuild offers unpicked Trainer ${id}`).toContain(id);
  }
  await page.screenshot({ path: 'test-results/no-free-trainers-390x844.png' });
});
