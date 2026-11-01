import { test, expect, type APIRequestContext, type Locator, type Page } from '@playwright/test';

// Issue #140: Greninja ex's Shifting Stream switches your Active [W] Pokémon
// with a Benched one. Using it from the open card must not leave the card open.
// Both decks are pinned, so the draft pool can't shift them.

type Mon = { damage: number; card: { id: string } } | null;
type Card = { id: string; stage: string };
type View = {
  phase: string;
  active: number;
  pending: { player: number } | null;
  players: { hand: Card[]; active: Mon; bench: Mon[] }[];
};

const GRENINJA = 'B1-073';
const FROAKIE = 'A1-087';
const FROGADIER = 'A1-088';
const BENCHER = 'A1-211';
const FILLER = ['A1-211', 'A1-202', 'B4-015', 'B4-112', 'A3-044', 'A2-112', 'A1-150'];
const DECK = [FROAKIE, FROGADIER, GRENINJA, ...FILLER].flatMap((id) => [id, id]);
const BOT = ['A1-211', 'A1-202', 'B4-015', 'B4-112', 'A3-044', 'A2-112', 'A1-150', 'B4-036', 'P-A-001', 'A4b-373'].flatMap((id) => [id, id]);

async function post(api: APIRequestContext, url: string, data?: object): Promise<View> {
  const res = await api.post(url, data ? { data } : undefined);
  const body = JSON.parse(await res.text());
  if (body.error) throw new Error(url + ': ' + body.error);
  return body.state ?? body;
}

const actor = (v: View) => (v.pending ? v.pending.player : v.active);

async function runBotUntilYourTurn(api: APIRequestContext): Promise<View> {
  let v = await post(api, '/api/bot');
  for (let i = 0; i < 40 && v.phase === 'main' && actor(v) !== 0; i++) v = await post(api, '/api/bot');
  return v;
}

// Evolves your Active with whatever the hand allows; refused moves are skipped.
async function evolveAt(api: APIRequestContext, v: View, target = 0): Promise<View> {
  for (const card of v.players[0].hand.filter((c) => c.stage !== 'Basic')) {
    const res = await api.post('/api/move', { data: { type: 'Evolve', player: 0, handCardId: card.id, target } });
    const body = await res.json();
    if (body.ok) return body.state;
  }
  return v;
}

// Plays on, evolving each turn, until a Greninja is your Active and the bot's
// Active survives, on the first seed that gets there.
async function toGreninjaExActive(api: APIRequestContext) {
  for (let seed = 1; seed < 60; seed++) {
    const start = await post(api, '/api/new', { seed, you: DECK, bot: BOT });
    const froakie = start.players[0].hand.find((c) => c.id === FROAKIE);
    if (!froakie) continue;
    await post(api, '/api/move', { type: 'SetupActive', player: 0, cardId: froakie.id });
    const mates = start.players[0].hand.filter((c) => c.id === BENCHER);
    if (mates.length < 2) continue;
    await post(api, '/api/move', { type: 'SetupBench', player: 0, cardId: mates[0].id, slot: 0 });
    await post(api, '/api/move', { type: 'SetupBench', player: 0, cardId: mates[1].id, slot: 1 });
    await post(api, '/api/move', { type: 'SetupReady', player: 0 });
    for (let i = 0; i < 6 && (await post(api, '/api/bot')).phase === 'setup'; i++);
    let v = await runBotUntilYourTurn(api);
    for (let turn = 0; turn < 6 && v.phase === 'main' && v.players[0].active; turn++) {
      v = await evolveAt(api, v);
      if (v.players[0].active?.card.id === GRENINJA) return;
      await post(api, '/api/move', { type: 'EndTurn', player: 0 });
      v = await runBotUntilYourTurn(api);
    }
  }
  throw new Error('no seed gets a Greninja into the Active Spot');
}

