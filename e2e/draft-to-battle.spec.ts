import { test, expect, type Page, type Request } from '@playwright/test';
import { readFileSync } from 'node:fs';

// 1.31 at phone size: a draft finished in the real app starts its battle on the
// new board (src/battle) with the drafted 20 as your deck, and Rematch from the
// end screen starts again with that same deck.
//
// The draft picks Basics first and new names first, and deckbuild takes only
// drafted cards: the free universals and under-evolutions don't all play yet
// (docs/questions/M1.md [1.30]), so this path keeps every card playable.

type Raw = { id: string; name: string; stage: string };
const stageById = new Map(
  (JSON.parse(readFileSync('public/data/cards.json', 'utf8')) as Raw[]).map((c) => [c.id, c]),
);

const idOf = (title: string) => title.split(' · ').pop() ?? '';

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

async function buildFromDraft(page: Page, pokeBalls = 0) {
  await expect(page.getByRole('heading', { name: 'Deckbuilding' })).toBeVisible();
  const universals = page.locator('section', { has: page.getByRole('heading', { name: /^Free universals/ }) });
  const plusButton = page.getByRole('button', { name: '+' });
  const pokeBall = universals
    .locator('div', { has: page.getByAltText('Poké Ball') })
    .filter({ has: plusButton })
    .last()
    .locator(plusButton);
  for (let i = 0; i < pokeBalls; i++) await pokeBall.click();
  const drafted = page.locator('section', { has: page.getByRole('heading', { name: /^Your draft/ }) });
  const plus = drafted.getByRole('button', { name: '+' });
  for (let i = 0; i < (await plus.count()); i++) {
    for (let copy = 0; copy < 2 && (await plus.nth(i).isEnabled()); copy++) await plus.nth(i).click();
  }
  await expect(page.getByText('Deck: 20 / 20')).toBeVisible();
  await page.getByRole('button', { name: 'Finalize deck' }).click();
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
  await expect(page.getByTestId('deck-short')).toBeHidden();
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

test('a drafted deck with a card the engine cannot play yet says so on the board', async ({ page }) => {
  test.setTimeout(90_000);
  await page.goto('/');
  await page.getByRole('button', { name: 'Start draft' }).click();
  await draftSixteen(page);
  await page.getByRole('button', { name: 'Skip shop' }).click();
  await buildFromDraft(page, 1);

  const started = page.waitForRequest(isNewGame);
  await page.getByRole('button', { name: /Play deck/ }).click();
  expect(deckSent(await started)).toHaveLength(20);

  await expect(page.getByTestId('battle-board')).toBeVisible();
  await passIntro(page);
  expect(await yourCardCount(page)).toBe(19);
  await expect(page.getByTestId('deck-short')).toContainText('1');
  await page.waitForFunction(() => [...document.images].every((i) => i.complete));
  await page.screenshot({ path: 'test-results/draft-to-battle-short-390x844.png' });
});
