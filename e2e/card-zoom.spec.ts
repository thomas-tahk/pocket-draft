import { test, expect, type APIRequestContext, type Locator, type Page } from '@playwright/test';

// F2 at phone size: holding a hand card ~350 ms without moving opens it large
// and centred; starting to drag shrinks it back to the normal-size dragged
// card. A quick tap never opens it. Long-press on a board card opens it too.

type Card = { id: string; name: string; stage: string };
type View = { phase: string; active: number; pending: { player: number } | null; players: { hand: Card[]; active: { card: Card } | null }[] };

// Pinned so draft-pool changes can't shift the bot's deck.
const BOT = ['A1-211', 'A1-202', 'B4-015', 'B4-112', 'A3-044', 'A2-112', 'A1-150', 'B4-036', 'P-A-001', 'A4b-373'].flatMap((id) => [id, id]);

async function post(api: APIRequestContext, url: string, data?: object): Promise<View> {
  const res = await api.post(url, data ? { data } : undefined);
  const body = await res.json();
  return body.state ?? body;
}

// Seed 7: you go first; on your second turn a Basic is still in hand and the Bench is empty.
async function toYourSecondTurn(api: APIRequestContext): Promise<{ basic: Card; oppActive: Card }> {
  const start = await post(api, '/api/new', { seed: 7, bot: BOT });
  const basics = start.players[0].hand.filter((c) => c.stage === 'Basic');
  await post(api, '/api/move', { type: 'SetupActive', player: 0, cardId: basics[0].id });
  await post(api, '/api/move', { type: 'SetupReady', player: 0 });
  for (let i = 0; i < 6 && (await post(api, '/api/bot')).phase === 'setup'; i++);
  await post(api, '/api/move', { type: 'EndTurn', player: 0 });
  let v = start;
  for (let i = 0; i < 40; i++) {
    v = await post(api, '/api/bot');
    const actor = v.pending ? v.pending.player : v.active;
    if (v.phase !== 'main' || actor === 0) break;
  }
  return { basic: basics[1], oppActive: v.players[1].active!.card };
}

// A point on the card's visible left edge: hand cards overlap to the right.
async function grabPoint(l: Locator) {
  const b = (await l.boundingBox())!;
  return { x: b.x + Math.min(10, b.width / 2), y: b.y + b.height / 2 };
}

async function centre(l: Locator) {
  const b = (await l.boundingBox())!;
  return { x: b.x + b.width / 2, y: b.y + b.height / 2 };
}

const handCard = (page: Page, id: string) => page.getByTestId('my-hand').locator(`[data-card-id="${id}"]`).first();
const slot = (page: Page, name: string) => page.getByTestId('my-bench').locator(`[data-slot="${name}"]`);

async function open(page: Page) {
  await page.goto('/?battle');
  await expect(page.getByTestId('battle-board')).toBeVisible();
  await expect(page.getByTestId('turn-banner')).toBeHidden({ timeout: 3000 });
}

test('hold a hand card to zoom it; dragging shrinks it back and still plays', async ({ page, request }) => {
  const { basic } = await toYourSecondTurn(request);
  await open(page);
  const zoom = page.getByTestId('card-zoom');
  const p = await grabPoint(handCard(page, basic.id));

  // A quick tap never zooms (it holds the card for tap-tap instead).
  await page.mouse.click(p.x, p.y);
  await page.waitForTimeout(500);
  await expect(zoom).toHaveCount(0);
  await page.mouse.click(195, 20); // cancel the tap-hold

  // Hold without moving: nothing yet at 150 ms, the big centred card by ~350 ms.
  await page.mouse.move(p.x, p.y);
  await page.mouse.down();
  await page.waitForTimeout(150);
  await expect(zoom).toHaveCount(0);
  await expect(zoom).toBeVisible({ timeout: 1000 });
  await expect(zoom.locator('img')).toHaveAttribute('alt', basic.name);
  await page.waitForTimeout(200); // let the 120 ms grow-in finish before measuring
  const board = (await page.getByTestId('battle-board').boundingBox())!;
  const z = (await zoom.getByRole('img').boundingBox())!;
  expect(z.width).toBeGreaterThan(board.width * 0.75);
  expect(Math.abs(z.x + z.width / 2 - (board.x + board.width / 2))).toBeLessThan(4);
  expect(Math.abs(z.y + z.height / 2 - (board.y + board.height / 2))).toBeLessThan(4);
  await page.waitForFunction(() => [...document.images].every((i) => i.complete));
  await page.screenshot({ path: 'test-results/battle-card-zoom.png' });

  // Start dragging: the zoom closes, the normal-size card follows the finger, the drop plays it.
  const to = await centre(slot(page, 'left'));
  await page.mouse.move(p.x, p.y - 20, { steps: 3 });
  await expect(zoom).toHaveCount(0);
  await expect(page.getByTestId('play-ghost')).toBeVisible();
  await page.mouse.move(to.x, to.y, { steps: 6 });
  await page.mouse.up();
  await expect(slot(page, 'left').locator('img')).toHaveAttribute('alt', basic.name);
});

test('releasing a long hold closes the zoom without holding the card; long-press opens a board card', async ({ page, request }) => {
  const { basic, oppActive } = await toYourSecondTurn(request);
  await open(page);
  const zoom = page.getByTestId('card-zoom');

  const p = await grabPoint(handCard(page, basic.id));
  await page.mouse.move(p.x, p.y);
  await page.mouse.down();
  await expect(zoom).toBeVisible({ timeout: 1000 });
  await page.mouse.up();
  await expect(zoom).toHaveCount(0);
  await expect(page.getByTestId('play-hint')).toHaveCount(0);

  const o = await centre(page.getByTestId('opp-active'));
  await page.mouse.move(o.x, o.y);
  await page.mouse.down();
  await expect(zoom).toBeVisible({ timeout: 1000 });
  await expect(zoom.locator('img')).toHaveAttribute('alt', oppActive.name);
  await page.mouse.up();
  await expect(zoom).toHaveCount(0);
});
