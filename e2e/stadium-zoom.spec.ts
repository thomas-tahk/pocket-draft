import { test, expect, type APIRequestContext } from '@playwright/test';

// F2 for the Stadium: holding the Stadium in play opens it large and centred,
// like a Pokémon in play, so its text can be read (play-test 2026-10-27).

type Card = { id: string; name: string; stage: string };
type View = { phase: string; active: number; pending: { player: number } | null; stadium?: Card | null; players: { hand: Card[] }[] };

const STADIUM = 'B2-153'; // Training Area
const DECK = ['A1-211', 'A1-202', 'B4-015', 'B4-112', 'A3-044', 'A2-112', 'A1-150', 'B4-036', STADIUM, 'A4b-373'].flatMap((id) => [id, id]);
// Pinned so draft-pool changes can't shift the bot's deck.
const BOT = ['A1-211', 'A1-202', 'B4-015', 'B4-112', 'A3-044', 'A2-112', 'A1-150', 'B4-036', 'P-A-001', 'A4b-373'].flatMap((id) => [id, id]);

async function post(api: APIRequestContext, url: string, data?: object): Promise<View> {
  const res = await api.post(url, data ? { data } : undefined);
  const body = await res.json();
  return body.state ?? body;
}

const actor = (v: View) => (v.pending ? v.pending.player : v.active);

// Ends your turns until Training Area is in hand, then plays it.
async function toStadiumInPlay(api: APIRequestContext): Promise<Card> {
  const start = await post(api, '/api/new', { seed: 5, you: DECK, bot: BOT });
  const basic = start.players[0].hand.find((c) => c.stage === 'Basic')!;
  await post(api, '/api/move', { type: 'SetupActive', player: 0, cardId: basic.id });
  await post(api, '/api/move', { type: 'SetupReady', player: 0 });
  let v = start;
  for (let i = 0; i < 6 && (v = await post(api, '/api/bot')).phase === 'setup'; i++);
  for (let turn = 0; turn < 8; turn++) {
    for (let i = 0; i < 40 && v.phase === 'main' && actor(v) !== 0; i++) v = await post(api, '/api/bot');
    if (v.players[0].hand.some((c) => c.id === STADIUM)) {
      v = await post(api, '/api/move', { type: 'PlayTrainer', player: 0, cardId: STADIUM });
      if (v.stadium?.id === STADIUM) return v.stadium;
    }
    v = await post(api, '/api/move', { type: 'EndTurn', player: 0 });
  }
  throw new Error('Training Area never reached play');
}

test('holding the Stadium in play zooms it; releasing closes it', async ({ page, request }) => {
  const stadium = await toStadiumInPlay(request);
  await page.goto('/?battle');
  await expect(page.getByTestId('battle-board')).toBeVisible();
  await expect(page.getByTestId('turn-banner')).toBeHidden({ timeout: 3000 });
  const zoom = page.getByTestId('card-zoom');

  const spot = page.getByTestId('stadium');
  await expect(spot.locator('img')).toHaveAttribute('alt', stadium.name);
  const b = (await spot.boundingBox())!;
  await page.mouse.move(b.x + b.width / 2, b.y + b.height / 2);
  await page.mouse.down();
  await expect(zoom).toBeVisible({ timeout: 1000 });
  await expect(zoom.locator('img')).toHaveAttribute('alt', stadium.name);
  await page.waitForTimeout(200); // let the grow-in finish
  const board = (await page.getByTestId('battle-board').boundingBox())!;
  const z = (await zoom.getByRole('img').boundingBox())!;
  expect(z.width).toBeGreaterThan(board.width * 0.75);
  await page.screenshot({ path: 'test-results/stadium-zoom.png' });
  await page.mouse.up();
  await expect(zoom).toHaveCount(0);
});
