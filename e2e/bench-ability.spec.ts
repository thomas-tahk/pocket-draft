import { test, expect, type APIRequestContext, type Locator, type Page } from '@playwright/test';

// F10 at phone size: tap a Benched Pokémon with a usable Ability to open its
// card and use it. Ambipom's Catching Tail ("put a random Pokémon Tool card from
// your deck into your hand") works from the Bench, once per turn. Both decks are
// pinned, so the draft pool can't shift them.

type Card = { id: string; name: string };
type View = { phase: string; active: number; pending: { player: number } | null; players: { hand: Card[] }[] };

const AIPOM = 'B4-132';
const AMBIPOM = 'A3b-059';
const FILLER = ['A1-211', 'A1-202', 'B4-015', 'B4-112', 'A3-044', 'A2-112'];
// Two Tools in the deck so Catching Tail has something to find.
const DECK = [AIPOM, AMBIPOM, 'A2-148', 'A2-147', ...FILLER].flatMap((id) => [id, id]);
const BOT = ['A1-211', 'A1-202', 'B4-015', 'B4-112', 'A3-044', 'A2-112', 'A1-150', 'B4-036', 'P-A-001', 'A4b-373'].flatMap((id) => [id, id]);

async function post(api: APIRequestContext, url: string, data?: object): Promise<View> {
  const res = await api.post(url, data ? { data } : undefined);
  const text = await res.text();
  if (!res.ok()) throw new Error(`${url}: ${text}`);
  const body = JSON.parse(text);
  return body.state ?? body;
}

const actor = (v: View) => (v.pending ? v.pending.player : v.active);

async function runBotUntilYourTurn(api: APIRequestContext): Promise<View> {
  let v = await post(api, '/api/bot');
  for (let i = 0; i < 40 && v.phase === 'main' && actor(v) !== 0; i++) v = await post(api, '/api/bot');
  return v;
}

// Plays on until Ambipom is in your hand with Aipom on your Bench, then evolves.
// The first seed whose opening hand holds an Aipom and another Basic is used.
async function toBenchedAmbipom(api: APIRequestContext) {
  for (let seed = 1; seed < 200; seed++) {
    const start = await post(api, '/api/new', { seed, you: DECK, bot: BOT });
    const hand = start.players[0].hand;
    const aipom = hand.find((c) => c.id === AIPOM);
    const active = hand.find((c) => FILLER.includes(c.id) && c.id !== 'A2-112' && c.id !== 'B4-112');
    if (aipom && active && (await evolveBenched(api, aipom.id, active.id))) return;
  }
  throw new Error('no seed deals an Aipom and a Basic and then an Ambipom');
}

async function evolveBenched(api: APIRequestContext, aipom: string, active: string): Promise<boolean> {
  await post(api, '/api/move', { type: 'SetupActive', player: 0, cardId: active });
  await post(api, '/api/move', { type: 'SetupBench', player: 0, cardId: aipom, slot: 0 });
  await post(api, '/api/move', { type: 'SetupReady', player: 0 });
  for (let i = 0; i < 6 && (await post(api, '/api/bot')).phase === 'setup'; i++);
  let v = await runBotUntilYourTurn(api);
  for (let turn = 0; turn < 5; turn++) {
    const ambipom = v.players[0].hand.find((c) => c.id === AMBIPOM);
    const done = ambipom && (await post(api, '/api/move', { type: 'Evolve', player: 0, handCardId: ambipom.id, target: 1 }));
    if (done) return true;
    await post(api, '/api/move', { type: 'EndTurn', player: 0 });
    v = await runBotUntilYourTurn(api);
  }
  return false;
}

async function tap(page: Page, l: Locator) {
  const b = (await l.boundingBox())!;
  await page.mouse.click(b.x + b.width / 2, b.y + b.height / 2);
}

async function openBoard(page: Page) {
  await page.goto('/?battle');
  await expect(page.getByTestId('battle-board')).toBeVisible();
  await expect(page.getByTestId('turn-banner')).toBeHidden({ timeout: 3000 });
}

const sheet = (page: Page) => page.getByTestId('active-sheet');
const slot = (page: Page, name: 'left' | 'behind' | 'right') => page.getByTestId('my-bench').locator(`[data-slot="${name}"]`);

test('tap a Benched Pokémon: its card lists the Ability; using it works once', async ({ page, request }) => {
  await toBenchedAmbipom(request);
  await openBoard(page);
  const hand = page.getByTestId('my-hand').getByTestId('hand-card');
  const before = await hand.count();

  await tap(page, slot(page, 'left'));
  await expect(sheet(page)).toBeVisible();
  await expect(sheet(page).getByTestId('sheet-attack')).toHaveCount(0);
  await expect(sheet(page).getByTestId('sheet-retreat')).toHaveCount(0);
  const ability = sheet(page).getByTestId('sheet-ability');
  await expect(ability).toContainText('Catching Tail');
  await expect(ability).toHaveAttribute('data-usable', 'true');
  await page.waitForFunction(() => [...document.images].every((i) => i.complete));
  await page.screenshot({ path: 'test-results/bench-ability-open.png' });

  await ability.click();
  await expect(sheet(page)).toHaveCount(0);
  await expect(hand).toHaveCount(before + 1);

  // Once during your turn: the Bench Pokémon has nothing left to use, so it doesn't open.
  await tap(page, slot(page, 'left'));
  await expect(sheet(page)).toHaveCount(0);
});

test('a Benched Pokémon with no Ability to use does not open', async ({ page, request }) => {
  await toBenchedAmbipom(request);
  await openBoard(page);
  await tap(page, slot(page, 'behind')); // an empty slot
  await expect(sheet(page)).toHaveCount(0);
});
