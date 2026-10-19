import { test, expect, type Locator, type Page, type Request } from '@playwright/test';
import { readFileSync } from 'node:fs';

// 1.31 at phone size: a draft finished in the real app starts its battle on the
// new board (src/battle) with the drafted 20 as your deck, and Rematch from the
// end screen starts again with that same deck.
//
// The draft picks Basics first and new names first. Deckbuild offers only the
// cards the engine can play (2.10), so any 20 built there reach the board whole.

type Raw = { id: string; name: string; stage: string };
const stageById = new Map(
  (JSON.parse(readFileSync('public/data/cards.json', 'utf8')) as Raw[]).map((c) => [c.id, c]),
);

const idOf = (title: string) => title.split(' · ').pop() ?? '';
const working = new Set(
  (JSON.parse(readFileSync('public/data/draft-pool.json', 'utf8')) as { working: string[] }).working,
);

async function draftSixteen(page: Page): Promise<Set<string>> {
  const picked = new Set<string>();
  const names = new Set<string>();
  for (let pack = 1; pack <= 16; pack++) {
    await expect(page.getByText(`Pack ${pack} / 16`)).toBeVisible();
    const offer = page.locator('main button[title]');
    await expect(offer).toHaveCount(5);
    const ids = (await offer.evaluateAll((els) => els.map((e) => e.getAttribute('title') ?? ''))).map(idOf);
    const fresh = ids.filter((id) => !names.has(stageById.get(id)!.name));
    const basic = fresh.find((id) => stageById.get(id)!.stage === 'Basic');
    const id = basic ?? fresh[0] ?? ids[0];
    picked.add(id);
    names.add(stageById.get(id)!.name);
    await offer.nth(ids.indexOf(id)).click();
  }
  return picked;
}

async function buildFromDraft(page: Page) {
  await expect(page.getByRole('heading', { name: 'Deckbuilding' })).toBeVisible();
  const drafted = page.locator('section', { has: page.getByRole('heading', { name: /^Your draft/ }) });
  await fillFrom(drafted);
  await expect(page.getByText('Deck: 20 / 20')).toBeVisible();
  await page.getByRole('button', { name: 'Finalize deck' }).click();
}

// Adds up to two copies of every card in the section, in order, until 20.
async function fillFrom(section: Locator) {
  const plus = section.getByRole('button', { name: '+' });
  for (let i = 0; i < (await plus.count()); i++) {
    for (let copy = 0; copy < 2 && (await plus.nth(i).isEnabled()); copy++) await plus.nth(i).click();
  }
}

const deckSent = (req: Request) => (req.postDataJSON() as { you: string[] }).you;
const isNewGame = (req: Request) => req.url().endsWith('/api/new') && req.method() === 'POST';

// The opening coin (S4): tap it if it's yours, then wait for setup's board.
async function passIntro(page: Page) {
  const coin = page.getByTestId('setup-coin');
  await expect(page.getByTestId('setup-intro')).toBeVisible();
  if (await coin.isEnabled()) await coin.click();
  await expect(page.getByTestId('setup-intro')).toHaveCount(0, { timeout: 5000 });
  await expect(page.getByTestId('setup-ready')).toBeVisible();
}

async function yourCardCount(page: Page) {
  const v = await (await page.request.get('/api/state')).json();
  return v.players[0].handCount + v.players[0].deckCount;
}

test('a finished draft starts on the new board with its 20, and Rematch keeps them', async ({ page }) => {
  test.setTimeout(90_000);
  await page.goto('/');
  await page.getByRole('button', { name: 'Start draft' }).click();
  const picked = await draftSixteen(page);
  await page.getByRole('button', { name: 'Skip shop' }).click();
  await buildFromDraft(page);

  const started = page.waitForRequest(isNewGame);
  await page.getByRole('button', { name: /Play deck/ }).click();
  const deck = deckSent(await started);

  await expect(page.getByTestId('battle-board')).toBeVisible();
  await passIntro(page);
  expect(deck).toHaveLength(20);
  for (const id of deck) expect(picked, `deck card ${id}`).toContain(id);
  expect(await yourCardCount(page)).toBe(20);
  await page.waitForFunction(() => [...document.images].every((i) => i.complete));
  await page.screenshot({ path: 'test-results/draft-to-battle-390x844.png' });

  await page.getByTestId('menu').click();
  await page.getByTestId('menu-concede').click();
  await page.getByTestId('confirm-ok').click();
  const end = page.getByTestId('end-screen');
  await expect(end).toBeVisible();

  const rematch = page.waitForRequest(isNewGame);
  await end.getByRole('button', { name: 'Rematch' }).click();
  expect([...deckSent(await rematch)].sort()).toEqual([...deck].sort());
  await expect(end).toBeHidden();
  await passIntro(page);
  expect(await yourCardCount(page)).toBe(20);
});

test('deckbuild offers only cards the engine can play, so any deck built there plays all 20', async ({ page }) => {
  test.setTimeout(90_000);
  await page.goto('/');
  await page.getByRole('button', { name: 'Start draft' }).click();
  await draftSixteen(page);
  await page.getByRole('button', { name: 'Skip shop' }).click();
  await expect(page.getByRole('heading', { name: 'Deckbuilding' })).toBeVisible();

  const offered = (await page.locator('main section [title]').evaluateAll((els) => els.map((e) => e.getAttribute('title') ?? ''))).map(idOf);
  expect(offered.length).toBeGreaterThan(0);
  for (const id of offered) expect(working, `deckbuild offers ${id}`).toContain(id);

  // Universals and under-evolutions first, then the draft, so the deck uses
  // every kind of card deckbuild can add.
  for (const name of [/^Free universals/, /^Free expansions/, /^Your draft/]) {
    await fillFrom(page.locator('section', { has: page.getByRole('heading', { name }) }));
  }
  await expect(page.getByText('Deck: 20 / 20')).toBeVisible();
  await page.getByRole('button', { name: 'Finalize deck' }).click();

  const started = page.waitForRequest(isNewGame);
  await page.getByRole('button', { name: /Play deck/ }).click();
  expect(deckSent(await started)).toHaveLength(20);

  await expect(page.getByTestId('battle-board')).toBeVisible();
  await passIntro(page);
  expect(await yourCardCount(page)).toBe(20);
});
