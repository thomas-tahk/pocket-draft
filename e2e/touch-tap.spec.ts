import { test, expect, type APIRequestContext, type CDPSession, type Locator, type Page } from '@playwright/test';

// Owner 2026-10-19 (5): tap-tap works with a finger at 390x844. Tap a Basic
// then a Bench slot and it is placed; tap the Energy Zone then a Pokemon and it
// attaches; holding still zooms and letting go keeps nothing held. Every touch
// here is a real touch event sent through CDP, not a mouse click.

type Card = { id: string; name: string; stage: string };
type View = { phase: string; active: number; pending: { player: number } | null; players: { hand: Card[] }[] };

// Pinned so draft-pool changes can't shift the bot's deck.
const BOT = ['A1-211', 'A1-202', 'B4-015', 'B4-112', 'A3-044', 'A2-112', 'A1-150', 'B4-036', 'P-A-001', 'A4b-373'].flatMap((id) => [id, id]);

async function post(api: APIRequestContext, url: string, data?: object): Promise<View> {
  const res = await api.post(url, data ? { data } : undefined);
  const body = await res.json();
  return body.state ?? body;
}

// Seed 7: you go first; on your second turn a Basic is in hand, the Bench is empty, energy is ready.
async function toYourSecondTurn(api: APIRequestContext): Promise<Card> {
  const start = await post(api, '/api/new', { seed: 7, bot: BOT, preset: 'fire' });
  const basics = start.players[0].hand.filter((c) => c.stage === 'Basic');
  await post(api, '/api/move', { type: 'SetupActive', player: 0, cardId: basics[0].id });
  await post(api, '/api/move', { type: 'SetupReady', player: 0 });
  for (let i = 0; i < 6 && (await post(api, '/api/bot')).phase === 'setup'; i++);
  await post(api, '/api/move', { type: 'EndTurn', player: 0 });
  for (let i = 0; i < 40; i++) {
    const v = await post(api, '/api/bot');
    if (v.phase !== 'main' || (v.pending ? v.pending.player : v.active) === 0) break;
  }
  return basics[1];
}

type Point = { x: number; y: number };

async function centre(l: Locator): Promise<Point> {
  const b = (await l.boundingBox())!;
  return { x: b.x + b.width / 2, y: b.y + b.height / 2 };
}

// Hand cards overlap to the right, so a finger lands on the visible left edge.
async function leftEdge(l: Locator): Promise<Point> {
  const b = (await l.boundingBox())!;
  return { x: b.x + Math.min(10, b.width / 2), y: b.y + b.height / 2 };
}

const touch = (cdp: CDPSession, type: string, p?: Point) =>
  cdp.send('Input.dispatchTouchEvent', { type, touchPoints: p ? [p] : [] });

async function fingerTap(cdp: CDPSession, p: Point) {
  await touch(cdp, 'touchStart', p);
  await touch(cdp, 'touchEnd');
}

const handCard = (page: Page, id: string) => page.getByTestId('my-hand').locator(`[data-card-id="${id}"]`).first();
const slot = (page: Page, name: string) => page.getByTestId('my-bench').locator(`[data-slot="${name}"]`);

test('with a finger: tap a Basic then a Bench slot, tap the Energy Zone then a Pokemon, hold to zoom', async ({ page, request }) => {
  const basic = await toYourSecondTurn(request);
  await page.goto('/?battle');
  await expect(page.getByTestId('turn-banner')).toBeHidden({ timeout: 3000 });
  const cdp = await page.context().newCDPSession(page);

  // Tap the Basic: it is picked up, the Bench glows. Tap the slot: placed.
  await fingerTap(cdp, await leftEdge(handCard(page, basic.id)));
  await expect(page.getByTestId('play-hint')).toContainText('Bench');
  await expect(slot(page, 'right')).toHaveAttribute('data-glow', 'true');
  await fingerTap(cdp, await centre(slot(page, 'right')));
  await expect(slot(page, 'right').locator('img')).toHaveAttribute('alt', basic.name);
  await expect(page.getByTestId('play-hint')).toHaveCount(0);

  // Tap the Energy Zone, then the benched Pokemon: the energy attaches.
  await fingerTap(cdp, await centre(page.getByTestId('my-energy').getByTestId('energy-current')));
  await expect(page.getByTestId('play-hint')).toHaveText('Choose a Pokémon to attach energy to');
  await fingerTap(cdp, await centre(slot(page, 'right')));
  await expect(slot(page, 'right').getByTestId('energy-icon')).toHaveCount(1);

  // Holding still on your Active zooms it; lifting the finger closes the zoom and holds nothing.
  await touch(cdp, 'touchStart', await centre(page.getByTestId('my-active')));
  await expect(page.getByTestId('card-zoom')).toBeVisible();
  await touch(cdp, 'touchEnd');
  await expect(page.getByTestId('card-zoom')).toHaveCount(0);
  await expect(page.getByTestId('play-hint')).toHaveCount(0);
});
