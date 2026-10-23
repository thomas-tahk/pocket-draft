import { test, expect, type APIRequestContext, type Locator, type Page } from '@playwright/test';
import { FIRE } from './decks';

// F1, F3–F5, F17 at phone size: press-and-drag a card onto the board. Legal
// spots glow, the rest dims, a line at the bottom says what to do; dropping
// elsewhere cancels; desktop tap-then-tap works; a card with nowhere to go
// flashes the engine's reason. The game is driven through the API to your
// second turn (energy available, a Basic still in hand, Bench empty).

type Card = { id: string; name: string; stage: string };
type View = { phase: string; active: number; pending: { player: number } | null; players: { hand: Card[] }[] };

// Pinned so draft-pool changes can't shift the bot's deck.
const BOT = ['A1-211', 'A1-202', 'B4-015', 'B4-112', 'A3-044', 'A2-112', 'A1-150', 'B4-036', 'P-A-001', 'A4b-373'].flatMap((id) => [id, id]);

async function post(api: APIRequestContext, url: string, data?: object): Promise<View> {
  const res = await api.post(url, data ? { data } : undefined);
  const body = await res.json();
  return body.state ?? body;
}

async function runBotUntilYourTurn(api: APIRequestContext) {
  for (let i = 0; i < 40; i++) {
    const v = await post(api, '/api/bot');
    const actor = v.pending ? v.pending.player : v.active;
    if (v.phase !== 'main' || actor === 0) return;
  }
}

