import { test, expect, type Page, type Request } from '@playwright/test';
import { readFileSync } from 'node:fs';
import { buildFromDraft, draftTwenty, fillFrom, idOf } from './helpers/draft';

// 1.31 at phone size: a draft finished in the real app starts its battle on the
// new board (src/battle) with the drafted 20 as your deck, and Rematch from the
// end screen starts again with that same deck.
//
// The draft picks Basics first and new names first. Deckbuild offers only the
// cards the engine can play (2.10), so any 20 built there reach the board whole.

const working = new Set(
  (JSON.parse(readFileSync('public/data/draft-pool.json', 'utf8')) as { working: string[] }).working,
);

const deckSent = (req: Request) => (req.postDataJSON() as { you: string[] }).you;
const isNewGame = (req: Request) => req.url().endsWith('/api/new') && req.method() === 'POST';

// The opening coin (S4) never waits (2.11): with no tap it lands, shows the
// result and who goes first, then setup's board is up.
async function passIntro(page: Page) {
  await expect(page.getByTestId('setup-intro')).toBeVisible();
  await expect(page.getByTestId('setup-coin')).toHaveAttribute('data-face', /^(heads|tails)$/);
  await expect(page.getByTestId('setup-order')).toHaveText(/^You go (first|second)$/);
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
  const picked = await draftTwenty(page);
  await page.getByRole('button', { name: 'Skip shop' }).click();
  await buildFromDraft(page);

  const started = page.waitForRequest(isNewGame);
  await page.getByRole('button', { name: /Next battle/ }).click();
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
  await draftTwenty(page);
  await page.getByRole('button', { name: 'Skip shop' }).click();
  await expect(page.getByRole('heading', { name: 'Deckbuilding' })).toBeVisible();

  const offered = (await page.locator('main section [title]').evaluateAll((els) => els.map((e) => e.getAttribute('title') ?? ''))).map(idOf);
  expect(offered.length).toBeGreaterThan(0);
  for (const id of offered) expect(working, `deckbuild offers ${id}`).toContain(id);

  // Under-evolutions first, then the draft, so the deck uses every kind of
  // card deckbuild can add.
  for (const name of [/^Free expansions/, /^Your draft/]) {
    await fillFrom(page.locator('section', { has: page.getByRole('heading', { name }) }));
  }
  await expect(page.getByText('Deck: 20 / 20')).toBeVisible();
  await page.getByRole('button', { name: 'Finalize deck' }).click();

  const started = page.waitForRequest(isNewGame);
  await page.getByRole('button', { name: /Next battle/ }).click();
  expect(deckSent(await started)).toHaveLength(20);

  await expect(page.getByTestId('battle-board')).toBeVisible();
  await passIntro(page);
  expect(await yourCardCount(page)).toBe(20);
});
