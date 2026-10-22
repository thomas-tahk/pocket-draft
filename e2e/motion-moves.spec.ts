import { test, expect, type APIRequestContext } from '@playwright/test';
import { animationsSoFar, logAnimations, settled } from './fx-log';

// M3-3: the moves that are not a plain hand-to-Bench drop. Draws and plays of a
// whole game, evolution and knock-outs are covered in full-game.spec.ts (motion on).

type Card = { id: string; stage: string };
type View = { phase: string; active: number; pending: { player: number } | null; players: { hand: Card[] }[] };

const POTION = 'P-A-001';
const RESEARCH = 'A4b-373';
const BOT = ['A1-211', 'A1-202', 'B4-015', 'B4-112', 'A3-044', 'A2-112', 'A1-150', 'B4-036', POTION, RESEARCH].flatMap((id) => [id, id]);

async function post(api: APIRequestContext, url: string, data?: object): Promise<View> {
  const res = await api.post(url, data ? { data } : undefined);
  const body = await res.json();
  return body.state ?? body;
}

async function botPlaysUntilYou(api: APIRequestContext) {
  for (let i = 0; i < 40; i++) {
    const v = await post(api, '/api/bot');
    if (v.phase !== 'main' || (v.pending ? v.pending.player : v.active) === 0) return;
  }
}

// Seed 5, as in trainer-play.spec.ts: end turns until Professor's Research is in your hand.
async function toResearchTurn(api: APIRequestContext) {
  const start = await post(api, '/api/new', { seed: 5, you: BOT, bot: BOT });
  const basic = start.players[0].hand.find((c) => c.stage === 'Basic')!;
  await post(api, '/api/move', { type: 'SetupActive', player: 0, cardId: basic.id });
  await post(api, '/api/move', { type: 'SetupReady', player: 0 });
  for (let i = 0; i < 6 && (await post(api, '/api/bot')).phase === 'setup'; i++);
  for (let turn = 0; turn < 6; turn++) {
    await post(api, '/api/move', { type: 'EndTurn', player: 0 });
    await botPlaysUntilYou(api);
    if (((await (await api.get('/api/state')).json()) as View).players[0].hand.some((c) => c.id === RESEARCH)) return;
  }
  throw new Error("no Professor's Research in hand");
}

test('a Trainer shows in the middle and goes to the discard; the cards it draws fly to the hand', async ({ page, request }) => {
  await toResearchTurn(request);
  await page.emulateMedia({ reducedMotion: 'no-preference' });
  await logAnimations(page);
  await page.goto('/?battle');
  await expect(page.getByTestId('turn-banner')).toBeHidden({ timeout: 3000 });
  await settled(page);
  const board = (await page.getByTestId('battle-board').boundingBox())!;
  const hand = (await page.getByTestId('my-hand').locator(`[data-card-id="${RESEARCH}"]`).first().boundingBox())!;
  await page.mouse.move(hand.x + 10, hand.y + hand.height / 2);
  await page.mouse.down();
  await page.mouse.move(board.x + board.width / 2, board.y + board.height * 0.43, { steps: 8 });
  await page.mouse.up();
  await expect(page.getByTestId('my-discard')).toContainText('1');
  const log = await animationsSoFar(page);
  const trainer = log.filter((a) => a.copy && a.stops === 4);
  expect(trainer, 'one copy of the Trainer flies via the middle').toHaveLength(1);
  expect(trainer[0].props).toContain('translate');
  expect(log.filter((a) => a.zone === 'my-hand' && a.key.startsWith('p0:')).length, 'the drawn cards and the re-fanned hand move').toBeGreaterThan(0);
  await settled(page);
  await expect(page.locator('.bt-board > .bt-card')).toHaveCount(0); // the copy is gone
});

// Seed 7 as in active-card.spec.ts: Entei ex Active, Heatran on the Bench, Fire on the Active.
test('retreat swaps the two Pokémon', async ({ page, request }) => {
  await post(request, '/api/new', { seed: 7, bot: BOT });
  await post(request, '/api/move', { type: 'SetupActive', player: 0, cardId: 'B2a-126' });
  await post(request, '/api/move', { type: 'SetupBench', player: 0, cardId: 'A4b-076', slot: 0 });
  await post(request, '/api/move', { type: 'SetupReady', player: 0 });
  for (let i = 0; i < 6 && (await post(request, '/api/bot')).phase === 'setup'; i++);
  await botPlaysUntilYou(request);
  await post(request, '/api/move', { type: 'AttachEnergy', player: 0, target: 0 });
  await post(request, '/api/move', { type: 'EndTurn', player: 0 });
  await botPlaysUntilYou(request);
  await post(request, '/api/move', { type: 'AttachEnergy', player: 0, target: 0 });

  await page.emulateMedia({ reducedMotion: 'no-preference' });
  await logAnimations(page);
  await page.goto('/?battle');
  await expect(page.getByTestId('turn-banner')).toBeHidden({ timeout: 3000 });
  await settled(page);
  await page.getByTestId('active-card').click();
  await page.getByTestId('sheet-retreat').click();
  await page.getByTestId('my-bench').locator('[data-slot="left"]').click();
  await expect(page.getByTestId('my-active').locator('img')).toHaveAttribute('alt', 'Heatran');
  const log = await animationsSoFar(page);
  expect(log.some((a) => a.zone === 'my-active' && a.key.startsWith('p0:A4b-076#')), 'Heatran moves up to the Active').toBe(true);
  expect(log.some((a) => a.zone === 'my-bench' && a.key.startsWith('p0:B2a-126#')), 'Entei ex moves down to the Bench').toBe(true);
});