// Greninja ex on the Bench behind a Water Active, a second Benched Pokémon beside it.
async function toGreninjaExBenched(api: APIRequestContext) {
  for (let seed = 1; seed < 120; seed++) {
    const start = await post(api, '/api/new', { seed, you: DECK, bot: BOT });
    const hand = start.players[0].hand;
    const froakies = hand.filter((c) => c.id === FROAKIE);
    const mate = hand.find((c) => c.stage === 'Basic' && c.id !== FROAKIE);
    if (froakies.length < 2 || !mate) continue;
    await post(api, '/api/move', { type: 'SetupActive', player: 0, cardId: froakies[0].id });
    await post(api, '/api/move', { type: 'SetupBench', player: 0, cardId: froakies[1].id, slot: 0 });
    await post(api, '/api/move', { type: 'SetupBench', player: 0, cardId: mate.id, slot: 1 });
    await post(api, '/api/move', { type: 'SetupReady', player: 0 });
    for (let i = 0; i < 6 && (await post(api, '/api/bot')).phase === 'setup'; i++);
    let v = await runBotUntilYourTurn(api);
    for (let turn = 0; turn < 8 && v.phase === 'main' && v.players[0].active; turn++) {
      v = await evolveAt(api, v, 1);
      if (v.players[0].bench.some((m) => m?.card.id === GRENINJA)) return;
      await post(api, '/api/move', { type: 'EndTurn', player: 0 });
      v = await runBotUntilYourTurn(api);
    }
  }
  throw new Error('no seed puts a Greninja ex on the Bench');
}

async function tap(page: Page, l: Locator) {
  const b = (await l.boundingBox())!;
  await page.mouse.click(b.x + b.width / 2, b.y + b.height / 2);
}

// Card images are blocked in the sandbox: answer them with a 1px gif.
const GIF = Buffer.from('R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7', 'base64');

for (const [name, size] of [['phone', { width: 375, height: 700 }], ['desktop', { width: 1280, height: 800 }]] as const) {
  test(`Shifting Stream from the open card leaves the board playable (${name})`, async ({ page, request }) => {
    await page.setViewportSize(size);
    await page.route(/digitaloceanspaces/, (r) => r.fulfill({ contentType: 'image/gif', body: GIF }));
    await toGreninjaExActive(request);
    await page.goto('/?battle');
    await expect(page.getByTestId('battle-board')).toBeVisible();
    await expect(page.getByTestId('turn-banner')).toBeHidden({ timeout: 3000 });

    await tap(page, page.getByTestId('active-card'));
    const sheet = page.getByTestId('active-sheet');
    const ability = sheet.getByTestId('sheet-ability');
    await expect(ability).toContainText('Shifting Stream');
    await ability.click();

    // Two Benched Pokémon: the board asks which one to switch in.
    const lit = page.getByTestId('my-bench').locator('[data-lit="true"][data-pickable="true"]').first();
    await expect(lit).toBeVisible();
    await tap(page, lit);

    // The card is gone and Greninja ex is on the Bench; the game goes on.
    await expect(sheet).toHaveCount(0);
    await expect(page.getByTestId('my-bench').locator('[data-zoom="B1-073"]')).toHaveCount(1);
    await page.screenshot({ path: `test-results/shifting-stream-${name}.png` });
    const end = page.getByRole('button', { name: 'End Turn' });
    await expect(end).toBeEnabled();
    await end.click();
    await expect(end).toBeDisabled();
  });
}

test('Shifting Stream from a Benched Greninja ex: choosing it again does not reopen a card', async ({ page, request }) => {
  await page.route(/digitaloceanspaces/, (r) => r.fulfill({ contentType: 'image/gif', body: GIF }));
  await toGreninjaExBenched(request);
  await page.goto('/?battle');
  await expect(page.getByTestId('battle-board')).toBeVisible();
  await expect(page.getByTestId('turn-banner')).toBeHidden({ timeout: 3000 });

  // The Bench fills left to right, so Greninja ex (evolved from the first Benched Froakie) is on the left.
  const greninja = page.getByTestId('my-bench').locator('[data-slot="left"]');
  await expect(greninja.locator('[data-zoom="B1-073"]')).toHaveCount(1);
  await tap(page, greninja);
  const sheet = page.getByTestId('active-sheet');
  await expect(sheet.getByTestId('sheet-ability')).toContainText('Shifting Stream');
  await sheet.getByTestId('sheet-ability').click();

  // Switch Greninja ex itself in: the Active ends up Greninja ex, the old Active on the Bench.
  await tap(page, page.getByTestId('my-bench').locator('[data-slot="left"] [data-lit="true"][data-pickable="true"]'));
  await expect(page.getByTestId('active-card').locator('[data-zoom="B1-073"]')).toHaveCount(1);
  await expect(sheet).toHaveCount(0);
  await page.screenshot({ path: 'test-results/shifting-stream-bench.png' });
  await expect(page.getByRole('button', { name: 'End Turn' })).toBeEnabled();
});