// Seed 7: you go first. Active only, so the Bench is empty and a Basic stays in hand.
async function toYourSecondTurn(api: APIRequestContext): Promise<Card> {
  const start = await post(api, '/api/new', { seed: 7, bot: BOT, you: FIRE });
  const basics = start.players[0].hand.filter((c) => c.stage === 'Basic');
  await post(api, '/api/move', { type: 'SetupActive', player: 0, cardId: basics[0].id });
  await post(api, '/api/move', { type: 'SetupReady', player: 0 });
  for (let i = 0; i < 6 && (await post(api, '/api/bot')).phase === 'setup'; i++);
  await post(api, '/api/move', { type: 'EndTurn', player: 0 });
  await runBotUntilYourTurn(api);
  return basics[1];
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

async function drag(page: Page, from: { x: number; y: number }, to: Locator, midway?: () => Promise<void>) {
  const b = await centre(to);
  await page.mouse.move(from.x, from.y);
  await page.mouse.down();
  await page.mouse.move((from.x + b.x) / 2, (from.y + b.y) / 2, { steps: 6 });
  if (midway) await midway();
  await page.mouse.move(b.x, b.y, { steps: 6 });
  await page.mouse.up();
}

const handCard = (page: Page, id: string) => page.getByTestId('my-hand').locator(`[data-card-id="${id}"]`).first();
const slot = (page: Page, name: string) => page.getByTestId('my-bench').locator(`[data-slot="${name}"]`);

test('drag a Basic to a Bench slot and energy to the Active', async ({ page, request }) => {
  const basic = await toYourSecondTurn(request);
  await page.goto('/?battle');
  await expect(page.getByTestId('battle-board')).toBeVisible();
  const active = page.getByTestId('my-active');
  const activeSpot = active.locator('[data-spot="active"]');
  await expect(page.getByTestId('turn-banner')).toBeHidden({ timeout: 3000 }); // a clear screenshot

  // Basic: only the empty Bench slots glow; the board dims; the hint says what to do.
  await drag(page, await grabPoint(handCard(page, basic.id)), slot(page, 'left'), async () => {
    for (const s of ['left', 'behind', 'right']) await expect(slot(page, s)).toHaveAttribute('data-glow', 'true');
    await expect(activeSpot).not.toHaveAttribute('data-glow', 'true');
    await expect(page.getByTestId('play-dim')).toBeVisible();
    await expect(page.getByTestId('play-hint')).toContainText('Bench');
    await expect(page.getByTestId('play-ghost')).toBeVisible();
    await page.waitForFunction(() => [...document.images].every((i) => i.complete));
    await page.screenshot({ path: 'test-results/battle-drag-in-progress.png' });
  });
  await expect(slot(page, 'left').locator('img')).toHaveAttribute('alt', basic.name);
  await expect(page.getByTestId('play-hint')).toHaveCount(0);
  await expect(page.getByTestId('play-dim')).toHaveCount(0);

  // Energy: drag this turn's energy from the zone onto the Active.
  const energy = page.getByTestId('my-energy').getByTestId('energy-current');
  await drag(page, await centre(energy), active, async () => {
    await expect(page.getByTestId('play-hint')).toHaveText('Choose a Pokémon to attach energy to');
    await expect(activeSpot).toHaveAttribute('data-glow', 'true');
    await expect(slot(page, 'left')).toHaveAttribute('data-glow', 'true');
    await expect(slot(page, 'right')).not.toHaveAttribute('data-glow', 'true');
  });
  await expect(active.getByTestId('energy-icon')).toHaveCount(1);
  await expect(energy).toHaveAttribute('data-spent', 'true');

  // F17: the energy is spent, so picking it up again flashes the engine's reason.
  const e = await centre(energy);
  await page.mouse.move(e.x, e.y);
  await page.mouse.down();
  await expect(page.getByTestId('play-reason')).toHaveText(/energy was already attached this turn/i);
  await expect(page.getByTestId('play-hint')).toHaveCount(0);
  await page.mouse.up();
  await expect(page.getByTestId('play-reason')).toHaveCount(0, { timeout: 4000 });
});

test('drop elsewhere cancels; tap-then-tap plays; touch drags energy', async ({ page, request }) => {
  const basic = await toYourSecondTurn(request);
  await page.goto('/?battle');
  await expect(page.getByTestId('battle-board')).toBeVisible();
  const card = handCard(page, basic.id);

  // Dropping on a spot that doesn't glow cancels: the card stays in hand.
  await drag(page, await grabPoint(card), page.getByTestId('opp-active'));
  await expect(card).toBeVisible();
  await expect(page.getByTestId('play-hint')).toHaveCount(0);
  await expect(slot(page, 'right').locator('img')).toHaveCount(0);

  // Tap the card, tap somewhere that isn't a spot: cancelled.
  const p = await grabPoint(card);
  await page.mouse.click(p.x, p.y);
  await expect(page.getByTestId('play-hint')).toContainText('Bench');
  await expect(slot(page, 'right')).toHaveAttribute('data-glow', 'true');
  await page.mouse.click(195, 20);
  await expect(page.getByTestId('play-hint')).toHaveCount(0);

  // Tap the card, then tap a glowing slot: played there.
  await page.mouse.click(p.x, p.y);
  const r = await centre(slot(page, 'right'));
  await page.mouse.click(r.x, r.y);
  await expect(slot(page, 'right').locator('img')).toHaveAttribute('alt', basic.name);

  // Touch uses the same pointer path: drag the energy onto the right Bench slot by finger.
  const cdp = await page.context().newCDPSession(page);
  const from = await centre(page.getByTestId('my-energy').getByTestId('energy-current'));
  const to = await centre(slot(page, 'right'));
  const touch = (type: string, x: number, y: number) =>
    cdp.send('Input.dispatchTouchEvent', { type, touchPoints: type === 'touchEnd' ? [] : [{ x, y }] });
  await touch('touchStart', from.x, from.y);
  for (let i = 1; i <= 8; i++) await touch('touchMove', from.x + ((to.x - from.x) * i) / 8, from.y + ((to.y - from.y) * i) / 8);
  await expect(page.getByTestId('play-hint')).toHaveText('Choose a Pokémon to attach energy to');
  await touch('touchEnd', to.x, to.y);
  await expect(slot(page, 'right').getByTestId('energy-icon')).toHaveCount(1);
});
