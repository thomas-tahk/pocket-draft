import { expect, type Locator, type Page } from '@playwright/test';
import { readFileSync } from 'node:fs';

// Drives the real draft: 20 packs, skip shop, deckbuild to 20, finalize.

type Raw = { id: string; name: string; stage: string };
const stageById = new Map(
  (JSON.parse(readFileSync('public/data/cards.json', 'utf8')) as Raw[]).map((c) => [c.id, c]),
);

export const idOf = (title: string) => title.split(' · ').pop() ?? '';

export async function draftTwenty(page: Page): Promise<Set<string>> {
  const picked = new Set<string>();
  const names = new Set<string>();
  for (let pack = 1; pack <= 20; pack++) {
    await expect(page.getByText(`Pack ${pack} / 20`)).toBeVisible();
    const offer = page.locator('main button[title]');
    await expect(offer).toHaveCount(5);
    const ids = (await offer.evaluateAll((els) => els.map((e) => e.getAttribute('title') ?? ''))).map(idOf);
    // A picked name never comes back (#124), so every offer is fresh.
    for (const id of ids) expect(names.has(stageById.get(id)!.name), `pack ${pack} re-offered ${id}`).toBe(false);
    const id = ids.find((i) => stageById.get(i)!.stage === 'Basic') ?? ids[0];
    picked.add(id);
    names.add(stageById.get(id)!.name);
    await offer.nth(ids.indexOf(id)).click();
  }
  return picked;
}

export async function buildFromDraft(page: Page) {
  await expect(page.getByRole('heading', { name: 'Deckbuilding' })).toBeVisible();
  const drafted = page.locator('section', { has: page.getByRole('heading', { name: /^Your draft/ }) });
  await fillFrom(drafted);
  await expect(page.getByText('Deck: 20 / 20')).toBeVisible();
  await page.getByRole('button', { name: 'Finalize deck' }).click();
}

// Adds up to two copies of every card in the section, in order, until 20.
export async function fillFrom(section: Locator) {
  const plus = section.getByRole('button', { name: '+' });
  for (let i = 0; i < (await plus.count()); i++) {
    for (let copy = 0; copy < 2 && (await plus.nth(i).isEnabled()); copy++) await plus.nth(i).click();
  }
}


// A whole real draft, ending on the review screen with the first run active.
export async function draftRealDeck(page: Page) {
  await page.getByRole('button', { name: 'Start draft' }).click();
  await draftTwenty(page);
  await page.getByRole('button', { name: 'Skip shop' }).click();
  await buildFromDraft(page);
}
