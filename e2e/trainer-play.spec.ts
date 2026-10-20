import { test, expect, type APIRequestContext, type Locator, type Page } from '@playwright/test';

// F6, F7: a targeted Trainer (Potion) glows only its legal targets and plays
// on one; a Trainer with no target (Professor's Research) lights the whole
// board but the hands and plays when dropped anywhere on it.

type Card = { id: string; stage: string };
type Mon = { damage: number; remainingHp: number };
type View = { phase: string; active: number; firstPlayer: number; pending: { player: number } | null; players: { hand: Card[]; active: Mon | null }[] };

const POTION = 'P-A-001';
const RESEARCH = 'A4b-373';
// Sturdy Basics so your Active survives the hit it needs for a Potion.
const DECK = ['A1-211', 'A1-202', 'B4-015', 'B4-112', 'A3-044', 'A2-112', 'A1-150', 'B4-036', POTION, RESEARCH].flatMap((id) => [id, id]);

// Pinned so draft-pool changes can't shift the bot's deck.
const BOT = ['A1-211', 'A1-202', 'B4-015', 'B4-112', 'A3-044', 'A2-112', 'A1-150', 'B4-036', 'P-A-001', 'A4b-373'].flatMap((id) => [id, id]);

async function post(api: APIRequestContext, url: string, data?: object): Promise<View> {
  const res = await api.post(url, data ? { data } : undefined);
  const body = await res.json();
  return body.state ?? body;
}

const actor = (v: View) => (v.pending ? v.pending.player : v.active);

async function runBotUntilYourTurn(api: APIRequestContext): Promise<View> {
  let v = await post(api, '/api/bot');
  for (let i = 0; i < 40 && v.phase === 'main' && actor(v) !== 0; i++) v = await post(api, '/api/bot');
  return v;
}

// Seed 5 with a pinned bot deck: ends your turns until your Active has damage and both
// Trainers are in hand.
async function toDamagedTurn(api: APIRequestContext): Promise<View> {
  const start = await post(api, '/api/new', { seed: 5, you: DECK, bot: BOT });
  const basic = start.players[0].hand.find((c) => c.stage === 'Basic')!;
  await post(api, '/api/move', { type: 'SetupActive', player: 0, cardId: basic.id });
  await post(api, '/api/move', { type: 'SetupReady', player: 0 });
  for (let i = 0; i < 6 && (await post(api, '/api/bot')).phase === 'setup'; i++);
  for (let turn = 0; turn < 6; turn++) {
    await post(api, '/api/move', { type: 'EndTurn', player: 0 });
    const v = await runBotUntilYourTurn(api);
    const ids = v.players[0].hand.map((c) => c.id);
    if (v.players[0].active!.damage > 0 && ids.includes(POTION) && ids.includes(RESEARCH)) return v;
  }
  throw new Error('no damaged turn with both Trainers in hand');
}

async function grabPoint(l: Locator) {
  const b = (await l.boundingBox())!;
  return { x: b.x + Math.min(10, b.width / 2), y: b.y + b.height / 2 };
}

async function drag(page: Page, from: { x: number; y: number }, to: { x: number; y: number }, midway: () => Promise<void>) {
  await page.mouse.move(from.x, from.y);
  await page.mouse.down();
  await page.mouse.move((from.x + to.x) / 2, (from.y + to.y) / 2, { steps: 6 });
  await midway();
  await page.mouse.move(to.x, to.y, { steps: 6 });
  await page.mouse.up();
}

const handCard = (page: Page, id: string) => page.getByTestId('my-hand').locator(`[data-card-id="${id}"]`).first();

test('play a Supporter on the board and a Potion on your damaged Active', async ({ page, request }) => {
  const start = await toDamagedTurn(request);
  const handSize = start.players[0].hand.length;
  const hp = start.players[0].active!.remainingHp;
  await page.goto('/?battle');
  await expect(page.getByTestId('battle-board')).toBeVisible();
  await expect(page.getByTestId('turn-banner')).toBeHidden({ timeout: 3000 });
  const hand = page.getByTestId('my-hand').getByTestId('hand-card');
  await expect(hand).toHaveCount(handSize);

  // Professor's Research: the board between the hands glows; drop it on open board.
  const board = (await page.getByTestId('battle-board').boundingBox())!;
  const openBoard = { x: board.x + board.width / 2, y: board.y + board.height * 0.43 };
  await drag(page, await grabPoint(handCard(page, RESEARCH)), openBoard, async () => {
    await expect(page.getByTestId('board-glow')).toBeVisible();
    await expect(page.getByTestId('play-hint')).toContainText('anywhere on the board');
  });
  await expect(page.getByTestId('board-glow')).toBeHidden();
  await expect(hand).toHaveCount(handSize - 1 + 2);

  // Potion: only your damaged Active glows, not the opponent's.
  const mine = page.getByTestId('my-active').locator('[data-spot="active"]');
  const theirs = page.getByTestId('opp-active').locator('[data-spot="opp-active"]');
  const b = (await mine.boundingBox())!;
  await drag(page, await grabPoint(handCard(page, POTION)), { x: b.x + b.width / 2, y: b.y + b.height / 2 }, async () => {
    await expect(mine).toHaveAttribute('data-glow', 'true');
    await expect(theirs).toHaveAttribute('data-glow', 'false');
    await expect(page.getByTestId('board-glow')).toBeHidden();
  });
  await expect(page.getByTestId('my-active').getByTestId('hp')).toHaveText(String(Math.min(hp + 20, hp + start.players[0].active!.damage)));
  await expect(hand).toHaveCount(handSize);
});
