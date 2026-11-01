import { test, expect, type APIRequestContext, type Page } from '@playwright/test';

// Owner: "End Turn doesn't glow when there's no other legal action to play".
// With only End Turn (and Concede) legal, End Turn glows like any playable
// target; with any other legal play it doesn't. The server's legal moves are
// rewritten on the way to the page so both cases are exact.

const BOT = ['A1-211', 'A1-202', 'B4-015', 'B4-112', 'A3-044', 'A2-112', 'A1-150', 'B4-036', 'P-A-001', 'A4b-373'].flatMap((id) => [id, id]);
type Move = { type: string; player?: number; cardId?: string };

async function post(api: APIRequestContext, url: string, data?: object) {
  const res = await api.post(url, data ? { data } : undefined);
  const body = await res.json();
  return body.state ?? body;
}

async function toYourSecondTurn(api: APIRequestContext) {
  await toFirstMainTurn(api, { seed: 7, bot: BOT, preset: 'fire' });
  await post(api, '/api/move', { type: 'EndTurn', player: 0 });
  for (let i = 0; i < 40; i++) {
    const v = await post(api, '/api/bot');
    if (v.phase !== 'main' || (v.pending ? v.pending.player : v.active) === 0) return;
  }
}

// Setup through the API, ending on the first main turn (you go first on these seeds).
async function toFirstMainTurn(api: APIRequestContext, game: object) {
  const start = await post(api, '/api/new', game);
  if (!start.players) throw new Error('new game failed: ' + JSON.stringify(start));
  const basics = start.players[0].hand.filter((c: { stage: string }) => c.stage === 'Basic');
  await post(api, '/api/move', { type: 'SetupActive', player: 0, cardId: basics[0].id });
  await post(api, '/api/move', { type: 'SetupReady', player: 0 });
  for (let i = 0; i < 6 && (await post(api, '/api/bot')).phase === 'setup'; i++);
}

async function openWithLegalMoves(page: Page, moves?: Move[]) {
  if (moves) await page.route('**/api/state', async (route) => {
    const res = await route.fetch();
    const view = await res.json();
    view.players[0].legalMoves = moves;
    await route.fulfill({ response: res, json: view });
  });
  await page.route(/^https:\/\/(?!localhost).*\.(png|webp|jpg)/, (r) => r.fulfill({ status: 200, contentType: 'image/svg+xml', body: '<svg xmlns="http://www.w3.org/2000/svg" width="10" height="14"/>' }));
  await page.goto('/?battle');
  await expect(page.getByTestId('battle-board')).toBeVisible();
  await expect(page.getByTestId('turn-banner')).toBeHidden({ timeout: 3000 });
}

test('End Turn glows when it is the only legal play', async ({ page, request }) => {
  await toYourSecondTurn(request);
  await openWithLegalMoves(page, [{ type: 'EndTurn', player: 0 }, { type: 'Concede', player: 0 }]);
  const button = page.getByRole('button', { name: 'End Turn' });
  await expect(button).toHaveClass(/bt-usable/);
  if (process.env.END_TURN_SHOT) await page.screenshot({ path: process.env.END_TURN_SHOT });
});

test('End Turn does not glow while another play exists', async ({ page, request }) => {
  await toYourSecondTurn(request);
  await openWithLegalMoves(page, [{ type: 'EndTurn', player: 0 }, { type: 'Concede', player: 0 }, { type: 'Retreat', player: 0 }]);
  await expect(page.getByRole('button', { name: 'End Turn' })).not.toHaveClass(/bt-usable/);
});

// Play every Basic and Trainer legal until the board offers nothing else: the
// bench fills and any Fossil lands on it, leaving only moves with no control.
async function playOutHand(api: APIRequestContext) {
  for (let i = 0; i < 30; i++) {
    const view = await (await api.get('/api/state')).json();
    const moves: Move[] = view.players[0].legalMoves;
    const play = moves.find((m) => m.type === 'PlayTrainer') ?? moves.find((m) => m.type === 'PlayBasic');
    if (!play) return;
    await post(api, '/api/move', play);
  }
}

const legalTypes = async (api: APIRequestContext) => {
  const view = await (await api.get('/api/state')).json();
  return view.players[0].legalMoves.map((m: Move) => m.type).sort();
};

// No rewrite: a real game whose only legal moves are End Turn and Concede.
test('real game with only End Turn and Concede legal: End Turn glows', async ({ page, request }) => {
  await toFirstMainTurn(request, { seed: 1, preset: 'fire', bot: BOT });
  await playOutHand(request);
  expect(await legalTypes(request)).toEqual(['Concede', 'EndTurn']);
  await openWithLegalMoves(page);
  await expect(page.getByRole('button', { name: 'End Turn' })).toHaveClass(/bt-usable/);
});

// Fossils in hand: DiscardFossil is legal every turn but has no control on the board.
const FOSSIL_DECK = ['A1-211', 'A1-202', 'A4-092', 'A3-044', 'B2b-012', 'A1-216', 'A1-217', 'A2-144', 'A2-145', 'B1-216'].flatMap((id) => [id, id]);

test('real game with a Fossil on the board (DiscardFossil legal): End Turn glows', async ({ page, request }) => {
  await toFirstMainTurn(request, { seed: 1, you: FOSSIL_DECK, bot: FOSSIL_DECK });
  await playOutHand(request);
  expect(await legalTypes(request)).toEqual(['Concede', 'DiscardFossil', 'DiscardFossil', 'EndTurn']);
  await openWithLegalMoves(page);
  await expect(page.getByRole('button', { name: 'End Turn' })).toHaveClass(/bt-usable/);
});
